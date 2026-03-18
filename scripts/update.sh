#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APP_URL=${APP_URL:-http://localhost:8080}
AUTH_TOKEN=${AUTH_TOKEN:-}
WAIT_RETRIES=${WAIT_RETRIES:-12}
WAIT_SECONDS=${WAIT_SECONDS:-5}
DATA_DIR=${DATA_DIR:-"$ROOT_DIR/data"}
BACKUP_DIR=${BACKUP_DIR:-"$DATA_DIR/backups"}
ENV_FILE=${ENV_FILE:-"$ROOT_DIR/.env"}

BACKUP_URL="$APP_URL/api/v1/admin/backup"
STATUS_URL="$APP_URL/api/v1/status"
HEALTH_URL="$APP_URL/api/v1/health"
DASHBOARD_URL="$APP_URL/api/v1/dashboard"

body_matches() {
  body=$1
  pattern=$2
  printf '%s\n' "$body" | grep -Eq "$pattern"
}

strip_wrapping_quotes() {
  value=$1
  case "$value" in
    \"*\")
      value=${value#\"}
      value=${value%\"}
      ;;
    \'*\')
      value=${value#\'}
      value=${value%\'}
      ;;
  esac
  printf '%s\n' "$value"
}

load_env_value_from_env_file() {
  key=$1
  if [ ! -f "$ENV_FILE" ]; then
    return 1
  fi

  line=$(grep -E "^[[:space:]]*$key=" "$ENV_FILE" | tail -n 1 || true)
  if [ -z "$line" ]; then
    return 1
  fi

  value=${line#*=}
  value=$(printf '%s\n' "$value" | sed 's/[[:space:]]*$//')
  strip_wrapping_quotes "$value"
}

resolve_auth_token() {
  if [ -n "$AUTH_TOKEN" ]; then
    printf '%s\n' "$AUTH_TOKEN"
    return 0
  fi

  load_env_value_from_env_file AUTH_TOKEN || true
}

resolve_container_db_path() {
  if [ -n "${DB_PATH:-}" ]; then
    printf '%s\n' "$DB_PATH"
    return 0
  fi

  value=$(load_env_value_from_env_file DB_PATH || true)
  if [ -n "$value" ]; then
    printf '%s\n' "$value"
    return 0
  fi

  printf '%s\n' "/data/app.db"
}

resolve_host_db_path() {
  container_db_path=$1
  case "$container_db_path" in
    /data/*)
      printf '%s/%s\n' "$DATA_DIR" "${container_db_path#/data/}"
      ;;
    *)
      printf '%s\n' "$container_db_path"
      ;;
  esac
}

docker_compose() {
  (
    cd "$ROOT_DIR"
    docker compose "$@"
  )
}

backup_before_update() {
  if [ -n "$EFFECTIVE_AUTH_TOKEN" ]; then
    echo "[1/3] 更新前バックアップを実行します"
    curl --fail --silent --show-error \
      -X POST "$BACKUP_URL" \
      -H "Authorization: Bearer $EFFECTIVE_AUTH_TOKEN"
    printf '\n'
    return
  fi

  echo "[1/3] AUTH_TOKEN 未設定のため、アプリを停止してから更新前バックアップを作成します"
  docker_compose stop app >/dev/null

  if [ ! -f "$HOST_DB_PATH" ]; then
    echo "DB ファイルが見つかりません: $HOST_DB_PATH" >&2
    exit 1
  fi

  mkdir -p "$BACKUP_DIR"
  backup_path="$BACKUP_DIR/app-$(date +%Y%m%d-%H%M%S)-pre-update.db"
  cp "$HOST_DB_PATH" "$backup_path"
  echo "バックアップを作成しました: $backup_path"
}

wait_for_health() {
  attempt=1
  while [ "$attempt" -le "$WAIT_RETRIES" ]; do
    if health_body=$(curl --silent --show-error --fail "$HEALTH_URL" 2>/dev/null) &&
      body_matches "$health_body" '"status"[[:space:]]*:[[:space:]]*"ok"'; then
      echo "health 確認に成功しました"
      return 0
    fi

    sleep "$WAIT_SECONDS"
    attempt=$((attempt + 1))
  done

  echo "health 確認に失敗しました。アプリ起動完了を確認できませんでした。" >&2
  return 1
}

wait_for_dashboard() {
  attempt=1
  while [ "$attempt" -le "$WAIT_RETRIES" ]; do
    if dashboard_body=$(curl --silent --show-error --fail "$DASHBOARD_URL" 2>/dev/null) &&
      body_matches "$dashboard_body" '"generatedAt"[[:space:]]*:' &&
      body_matches "$dashboard_body" '"today"[[:space:]]*:' &&
      body_matches "$dashboard_body" '"tomorrow"[[:space:]]*:'; then
      printf '%s\n' "$dashboard_body"
      return 0
    fi

    sleep "$WAIT_SECONDS"
    attempt=$((attempt + 1))
  done

  echo "dashboard 確認に失敗しました。DB または設定読込が正常ではない可能性があります。" >&2
  return 1
}

wait_for_status() {
  attempt=1
  while [ "$attempt" -le "$WAIT_RETRIES" ]; do
    if status_body=$(curl --silent --show-error --fail "$STATUS_URL" -H "Authorization: Bearer $EFFECTIVE_AUTH_TOKEN" 2>/dev/null) &&
      body_matches "$status_body" '"ok"[[:space:]]*:[[:space:]]*true' &&
      body_matches "$status_body" '"garbageScheduleLoaded"[[:space:]]*:[[:space:]]*true'; then
      printf '%s\n' "$status_body"
      return 0
    fi

    sleep "$WAIT_SECONDS"
    attempt=$((attempt + 1))
  done

  echo "status 確認に失敗しました。DB または設定読込が正常ではない可能性があります。" >&2
  if [ -n "${status_body:-}" ]; then
    printf '%s\n' "$status_body" >&2
  fi
  return 1
}

EFFECTIVE_AUTH_TOKEN=$(resolve_auth_token)
CONTAINER_DB_PATH=$(resolve_container_db_path)
HOST_DB_PATH=$(resolve_host_db_path "$CONTAINER_DB_PATH")
backup_before_update

echo "[2/3] docker compose up --build -d を実行します"
docker_compose up --build -d

if [ -n "$EFFECTIVE_AUTH_TOKEN" ]; then
  echo "[3/3] 更新後の status を確認します"
  wait_for_status
else
  echo "[3/3] 更新後の health と dashboard を確認します"
  wait_for_health
  wait_for_dashboard
fi

echo "更新処理は完了しました。UI表示と主要操作を続けて確認してください。"
