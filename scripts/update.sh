#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APP_URL=${APP_URL:-http://localhost:8080}
AUTH_TOKEN=${AUTH_TOKEN:-}
WAIT_RETRIES=${WAIT_RETRIES:-12}
WAIT_SECONDS=${WAIT_SECONDS:-5}
DATA_DIR=${DATA_DIR:-"$ROOT_DIR/data"}
BACKUP_DIR=${BACKUP_DIR:-"$DATA_DIR/backups"}
DB_PATH=${DB_PATH:-"$DATA_DIR/app.db"}

AUTH_HEADER="Authorization: Bearer $AUTH_TOKEN"
BACKUP_URL="$APP_URL/api/v1/admin/backup"
STATUS_URL="$APP_URL/api/v1/status"
HEALTH_URL="$APP_URL/api/v1/health"

body_matches() {
  body=$1
  pattern=$2
  printf '%s\n' "$body" | grep -Eq "$pattern"
}

backup_before_update() {
  if [ -n "$AUTH_TOKEN" ]; then
    echo "[1/3] 更新前バックアップを実行します"
    curl --fail --silent --show-error \
      -X POST "$BACKUP_URL" \
      -H "$AUTH_HEADER"
    printf '\n'
    return
  fi

  echo "[1/3] AUTH_TOKEN 未設定のためファイルコピーで更新前バックアップを作成します"
  if [ ! -f "$DB_PATH" ]; then
    echo "DB ファイルが見つかりません: $DB_PATH" >&2
    exit 1
  fi

  mkdir -p "$BACKUP_DIR"
  backup_path="$BACKUP_DIR/app-$(date +%Y%m%d-%H%M%S)-pre-update.db"
  cp "$DB_PATH" "$backup_path"
  echo "バックアップを作成しました: $backup_path"
}

wait_for_health() {
  attempt=1
  while [ "$attempt" -le "$WAIT_RETRIES" ]; do
    health_body=$(curl --silent --show-error --fail "$HEALTH_URL" 2>/dev/null || true)
    if body_matches "$health_body" '"status"[[:space:]]*:[[:space:]]*"ok"'; then
      echo "health 確認に成功しました"
      return 0
    fi

    sleep "$WAIT_SECONDS"
    attempt=$((attempt + 1))
  done

  echo "health 確認に失敗しました。アプリ起動完了を確認できませんでした。" >&2
  return 1
}

wait_for_status() {
  attempt=1
  while [ "$attempt" -le "$WAIT_RETRIES" ]; do
    status_body=$(curl --silent --show-error --fail "$STATUS_URL" -H "$AUTH_HEADER" 2>/dev/null || true)
    if body_matches "$status_body" '"ok"[[:space:]]*:[[:space:]]*true' &&
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

backup_before_update

echo "[2/3] docker compose up --build -d を実行します"
cd "$ROOT_DIR"
docker compose up --build -d

if [ -n "$AUTH_TOKEN" ]; then
  echo "[3/3] 更新後の status を確認します"
  wait_for_status
else
  echo "[3/3] 更新後の health を確認します"
  wait_for_health
fi

echo "更新処理は完了しました。UI表示と主要操作を続けて確認してください。"
