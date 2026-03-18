#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TEST_TMPDIR=$(mktemp -d)
trap 'rm -rf "$TEST_TMPDIR"' EXIT INT TERM

FAKEBIN="$TEST_TMPDIR/fakebin"
mkdir -p "$FAKEBIN"

cat <<'EOF' > "$FAKEBIN/docker"
#!/bin/sh
echo "PWD=$PWD docker $*" >> "$TEST_LOG"
if [ "${DOCKER_FAIL:-0}" = "1" ] && [ "$1" = "compose" ] && [ "$2" = "stop" ] && [ "$3" = "app" ]; then
  exit 1
fi
exit 0
EOF
chmod +x "$FAKEBIN/docker"

write_default_fake_curl() {
cat <<'EOF' > "$FAKEBIN/curl"
#!/bin/sh
url=""
method="GET"
while [ "$#" -gt 0 ]; do
  case "$1" in
    -X)
      shift
      method=$1
      ;;
    -H|--header)
      shift
      ;;
    --fail|--silent|--show-error)
      ;;
    *)
      url=$1
      ;;
  esac
  shift
done

echo "$method $url" >> "$TEST_LOG"

case "$TEST_MODE:$url" in
  token:http://localhost:8080/api/v1/admin/backup)
    printf '{"ok":true}\n'
    ;;
  token:http://localhost:8080/api/v1/status)
    printf '{ "db": { "ok": true }, "config": { "garbageScheduleLoaded": true } }\n'
    ;;
  degraded:http://localhost:8080/api/v1/admin/backup)
    printf '{"ok":true}\n'
    ;;
  degraded:http://localhost:8080/api/v1/status)
    printf '{ "db": { "ok": false }, "config": { "garbageScheduleLoaded": true } }\n'
    ;;
  noauth:http://localhost:8080/api/v1/health)
    printf '{ "status": "ok" }\n'
    ;;
  noauth:http://localhost:8080/api/v1/dashboard)
    printf '{ "generatedAt": "2026-03-18T12:00:00+09:00", "garbage": { "today": {}, "tomorrow": {} } }\n'
    ;;
  *)
    exit 1
    ;;
esac
EOF
chmod +x "$FAKEBIN/curl"
}

write_default_fake_curl

run_script() {
  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE="$1" \
    DOCKER_FAIL="${4:-0}" \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$2/data" \
    BACKUP_DIR="$2/data/backups" \
    DB_PATH="$2/data/app.db" \
    AUTH_TOKEN="${3:-}" \
    sh "$ROOT_DIR/scripts/update.sh"
}

assert_contains() {
  file=$1
  pattern=$2
  if ! grep -Fq "$pattern" "$file"; then
    echo "expected pattern not found: $pattern" >&2
    echo "--- file: $file ---" >&2
    cat "$file" >&2
    exit 1
  fi
}

test_token_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/token"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  run_script token "$case_dir" "secret" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "POST http://localhost:8080/api/v1/admin/backup"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/status"
  assert_contains "$case_dir/log" "PWD=$ROOT_DIR docker compose up --build -d"
}

test_noauth_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/noauth"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  run_script noauth "$case_dir" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "PWD=$ROOT_DIR docker compose stop app"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/health"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/dashboard"
  ls "$case_dir"/data/backups/app-*-pre-update.db >/dev/null 2>&1
}

test_degraded_status_fails() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/degraded"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  if run_script degraded "$case_dir" "secret" > "$case_dir/stdout" 2>"$case_dir/stderr"; then
    echo "expected degraded status run to fail" >&2
    exit 1
  fi

  assert_contains "$case_dir/stderr" "status 確認に失敗しました"
}

test_partial_status_body_fails() {
  case_dir="$TEST_TMPDIR/partial"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  cat <<'EOF' > "$FAKEBIN/curl"
#!/bin/sh
url=""
method="GET"
while [ "$#" -gt 0 ]; do
  case "$1" in
    -X)
      shift
      method=$1
      ;;
    -H|--header)
      shift
      ;;
    --fail|--silent|--show-error)
      ;;
    *)
      url=$1
      ;;
  esac
  shift
done

echo "$method $url" >> "$TEST_LOG"

case "$url" in
  http://localhost:8080/api/v1/admin/backup)
    printf '{"ok":true}\n'
    exit 0
    ;;
  http://localhost:8080/api/v1/status)
    printf '{ "db": { "ok": true }, "config": { "garbageScheduleLoaded": true }'
    exit 22
    ;;
  *)
    exit 1
    ;;
esac
EOF
  chmod +x "$FAKEBIN/curl"

  if run_script token "$case_dir" "secret" > "$case_dir/stdout" 2>"$case_dir/stderr"; then
    echo "expected partial status body run to fail" >&2
    exit 1
  fi

  assert_contains "$case_dir/stderr" "status 確認に失敗しました"
}

test_stop_failure_fails() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/stopfail"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  if run_script noauth "$case_dir" "" 1 > "$case_dir/stdout" 2>"$case_dir/stderr"; then
    echo "expected stop failure run to fail" >&2
    exit 1
  fi
}

test_env_auth_token_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/envtoken"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"
  env_file="$case_dir/.env"
  printf 'AUTH_TOKEN=from-env\n' > "$env_file"

  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE=token \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$case_dir/data" \
    BACKUP_DIR="$case_dir/data/backups" \
    DB_PATH="$case_dir/data/app.db" \
    ENV_FILE="$env_file" \
    AUTH_TOKEN="" \
    sh "$ROOT_DIR/scripts/update.sh" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "POST http://localhost:8080/api/v1/admin/backup"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/status"
}

test_token_mode
test_noauth_mode
test_degraded_status_fails
test_partial_status_body_fails
test_stop_failure_fails
test_env_auth_token_mode

echo "update.sh tests passed"
