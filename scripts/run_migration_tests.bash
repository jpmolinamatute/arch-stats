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
    # $1 archer_id, $2 is_opened (true/false)
    local archer_id="$1"
    local opened="$2"
    local sql_statement="""
        INSERT INTO session (owner_archer_id, location, is_indoor, is_opened)
        VALUES ('$archer_id', 'Test Range', true, ${opened})
        RETURNING session_id;
"""
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
    if run_sql "INSERT INTO session (owner_archer_id, location, is_indoor, is_opened)
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
    header "slot_shooting_participants view"

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

    cnt="$(run_sql "SELECT count(*) FROM slot_shooting_participants WHERE session_id = '$s_open';")"
    cnt="${cnt//$'\n'/}"
    if [[ "$cnt" == "1" ]]; then
        pass "slot_shooting_participants returned 1 active shooter for open session"
    else
        fail "slot_shooting_participants returned '$cnt' (expected 1)"
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
