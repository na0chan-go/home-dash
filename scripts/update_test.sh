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
command_name=""
command_target=""
if [ "$1" = "compose" ] && [ "$2" = "--env-file" ]; then
  command_name=$4
  command_target=$5
elif [ "$1" = "compose" ]; then
  command_name=$2
  command_target=$3
fi
if [ "${DOCKER_FAIL:-0}" = "1" ] && [ "$command_name" = "stop" ] && [ "$command_target" = "app" ]; then
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
  mode=$1
  case_dir=$2
  auth_mode=${3:-unset}
  docker_fail=${4:-0}
  db_path=${5:-/data/app.db}
  env_file=${6:-$ROOT_DIR/.env}

  if [ "$auth_mode" = "__EMPTY__" ]; then
    env AUTH_TOKEN="" \
      PATH="$FAKEBIN:$PATH" \
      TEST_LOG="$TEST_LOG" \
      TEST_MODE="$mode" \
      DOCKER_FAIL="$docker_fail" \
      WAIT_RETRIES=1 \
      WAIT_SECONDS=0 \
      DATA_DIR="$case_dir/data" \
      BACKUP_DIR="$case_dir/data/backups" \
      DB_PATH="$db_path" \
      ENV_FILE="$env_file" \
      sh "$ROOT_DIR/scripts/update.sh"
    return
  fi

  if [ "$auth_mode" = "unset" ]; then
    env -u AUTH_TOKEN \
      PATH="$FAKEBIN:$PATH" \
      TEST_LOG="$TEST_LOG" \
      TEST_MODE="$mode" \
      DOCKER_FAIL="$docker_fail" \
      WAIT_RETRIES=1 \
      WAIT_SECONDS=0 \
      DATA_DIR="$case_dir/data" \
      BACKUP_DIR="$case_dir/data/backups" \
      DB_PATH="$db_path" \
      ENV_FILE="$env_file" \
      sh "$ROOT_DIR/scripts/update.sh"
    return
  fi

  env AUTH_TOKEN="$auth_mode" \
    PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE="$mode" \
    DOCKER_FAIL="$docker_fail" \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$case_dir/data" \
    BACKUP_DIR="$case_dir/data/backups" \
    DB_PATH="$db_path" \
    ENV_FILE="$env_file" \
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

  run_script noauth "$case_dir" unset > "$case_dir/stdout"

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

  if run_script noauth "$case_dir" unset 1 > "$case_dir/stdout" 2>"$case_dir/stderr"; then
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

  run_script token "$case_dir" unset 0 /data/app.db "$env_file" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "POST http://localhost:8080/api/v1/admin/backup"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/status"
  assert_contains "$case_dir/log" "docker compose --env-file $env_file up --build -d"
}

test_explicit_empty_auth_token_overrides_env() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/emptyauth"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"
  env_file="$case_dir/.env"
  printf 'AUTH_TOKEN=from-env\n' > "$env_file"

  run_script noauth "$case_dir" __EMPTY__ 0 /data/app.db "$env_file" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "PWD=$ROOT_DIR docker compose --env-file $env_file stop app"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/health"
  assert_contains "$case_dir/log" "GET http://localhost:8080/api/v1/dashboard"
}

test_env_db_path_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/envdb"
  mkdir -p "$case_dir/data/backups" "$case_dir/data/custom"
  printf 'db' > "$case_dir/data/custom/alt.db"
  TEST_LOG="$case_dir/log"
  env_file="$case_dir/.env"
  printf 'DB_PATH=/data/custom/alt.db\n' > "$env_file"

  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE=noauth \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$case_dir/data" \
    BACKUP_DIR="$case_dir/data/backups" \
    ENV_FILE="$env_file" \
    env -u AUTH_TOKEN sh "$ROOT_DIR/scripts/update.sh" > "$case_dir/stdout"

  ls "$case_dir"/data/backups/app-*-pre-update.db >/dev/null 2>&1
}

test_env_relative_db_path_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/envrel"
  mkdir -p "$case_dir/data/backups"
  TEST_LOG="$case_dir/log"
  env_file="$case_dir/.env"
  printf 'DB_PATH=tmp/rel.db\n' > "$env_file"

  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE=noauth \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$case_dir/data" \
    BACKUP_DIR="$case_dir/data/backups" \
    ENV_FILE="$env_file" \
    env -u AUTH_TOKEN sh "$ROOT_DIR/scripts/update.sh" > "$case_dir/stdout" 2>"$case_dir/stderr" && {
      echo "expected relative db path run to fail" >&2
      exit 1
    }

  assert_contains "$case_dir/stderr" "DB_PATH は /data 配下のみ対応しています"
}

test_explicit_empty_db_path_uses_default() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/emptydb"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"
  env_file="$case_dir/.env"
  printf 'DB_PATH=/data/custom/alt.db\n' > "$env_file"

  PATH="$FAKEBIN:$PATH" \
    TEST_LOG="$TEST_LOG" \
    TEST_MODE=noauth \
    WAIT_RETRIES=1 \
    WAIT_SECONDS=0 \
    DATA_DIR="$case_dir/data" \
    BACKUP_DIR="$case_dir/data/backups" \
    ENV_FILE="$env_file" \
    DB_PATH="" \
    env -u AUTH_TOKEN sh "$ROOT_DIR/scripts/update.sh" > "$case_dir/stdout"

  ls "$case_dir"/data/backups/app-*-pre-update.db >/dev/null 2>&1
}

test_relative_env_file_mode() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/relenv"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"
  printf 'AUTH_TOKEN=from-relative-env\n' > "$ROOT_DIR/.env.production"

  (
    cd "$ROOT_DIR/scripts"
    PATH="$FAKEBIN:$PATH" \
      TEST_LOG="$TEST_LOG" \
      TEST_MODE=token \
      WAIT_RETRIES=1 \
      WAIT_SECONDS=0 \
      DATA_DIR="$case_dir/data" \
      BACKUP_DIR="$case_dir/data/backups" \
      ENV_FILE=.env.production \
      env -u AUTH_TOKEN sh "$ROOT_DIR/scripts/update.sh" > "$case_dir/stdout"
  )

  assert_contains "$case_dir/log" "POST http://localhost:8080/api/v1/admin/backup"
  assert_contains "$case_dir/log" "docker compose --env-file $ROOT_DIR/.env.production up --build -d"
  rm -f "$ROOT_DIR/.env.production"
}

test_default_env_file_optional() {
  write_default_fake_curl
  case_dir="$TEST_TMPDIR/noenvfile"
  mkdir -p "$case_dir/data/backups"
  printf 'db' > "$case_dir/data/app.db"
  TEST_LOG="$case_dir/log"

  run_script noauth "$case_dir" > "$case_dir/stdout"

  assert_contains "$case_dir/log" "PWD=$ROOT_DIR docker compose stop app"
}

test_token_mode
test_noauth_mode
test_degraded_status_fails
test_partial_status_body_fails
test_stop_failure_fails
test_env_auth_token_mode
test_explicit_empty_auth_token_overrides_env
test_env_db_path_mode
test_env_relative_db_path_mode
test_explicit_empty_db_path_uses_default
test_relative_env_file_mode
test_default_env_file_optional

echo "update.sh tests passed"
