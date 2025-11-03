#!/usr/bin/env bash

# This script runs a suite of tests to validate database migrations.
# The migrations files are located in the backend/migrations directory.
# Every table, view, function, and index created in the DB must be tested by tests in this script.
# - Each test is self-contained and cleans up after itself

set -Eeuo pipefail

########################################
# Environment and psql setup
########################################
ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
ENV_FILE="${ROOT_DIR}/docker/.env"

# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"

########################################
# Pretty output helpers
########################################
RED="\033[31m"
GREEN="\033[32m"
YELLOW="\033[33m"
BLUE="\033[34m"
BOLD="\033[1m"
RESET="\033[0m"

hr() { printf '%*s\n' "${1:-80}" '' | tr ' ' -; }
header() {
    echo
    hr 80
    printf "%b%s%b\n" "$BOLD$BLUE" "$1" "$RESET"
    hr 80
}
info() { printf "%bℹ %s%b\n" "$YELLOW" "$1" "$RESET"; }
ok() { printf "%b✓ %s%b\n" "$GREEN" "$1" "$RESET"; }
err() { printf "%b✗ %s%b\n" "$RED" "$1" "$RESET"; }

fail_count=0
pass() { ok "$1"; }
fail() {
    err "$1"
    fail_count=$((fail_count + 1))
}

require_cmd() {
    command -v "$1" >/dev/null 2>&1 || {
        echo "Required command not found: $1" >&2
        exit 127
    }
}

load_env() {
    if [[ -f "$ENV_FILE" ]]; then
        # shellcheck source=../.env
        source "$ENV_FILE"
    else
        info "Env file not found at $ENV_FILE. Falling back to defaults."
    fi

    # Unified psql command (unaligned, tuples only, custom field separator, stop on error)

}

run_sql() {
    local sql_statement="${1}"
    # Print compact SQL for easier debugging (stderr)
    printf 'Executing SQL: %s\n' "${sql_statement//[$'\n\r']/ }" >&2
    # Use quiet, tuples-only, unaligned output for reliable command-sub substitution
    psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${POSTGRES_DB}" \
        -X -v ON_ERROR_STOP=1 \
        -qAt -F , \
        --pset footer=off \
        --pset pager=off \
        -c "${sql_statement}"
}

########################################
# Test utilities
########################################
random_suffix() { printf '%s' "${RANDOM}${RANDOM}${RANDOM}"; }

create_archer() {
    local suffix
    suffix="$(random_suffix)"
    local email="migtest_${suffix}@example.com"
    local gsubj="gs-${suffix}"
    local sql_statement="""
        INSERT INTO archer (
            email,
            first_name,
            last_name,
            date_of_birth,
            gender,
            bowstyle,
            draw_weight,
            google_subject
        )
        VALUES (
            '$email',
            'Test',
            '$suffix',
            '1990-01-01'::date,
            'male'::gender_type,
            'recurve'::bowstyle_type,
            40.3,
            '$gsubj'
        )
        RETURNING archer_id;
"""
    # Explicitly cast enum/date literals; include club_id as NULL to be explicit
    run_sql "${sql_statement}"
}

create_session() {
    # $1 archer_id, $2 is_opened (true/false), [$3 shot_per_round]
    local archer_id="$1"
    local opened="$2"
    local spr="${3:-}"
    local sql_statement

    if [[ -n "$spr" ]]; then
        sql_statement="""
            INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened, shot_per_round)
            VALUES ('$archer_id', 'Test Range', true, ${opened}, ${spr})
            RETURNING session_id;
"""
    else
        sql_statement="""
            INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened)
            VALUES ('$archer_id', 'Test Range', true, ${opened})
            RETURNING session_id;
"""
    fi
    run_sql "${sql_statement}"
}

create_target() {
    # $1 session_id, $2 distance, $3 lane
    local sid="$1"
    local dist="$2"
    local lane="$3"
    local sql_statement="""
        INSERT INTO target (session_id, distance, lane)
        VALUES ('$sid', $dist, $lane)
        RETURNING target_id;
"""
    run_sql "${sql_statement}"
}

