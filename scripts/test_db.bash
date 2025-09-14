#!/usr/bin/env bash

set -euo pipefail

# Lightweight validation tests for SQL functions/views:
# - validate_face_row
# - get_shot_score
# - session_performance view invariants
#
# Usage: scripts/test_db.bash

ROOT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
# shellcheck source=./lib/manage_docker
. "${ROOT_DIR}/scripts/lib/manage_docker"
# shellcheck source=../.env
. "${ROOT_DIR}/.env"

: "${POSTGRES_USER:?Environment variable POSTGRES_USER is not set}"
: "${POSTGRES_DB:?Environment variable POSTGRES_DB is not set}"
: "${POSTGRES_SOCKET_DIR:?Environment variable POSTGRES_SOCKET_DIR is not set}"

log() { printf '[INFO] %s\n' "$*"; }
pass() { printf '\e[32m[PASS]\e[0m %s\n' "$*"; }
fail() {
    printf '\e[31m[FAIL]\e[0m %s\n' "$*"
    FAILED=true
}

psql_exec() {
    local sql="$1"
    psql -v ON_ERROR_STOP=1 -tA -F $'\t' -U "${POSTGRES_USER}" -d "${POSTGRES_DB}" -c "${sql}"
}

assert_equals() {
    local name="$1" expected="$2" sql="$3" out
    if ! out=$(psql_exec "${sql}" 2>&1); then
        fail "${name}: query failed: ${out}"
        return 1
    fi
    out=$(echo -n "${out}" | tr -d '\n')
    if [[ "${out}" == "${expected}" ]]; then
        pass "${name}"
        return 0
    fi
    fail "${name}: expected='${expected}' got='${out}'"
}

assert_is_null() {
    local name="$1" sql="$2" out
    if ! out=$(psql_exec "SELECT (sub.q) IS NULL FROM ( ${sql} ) sub(q);" 2>&1); then
        fail "${name}: query failed: ${out}"
        return 1
    fi
    out=$(echo -n "${out}" | tr -d '\n')
    if [[ ${out} == "t" ]]; then
        pass "${name}"
        return 0
    fi
    fail "${name}: expected NULL"
}

assert_is_null_direct() {
    local name="$1" sql="$2" out
    if ! out=$(psql_exec "${sql}" 2>&1); then
        fail "${name}: query failed: ${out}"
        return 1
    fi
    out=$(echo -n "${out}" | tr -d '\n')
    if [[ -z ${out} || ${out} == "NULL" ]]; then
        pass "${name}"
        return 0
    fi
    fail "${name}: expected NULL got '${out}'"
}

assert_command_fails() {
    local name="$1" sql="$2"
    if psql_exec "${sql}" >/dev/null 2>&1; then
        fail "${name}: command succeeded"
        return 1
    fi
    pass "${name}"
    return 0
}

# --------------------
# Test helpers & setup
# --------------------

# Globals used by tests
wa_faces_id='417d902f-3a4b-4262-a530-6d36f3dd47e7'
target_id=""

ensure_target_id() {
    if [[ -n ${target_id} ]]; then return 0; fi
    target_id=$(psql_exec 'SELECT id FROM targets LIMIT 1;') || true
    target_id=$(echo -n "${target_id}" | tr -d '\n')
    if [[ -z ${target_id} ]]; then
        fail 'No target found (seed data missing?)'
        return 1
    fi
}

fetch_current_target() {
    # Refresh and populate: current_target_id, current_max_x, current_max_y
    psql_exec 'REFRESH MATERIALIZED VIEW current_target;' >/dev/null 2>&1 || true
    local ct_row
    ct_row=$(psql_exec "SELECT id || ',' || max_x || ',' || max_y FROM current_target LIMIT 1;") || true
    ct_row=$(echo -n "${ct_row}" | tr -d '\n')
    if [[ -z ${ct_row} ]]; then
        fail 'No current_target row found (need open session)'
        return 1
    fi
    IFS=',' read -r current_target_id current_max_x current_max_y <<<"${ct_row}"
    export current_target_id current_max_x current_max_y
}

# --------------------
# validate_face_row tests
# --------------------

test_validate_face_row_inside() {
    ensure_target_id || return 0
    assert_equals 'validate_face_row valid inside' 't' \
        "SELECT validate_face_row('${target_id}',25,25,'${wa_faces_id}');"
}

test_validate_face_row_too_close_min_edge() {
    ensure_target_id || return 0
    assert_equals 'validate_face_row too close to min edge' 'f' \
        "SELECT validate_face_row('${target_id}',10,10,'${wa_faces_id}');"
}

test_validate_face_row_exceeds_max_edge() {
    ensure_target_id || return 0
    assert_equals 'validate_face_row exceeds max edge' 'f' \
        "SELECT validate_face_row('${target_id}',115,50,'${wa_faces_id}');"
}

