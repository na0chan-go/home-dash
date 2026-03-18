#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
TEST_TMPDIR=$(mktemp -d)
trap 'rm -rf "$TEST_TMPDIR"' EXIT INT TERM

FAKEBIN="$TEST_TMPDIR/fakebin"
mkdir -p "$FAKEBIN"

cat <<'EOF' > "$FAKEBIN/docker"
#!/bin/sh
echo "docker $*" >> "$TEST_LOG"
exit 0
EOF
chmod +x "$FAKEBIN/docker"

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
  *)
    exit 1
    ;;
esac
EOF
chmod +x "$FAKEBIN/curl"

run_script() {
  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE="$1" \
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
  case_dir="$TEST_TMPDIR/token"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  run_script token "$case_dir" "secret" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "POST http://localhost:8080/api/v1/admin/backup"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/status"
  assert_contains "$case_dir/log" "docker compose up --build -d"
}

test_noauth_mode() {
  case_dir="$TEST_TMPDIR/noauth"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  run_script noauth "$case_dir" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/health"
  ls "$case_dir"/data/backups/app-*-pre-update.db >/dev/null 2>&1
}

test_degraded_status_fails() {
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

test_token_mode
test_noauth_mode
test_degraded_status_fails

echo "update.sh tests passed"