assign_slot() {
    # $1 target_id, $2 archer_id, $3 session_id, $4 face_type, $5 slot_letter, $6 is_shooting (true/false)
    local sql_statement="""
        INSERT INTO slot (
            target_id,
            archer_id,
            session_id,
            face_type,
            slot_letter,
            is_shooting,
            bowstyle,
            draw_weight,
            club_id
        )
        VALUES ('$1', '$2', '$3', '$4'::face_type, '$5', $6, 'recurve'::bowstyle_type, 40.0, NULL)
        RETURNING slot_id;
"""
    run_sql "${sql_statement}"
}

cleanup_ids() {
    # Accepts list of IDs to delete from known tables best-effort
    # We delete in FK order: slot -> target -> session -> archer
    local ids=("$@")
    for id in "${ids[@]}"; do
        [[ -n "$id" ]] || continue
        run_sql "DELETE FROM slot WHERE slot_id = '$id' OR target_id = '$id' OR archer_id = '$id' OR session_id = '$id';" || true
        run_sql "DELETE FROM target WHERE target_id = '$id' OR session_id = '$id';" || true
        run_sql "DELETE FROM session WHERE session_id = '$id' OR owner_archer_id = '$id';" || true
        run_sql "DELETE FROM archer WHERE archer_id = '$id';" || true
    done
}

########################################
# Tests
########################################

test_unique_open_session_per_archer() {
    header "unique_open_session_per_archer index"

    local archer_id s1 s2
    archer_id="$(create_archer)"
    archer_id="${archer_id//$'\n'/}"

    s1="$(create_session "$archer_id" true)"
    s1="${s1//$'\n'/}"
    pass "Created first open session: $s1"

    # Expect second open session for same archer to FAIL (unique partial index)
    if run_sql "INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened)
                            VALUES ('$archer_id','Another Range',true,true);" >/dev/null 2>&1; then
        fail "Second open session unexpectedly succeeded (index not enforced?)"
    else
        pass "Second open session correctly rejected"
    fi

    # Closed session should be allowed
    s2="$(create_session "$archer_id" false)"
    s2="${s2//$'\n'/}"
    pass "Created closed session: $s2"

    # Cleanup
    cleanup_ids "$s1" "$s2" "$archer_id"
}

test_get_next_lane_function() {
    header "get_next_lane function"

    local archer_id sid t1 t2 next
    archer_id="$(create_archer)"
    archer_id="${archer_id//$'\n'/}"
    sid="$(create_session "$archer_id" true)"
    sid="${sid//$'\n'/}"

    t1="$(create_target "$sid" 70 1)"
    t1="${t1//$'\n'/}"
    t2="$(create_target "$sid" 70 2)"
    t2="${t2//$'\n'/}"
    pass "Inserted target at lanes 1 and 2"

    next="$(run_sql "SELECT get_next_lane('$sid');")"
    next="${next//$'\n'/}"
    if [[ "$next" == "3" ]]; then
        pass "get_next_lane returned 3 as expected"
    else
        fail "get_next_lane returned '$next' (expected 3)"
    fi

    cleanup_ids "$t1" "$t2" "$sid" "$archer_id"
}