test_faces_check_constraint_enforcement() {
    ensure_target_id || return 0
    assert_command_fails 'faces CHECK constraint enforcement' \
        "INSERT INTO faces (x,y,human_identifier,is_reduced,wa_faces_id,target_id) VALUES (10,10,'INVALID_SQL_TEST',FALSE,'${wa_faces_id}','${target_id}');"
}

# --------------------
# get_shot_score tests
# --------------------

test_get_shot_score_null_coords() {
    fetch_current_target || return 0
    assert_is_null 'get_shot_score NULL coords' \
        "SELECT get_shot_score(NULL,NULL,'${current_target_id}',${current_max_x},${current_max_y})"
}

test_get_shot_score_negative_coord() {
    fetch_current_target || return 0
    assert_is_null 'get_shot_score negative coord' \
        "SELECT get_shot_score(-1,10,'${current_target_id}',${current_max_x},${current_max_y})"
}

test_get_shot_score_center_10() {
    fetch_current_target || return 0
    assert_equals 'get_shot_score center 10' '10' \
        "SELECT get_shot_score(30,30,'${current_target_id}',${current_max_x},${current_max_y});"
}

test_get_shot_score_outside_faces_zero() {
    fetch_current_target || return 0
    assert_equals 'get_shot_score outside faces => 0' '0' \
        "SELECT get_shot_score(119,119,'${current_target_id}',${current_max_x},${current_max_y});"
}

test_get_shot_score_no_faces_null() {
    # Create session + target without faces and evaluate directly
    assert_is_null_direct 'get_shot_score no faces' \
        "WITH s AS (
            INSERT INTO sessions (is_opened,start_time,end_time,location,is_indoor)
            VALUES (FALSE, now(), now(), 'TempNoFaces',FALSE)
            RETURNING id
        ), t AS (
            INSERT INTO targets (max_x,max_y,distance,session_id,target_sensor_id)
            SELECT 50,50,10,id, uuid_generate_v4() FROM s
            RETURNING id,max_x,max_y
        )
        SELECT get_shot_score(10,10,t.id,t.max_x,t.max_y) FROM t;"
}

# -------------------------------
# session_performance view tests
# -------------------------------

test_session_performance_count_matches_shots() {
    fetch_current_target || return 0
    assert_equals 'session_performance count matches shots count' 't' "
        SELECT (
            SELECT COUNT(*) FROM session_performance
        ) = (
            SELECT COUNT(*) FROM shots s
            WHERE s.target_sensor_id = (SELECT target_sensor_id FROM current_target LIMIT 1)
        );"
}

test_session_performance_score_null_when_no_landing() {
    fetch_current_target || return 0
    assert_equals 'session_performance: score NULL when no landing' '0' "
        SELECT COUNT(*) FROM session_performance
        WHERE arrow_landing_time IS NULL AND score IS NOT NULL;"
}

test_session_performance_score_within_bounds() {
    fetch_current_target || return 0
    assert_equals 'session_performance: score within [0,10]' '0' "
        SELECT COUNT(*) FROM session_performance
        WHERE score IS NOT NULL AND (score < 0 OR score > 10);"
}

test_session_performance_speed_distance_time_relation() {
    fetch_current_target || return 0
    assert_equals 'session_performance: speed-distance-time relation holds' '0' "
        SELECT COUNT(*) FROM session_performance
        WHERE arrow_landing_time IS NOT NULL
          AND ABS(
                arrow_speed * time_of_flight_seconds -
                (SELECT distance::float FROM current_target LIMIT 1)
          ) > 1e-6;"
}

test_session_performance_human_identifier_matches_arrows() {
    fetch_current_target || return 0
    assert_equals 'session_performance: human_identifier matches arrows' '0' "
        SELECT COUNT(*)
        FROM session_performance sp
        JOIN arrows a ON a.id = sp.arrow_id
        WHERE sp.human_identifier IS DISTINCT FROM a.human_identifier;"
}

# ---------------
# Main test runner
# ---------------

main() {
    FAILED=false
    start_docker

    log 'Testing validate_face_row'
    test_validate_face_row_inside
    test_validate_face_row_too_close_min_edge
    test_validate_face_row_exceeds_max_edge
    test_faces_check_constraint_enforcement

    log 'Testing get_shot_score'
    test_get_shot_score_null_coords
    test_get_shot_score_negative_coord
    test_get_shot_score_center_10
    test_get_shot_score_outside_faces_zero
    test_get_shot_score_no_faces_null

    log 'Testing session_performance view'
    test_session_performance_count_matches_shots
    test_session_performance_score_null_when_no_landing
    test_session_performance_score_within_bounds
    test_session_performance_speed_distance_time_relation
    test_session_performance_human_identifier_matches_arrows

    if [[ ${FAILED} == true ]]; then
        echo
        echo 'Some SQL function tests FAILED'
        exit 1
    fi
    echo
    echo 'All SQL function tests PASSED'
    exit 0
}

main "$@"
