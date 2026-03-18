#!/bin/sh

set -eu

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
APP_URL=${APP_URL:-http://localhost:8080}
AUTH_TOKEN=${AUTH_TOKEN:-}

if [ -z "$AUTH_TOKEN" ]; then
  echo "AUTH_TOKEN を設定してから実行してください。" >&2
  exit 1
fi

AUTH_HEADER="Authorization: Bearer $AUTH_TOKEN"
BACKUP_URL="$APP_URL/api/v1/admin/backup"
STATUS_URL="$APP_URL/api/v1/status"

echo "[1/3] 更新前バックアップを実行します"
curl --fail --silent --show-error \
  -X POST "$BACKUP_URL" \
  -H "$AUTH_HEADER"
printf '\n'

echo "[2/3] docker compose up --build -d を実行します"
cd "$ROOT_DIR"
docker compose up --build -d

echo "[3/3] 更新後の status を確認します"
curl --fail --silent --show-error \
  "$STATUS_URL" \
  -H "$AUTH_HEADER"
printf '\n'

echo "更新処理は完了しました。UI表示と主要操作を続けて確認してください。"