test_slot_shooting_participants_view() {
    header "active_slots view"

    local a1 a2 s_open s_closed t_open cnt

    a1="$(create_archer)"
    a1="${a1//$'\n'/}"
    a2="$(create_archer)"
    a2="${a2//$'\n'/}"

    s_open="$(create_session "$a1" true)"
    s_open="${s_open//$'\n'/}"
    s_closed="$(create_session "$a1" false)"
    s_closed="${s_closed//$'\n'/}"

    t_open="$(create_target "$s_open" 70 1)"
    t_open="${t_open//$'\n'/}"

    assign_slot "$t_open" "$a1" "$s_open" '40cm_full' 'A' true
    assign_slot "$t_open" "$a2" "$s_open" '40cm_full' 'B' false

    # Create slot in closed session (should not appear in view)
    local t_closed
    t_closed="$(create_target "$s_closed" 70 2)"
    t_closed="${t_closed//$'\n'/}"
    assign_slot "$t_closed" "$a2" "$s_closed" '40cm_full' 'A' true

    cnt="$(run_sql "SELECT count(*) FROM active_slots WHERE session_id = '$s_open';")"
    cnt="${cnt//$'\n'/}"
    if [[ "$cnt" == "1" ]]; then
        pass "active_slots returned 1 active shooter for open session"
    else
        fail "active_slots returned '$cnt' (expected 1)"
    fi

    cleanup_ids "$t_open" "$t_closed" "$s_open" "$s_closed" "$a1" "$a2"
}

test_get_available_targets_function() {
    header "get_available_targets function"

    local s a1 a2 a3 a4 a5 a6 a7 a8 t1 t2 t3
    a1="$(create_archer)"
    a1="${a1//$'\n'/}"
    a2="$(create_archer)"
    a2="${a2//$'\n'/}"
    a3="$(create_archer)"
    a3="${a3//$'\n'/}"
    a4="$(create_archer)"
    a4="${a4//$'\n'/}"
    a5="$(create_archer)"
    a5="${a5//$'\n'/}"
    a6="$(create_archer)"
    a6="${a6//$'\n'/}"
    a7="$(create_archer)"
    a7="${a7//$'\n'/}"
    a8="$(create_archer)"
    a8="${a8//$'\n'/}"

    s="$(create_session "$a1" true)"
    s="${s//$'\n'/}"

    t1="$(create_target "$s" 70 1)"
    t1="${t1//$'\n'/}"
    t2="$(create_target "$s" 70 2)"
    t2="${t2//$'\n'/}"
    t3="$(create_target "$s" 50 3)"
    t3="${t3//$'\n'/}"

    # Occupy t1 with 3 archer
    assign_slot "$t1" "$a1" "$s" '40cm_full' 'A' true
    assign_slot "$t1" "$a2" "$s" '40cm_full' 'B' true
    assign_slot "$t1" "$a3" "$s" '40cm_full' 'C' true

    # Occupy t2 with 4 distinct archer (should be excluded)
    assign_slot "$t2" "$a5" "$s" '40cm_full' 'A' true
    assign_slot "$t2" "$a6" "$s" '40cm_full' 'B' true
    assign_slot "$t2" "$a7" "$s" '40cm_full' 'C' true
    assign_slot "$t2" "$a8" "$s" '40cm_full' 'D' true

    # Query available target at distance 70
    local rows
    rows="$(run_sql "SELECT lane, occupied FROM get_available_targets('$s', 70) ORDER BY lane;")"
    # Expect a single line: "1,3"
    if [[ "$rows" == "1,3" ]]; then
        pass "get_available_targets returned lane 1 with occupied=3"
    else
        fail "get_available_targets returned '$rows' (expected '1,3')"
    fi

    cleanup_ids "$t1" "$t2" "$t3" "$s" "$a1" "$a2" "$a3" "$a4" "$a5" "$a6" "$a7" "$a8"
}

test_get_available_targets_occupancy_steps_18() {
    header "get_available_targets occupancy steps (distance 18)"

    local owner a1 a2 a3 a4 sid tid rows

    # Create 5 archers: one owner and four occupants
    owner="$(create_archer)"
    owner="${owner//$'\n'/}"
    a1="$(create_archer)"
    a1="${a1//$'\n'/}"
    a2="$(create_archer)"
    a2="${a2//$'\n'/}"
    a3="$(create_archer)"
    a3="${a3//$'\n'/}"
    a4="$(create_archer)"
    a4="${a4//$'\n'/}"

    sid="$(create_session "$owner" true)"
    sid="${sid//$'\n'/}"
    tid="$(create_target "$sid" 18 1)"
    tid="${tid//$'\n'/}"

    # 1st slot -> occupied should be 1
    assign_slot "$tid" "$a1" "$sid" '40cm_full' 'A' true >/dev/null
    rows="$(run_sql "SELECT lane, occupied FROM get_available_targets('$sid', 18);")"
    if [[ "$rows" == "1,1" ]]; then
        pass "get_available_targets shows occupied=1 after one slot"
    else
        fail "After 1 slot, got '$rows' (expected '1,1')"
    fi

    # 2nd slot -> occupied should be 2
    assign_slot "$tid" "$a2" "$sid" '40cm_full' 'B' true >/dev/null
    rows="$(run_sql "SELECT lane, occupied FROM get_available_targets('$sid', 18);")"
    if [[ "$rows" == "1,2" ]]; then
        pass "get_available_targets shows occupied=2 after two slots"
    else
        fail "After 2 slots, got '$rows' (expected '1,2')"
    fi

    # 3rd slot -> occupied should be 3
    assign_slot "$tid" "$a3" "$sid" '40cm_full' 'C' true >/dev/null
    rows="$(run_sql "SELECT lane, occupied FROM get_available_targets('$sid', 18);")"
    if [[ "$rows" == "1,3" ]]; then
        pass "get_available_targets shows occupied=3 after three slots"
    else
        fail "After 3 slots, got '$rows' (expected '1,3')"
    fi

    # 4th slot -> target becomes full and should be excluded (no rows)
    assign_slot "$tid" "$a4" "$sid" '40cm_full' 'D' true >/dev/null
    rows="$(run_sql "SELECT lane, occupied FROM get_available_targets('$sid', 18);")"
    if [[ -z "$rows" ]]; then
        pass "get_available_targets excludes full target (occupied=4) as expected"
    else
        fail "After 4 slots, expected no rows; got '$rows'"
    fi

    cleanup_ids "$tid" "$sid" "$owner" "$a1" "$a2" "$a3" "$a4"
}

test_get_slot_with_lane_function() {
    header "get_slot_with_lane function"

    local archer_id sid tid slot_id got_slot nonexist_slot_id

    archer_id="$(create_archer)"
    archer_id="${archer_id//$'\n'/}"
    sid="$(create_session "$archer_id" true)"
    sid="${sid//$'\n'/}"
    # Use lane 4 to make slot text deterministic (e.g., 4A)
    tid="$(create_target "$sid" 70 4)"
    tid="${tid//$'\n'/}"
    slot_id="$(assign_slot "$tid" "$archer_id" "$sid" '40cm_full' 'A' true)"
    slot_id="${slot_id//$'\n'/}"

    got_slot="$(run_sql "SELECT slot FROM get_slot_with_lane('$slot_id');")"
    got_slot="${got_slot//$'\n'/}"
    if [[ "$got_slot" == "4A" ]]; then
        pass "get_slot_with_lane returned expected slot label 4A"
    else
        fail "get_slot_with_lane returned '$got_slot' (expected '4A')"
    fi

    # Non-existent slot_id should return 0 rows
    nonexist_slot_id="$(run_sql "SELECT uuid_generate_v4();")"
    nonexist_slot_id="${nonexist_slot_id//$'\n'/}"
    got_slot="$(run_sql "SELECT slot FROM get_slot_with_lane('$nonexist_slot_id');")"
    if [[ -z "$got_slot" ]]; then
        pass "get_slot_with_lane returned no rows for unknown slot_id"
    else
        fail "get_slot_with_lane returned rows for unknown slot_id: '$got_slot'"
    fi

    cleanup_ids "$slot_id" "$tid" "$sid" "$archer_id"
}

test_get_active_slot_id_function() {
    header "get_active_slot_id function"

    local archer_id sid tid slot_id res rnd_archer

    archer_id="$(create_archer)"
    archer_id="${archer_id//$'\n'/}"
    sid="$(create_session "$archer_id" true)"
    sid="${sid//$'\n'/}"
    tid="$(create_target "$sid" 70 2)"
    tid="${tid//$'\n'/}"
    slot_id="$(assign_slot "$tid" "$archer_id" "$sid" '40cm_full' 'A' true)"
    slot_id="${slot_id//$'\n'/}"

    # Happy path: active shooter in open session
    res="$(run_sql "SELECT get_active_slot_id('$archer_id');")"
    res="${res//$'\n'/}"
    if [[ "$res" == "$slot_id" ]]; then
        pass "get_active_slot_id returned expected slot_id for active shooter"
    else
        fail "get_active_slot_id returned '$res' (expected '$slot_id')"
    fi

    # Inactive shooter should yield NULL
    run_sql "UPDATE slot SET is_shooting = FALSE WHERE slot_id = '$slot_id';" >/dev/null
    res="$(run_sql "SELECT get_active_slot_id('$archer_id');")"
    res="${res//$'\n'/}"
    if [[ -z "$res" ]]; then
        pass "get_active_slot_id returned NULL for non-active shooter"
    else
        fail "get_active_slot_id returned '$res' for non-active shooter (expected NULL)"
    fi

    # Closed session should yield NULL even if shooter is set active
    run_sql "UPDATE slot SET is_shooting = TRUE WHERE slot_id = '$slot_id'; UPDATE session SET is_opened = FALSE WHERE session_id = '$sid';" >/dev/null
    res="$(run_sql "SELECT get_active_slot_id('$archer_id');")"
    res="${res//$'\n'/}"
    if [[ -z "$res" ]]; then
        pass "get_active_slot_id returned NULL for closed session"
    else
        fail "get_active_slot_id returned '$res' for closed session (expected NULL)"
    fi

    # Non-existent archer should yield NULL
    rnd_archer="$(run_sql "SELECT uuid_generate_v4();")"
    rnd_archer="${rnd_archer//$'\n'/}"
    res="$(run_sql "SELECT get_active_slot_id('$rnd_archer');")"
    res="${res//$'\n'/}"
    if [[ -z "$res" ]]; then
        pass "get_active_slot_id returned NULL for unknown archer_id"
    else
        fail "get_active_slot_id returned '$res' for unknown archer_id (expected NULL)"
    fi

    cleanup_ids "$slot_id" "$tid" "$sid" "$archer_id"
}

test_get_next_lane_with_full_targets() {
    header "get_next_lane with two full targets"

    local a1 a2 a3 a4 a5 a6 a7 a8 sid t1 t2 next

    # Create 8 archers
    a1="$(create_archer)"
    a1="${a1//$'\n'/}"
    a2="$(create_archer)"
    a2="${a2//$'\n'/}"
    a3="$(create_archer)"
    a3="${a3//$'\n'/}"
    a4="$(create_archer)"
    a4="${a4//$'\n'/}"
    a5="$(create_archer)"
    a5="${a5//$'\n'/}"
    a6="$(create_archer)"
    a6="${a6//$'\n'/}"
    a7="$(create_archer)"
    a7="${a7//$'\n'/}"
    a8="$(create_archer)"
    a8="${a8//$'\n'/}"

    # Open session owned by a1
    sid="$(create_session "$a1" true)"
    sid="${sid//$'\n'/}"

    # Two targets at lanes 1 and 2
    t1="$(create_target "$sid" 70 1)"
    t1="${t1//$'\n'/}"
    t2="$(create_target "$sid" 70 2)"
    t2="${t2//$'\n'/}"

    # Fill target 1 (A-D)
    assign_slot "$t1" "$a1" "$sid" '40cm_full' 'A' true >/dev/null
    assign_slot "$t1" "$a2" "$sid" '40cm_full' 'B' true >/dev/null
    assign_slot "$t1" "$a3" "$sid" '40cm_full' 'C' true >/dev/null
    assign_slot "$t1" "$a4" "$sid" '40cm_full' 'D' true >/dev/null

    # Fill target 2 (A-D)
    assign_slot "$t2" "$a5" "$sid" '40cm_full' 'A' true >/dev/null
    assign_slot "$t2" "$a6" "$sid" '40cm_full' 'B' true >/dev/null
    assign_slot "$t2" "$a7" "$sid" '40cm_full' 'C' true >/dev/null
    assign_slot "$t2" "$a8" "$sid" '40cm_full' 'D' true >/dev/null

    # get_next_lane should return 3 regardless of fullness, since max lane is 2
    next="$(run_sql "SELECT get_next_lane('$sid');")"
    next="${next//$'\n'/}"
    if [[ "$next" == "3" ]]; then
        pass "get_next_lane returned 3 when two existing targets were full"
    else
        fail "get_next_lane returned '$next' with full targets (expected 3)"
    fi

    cleanup_ids "$t1" "$t2" "$sid" "$a1" "$a2" "$a3" "$a4" "$a5" "$a6" "$a7" "$a8"
}

test_session_shot_per_round_defaults_and_constraints() {
    header "session.shot_per_round defaults and constraints"

    local owner_default owner_custom sid_default sid_custom got_default got_custom

    # Default path: omit shot_per_round and expect DEFAULT 6
    owner_default="$(create_archer)"
    owner_default="${owner_default//$'\n'/}"
    sid_default="$(create_session "$owner_default" true)"
    sid_default="${sid_default//$'\n'/}"
    got_default="$(run_sql "SELECT shot_per_round FROM session WHERE session_id = '$sid_default';")"
    got_default="${got_default//$'\n'/}"
    if [[ "$got_default" == "6" ]]; then
        pass "session.shot_per_round defaults to 6 when omitted"
    else
        fail "session.shot_per_round default was '$got_default' (expected 6)"
    fi

    # Custom value path: provide explicit positive value
    owner_custom="$(create_archer)"
    owner_custom="${owner_custom//$'\n'/}"
    sid_custom="$(create_session "$owner_custom" true 3)"
    sid_custom="${sid_custom//$'\n'/}"
    got_custom="$(run_sql "SELECT shot_per_round FROM session WHERE session_id = '$sid_custom';")"
    got_custom="${got_custom//$'\n'/}"
    if [[ "$got_custom" == "3" ]]; then
        pass "session.shot_per_round accepts custom positive value (3)"
    else
        fail "session.shot_per_round read '$got_custom' (expected 3)"
    fi

    # Constraint check: zero should be rejected
    if run_sql "INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened, shot_per_round)
                VALUES ('$owner_custom', 'Zero Check', true, false, 0);" >/dev/null 2>&1; then
        fail "session.shot_per_round allowed 0 (expected CHECK constraint to reject)"
    else
        pass "session.shot_per_round rejects 0 per CHECK constraint"
    fi

    # Constraint check: negative should be rejected
    if run_sql "INSERT INTO session (owner_archer_id, session_location, is_indoor, is_opened, shot_per_round)
                VALUES ('$owner_custom', 'Negative Check', true, false, -1);" >/dev/null 2>&1; then
        fail "session.shot_per_round allowed negative value (expected rejection)"
    else
        pass "session.shot_per_round rejects negative values per CHECK constraint"
    fi

    cleanup_ids "$sid_default" "$sid_custom" "$owner_default" "$owner_custom"
}

test_notify_shot_insert_trigger() {
    header "notify_shot_insert() + trigger on shot inserts"

    local owner archer_a archer_b sid tid slot_a slot_b chan_a chan_b
    local listener_a_log listener_b_log listener_a_pid listener_b_pid

    # Test setup: 1 session, 1 target, 2 archers, 2 slots
    owner="$(create_archer)"
    owner="${owner//$'\n'/}"
    archer_a="$(create_archer)"
    archer_a="${archer_a//$'\n'/}"
    archer_b="$(create_archer)"
    archer_b="${archer_b//$'\n'/}"

    sid="$(create_session "$owner" true)"
    sid="${sid//$'\n'/}"
    tid="$(create_target "$sid" 18 1)"
    tid="${tid//$'\n'/}"

    slot_a="$(assign_slot "$tid" "$archer_a" "$sid" '40cm_full' 'A' true)"
    slot_a="${slot_a//$'\n'/}"
    slot_b="$(assign_slot "$tid" "$archer_b" "$sid" '40cm_full' 'B' true)"
    slot_b="${slot_b//$'\n'/}"

    chan_a="shot_insert_${slot_a}"
    chan_b="shot_insert_${slot_b}"

    # Start two listeners (one per slot channel) and capture their outputs
    listener_a_log="$(mktemp)"
    listener_b_log="$(mktemp)"

    psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${POSTGRES_DB}" \
        -X -v ON_ERROR_STOP=1 \
        >"${listener_a_log}" 2>&1 \
        <<SQL &
LISTEN "${chan_a}";
\echo listener_a_ready
-- Keep the connection alive and ensure notifications are flushed between SQL commands
SELECT pg_sleep(2);
SELECT pg_sleep(2);
SQL
    listener_a_pid=$!

    psql \
        -h "${POSTGRES_HOST}" \
        -p "${POSTGRES_PORT}" \
        -U "${POSTGRES_USER}" \
        -d "${POSTGRES_DB}" \
        -X -v ON_ERROR_STOP=1 \
        >"${listener_b_log}" 2>&1 \
        <<SQL &
LISTEN "${chan_b}";
\echo listener_b_ready
-- Keep the connection alive and ensure notifications are flushed between SQL commands
SELECT pg_sleep(2);
SELECT pg_sleep(2);
SQL
    listener_b_pid=$!

    # Give listeners a brief moment to register
    sleep 0.5

    # Insert 8 shots alternating across slots: A, B, A, B, A, B, A, B
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_a}', 1.1, 2.2, 10, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_b}', 1.2, 2.3, 9, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_a}', 1.3, 2.4, 8, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_b}', 1.4, 2.5, 7, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_a}', 1.5, 2.6, 6, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_b}', 1.6, 2.7, 5, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_a}', 1.7, 2.8, 4, NULL);" >/dev/null
    run_sql "INSERT INTO shot (slot_id, x, y, score, arrow_id) VALUES ('${slot_b}', 1.8, 2.9, 3, NULL);" >/dev/null

    # Wait for listeners to process notifications and exit
    wait "$listener_a_pid" || true
    wait "$listener_b_pid" || true

    # Validate listener outputs: each should receive exactly 4 notifications for its channel
    local count_a count_b wrong_a wrong_b
    count_a=$(grep -c "Asynchronous notification \"${chan_a}\"" "${listener_a_log}" || true)
    count_b=$(grep -c "Asynchronous notification \"${chan_b}\"" "${listener_b_log}" || true)
    wrong_a=$(grep -c "Asynchronous notification \"${chan_b}\"" "${listener_a_log}" || true)
    wrong_b=$(grep -c "Asynchronous notification \"${chan_a}\"" "${listener_b_log}" || true)

    if [[ "${count_a}" == "4" ]]; then
        pass "listener_a received 4 notifications for its channel"
    else
        fail "listener_a received ${count_a} notifications (expected 4)"
    fi

    if [[ "${count_b}" == "4" ]]; then
        pass "listener_b received 4 notifications for its channel"
    else
        fail "listener_b received ${count_b} notifications (expected 4)"
    fi

    if [[ "${wrong_a}" == "0" ]]; then
        pass "listener_a did not receive notifications for listener_b channel"
    else
        fail "listener_a received ${wrong_a} notifications for listener_b channel (expected 0)"
    fi

    if [[ "${wrong_b}" == "0" ]]; then
        pass "listener_b did not receive notifications for listener_a channel"
    else
        fail "listener_b received ${wrong_b} notifications for listener_a channel (expected 0)"
    fi

    # Enhanced validation: assert JSON payload contains correct slot_id and scores per listener
    # Verify slot_id appears in payloads exactly 4 times for the correct listener
    local payload_slot_a payload_slot_b foreign_slot_in_a foreign_slot_in_b
    payload_slot_a=$(grep -c "\"slot_id\":\"${slot_a}\"" "${listener_a_log}" || true)
    payload_slot_b=$(grep -c "\"slot_id\":\"${slot_b}\"" "${listener_b_log}" || true)
    foreign_slot_in_a=$(grep -c "\"slot_id\":\"${slot_b}\"" "${listener_a_log}" || true)
    foreign_slot_in_b=$(grep -c "\"slot_id\":\"${slot_a}\"" "${listener_b_log}" || true)

    if [[ "${payload_slot_a}" == "4" ]]; then
        pass "listener_a payloads include correct slot_id ${slot_a} exactly 4 times"
    else
        fail "listener_a payloads include slot_id ${slot_a} ${payload_slot_a} time(s) (expected 4)"
    fi

    if [[ "${payload_slot_b}" == "4" ]]; then
        pass "listener_b payloads include correct slot_id ${slot_b} exactly 4 times"
    else
        fail "listener_b payloads include slot_id ${slot_b} ${payload_slot_b} time(s) (expected 4)"
    fi

    if [[ "${foreign_slot_in_a}" == "0" ]]; then
        pass "listener_a payloads do not include foreign slot_id ${slot_b}"
    else
        fail "listener_a payloads include foreign slot_id ${slot_b} ${foreign_slot_in_a} time(s) (expected 0)"
    fi

    if [[ "${foreign_slot_in_b}" == "0" ]]; then
        pass "listener_b payloads do not include foreign slot_id ${slot_a}"
    else
        fail "listener_b payloads include foreign slot_id ${slot_a} ${foreign_slot_in_b} time(s) (expected 0)"
    fi

    # Verify expected score values appear once per listener
    local score expected_count actual_count
    # listener_a expected scores from inserts for slot_a
    for score in 10 8 6 4; do
        expected_count=1
        actual_count=$(grep -c "\"score\":${score}\b" "${listener_a_log}" || true)
        if [[ "${actual_count}" == "${expected_count}" ]]; then
            pass "listener_a payloads include score=${score} exactly once"
        else
            fail "listener_a payloads include score=${score} ${actual_count} time(s) (expected ${expected_count})"
        fi
    done

    # listener_b expected scores from inserts for slot_b
    for score in 9 7 5 3; do
        expected_count=1
        actual_count=$(grep -c "\"score\":${score}\b" "${listener_b_log}" || true)
        if [[ "${actual_count}" == "${expected_count}" ]]; then
            pass "listener_b payloads include score=${score} exactly once"
        else
            fail "listener_b payloads include score=${score} ${actual_count} time(s) (expected ${expected_count})"
        fi
    done

    # Cleanup temp logs and DB entities
    rm -f "${listener_a_log}" "${listener_b_log}"
    cleanup_ids "$slot_a" "$slot_b" "$tid" "$sid" "$owner" "$archer_a" "$archer_b"
}

main() {
    ########################################
    # Main
    ########################################
    start_docker
    require_cmd psql
    load_env

    info "Connecting to postgres://${POSTGRES_USER}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}"

    # Sanity check connectivity
    if run_sql "SELECT 1;" >/dev/null 2>&1; then
        pass "Database connectivity OK"
    else
        fail "Cannot connect to database"
        exit 2
    fi

    test_unique_open_session_per_archer
    test_get_next_lane_function
    test_slot_shooting_participants_view
    test_get_available_targets_function
    test_get_available_targets_occupancy_steps_18
    test_get_slot_with_lane_function
    test_get_active_slot_id_function
    test_get_next_lane_with_full_targets
    test_session_shot_per_round_defaults_and_constraints
    test_notify_shot_insert_trigger

    echo
    hr 80
    stop_docker
    if [[ "$fail_count" -eq 0 ]]; then
        ok "All migration tests passed"
        exit 0
    else
        err "$fail_count test(s) failed"
        exit 1
    fi

}

main "$@"
