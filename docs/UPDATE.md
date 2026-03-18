# UPDATE手順

HomeDash の更新は「更新前バックアップを必ず取得してから進める」ことを最優先にします。  
バックアップに失敗した場合は、更新を続行しません。

## 1. 更新前チェック

### 1. 現在のバージョン確認

現在のコミットまたはタグを確認します。

```bash
git rev-parse --short HEAD
git describe --tags --always
```

### 2. `/api/v1/status` の確認

更新前にアプリ状態を確認します。`AUTH_TOKEN` を設定している前提です。

```bash
curl http://localhost:8080/api/v1/status \
  -H "Authorization: Bearer <token>"
```

確認ポイント:

- `db.ok` が `true`
- `config.garbageScheduleLoaded` が `true`
- `auth.enabled` が想定どおり
- `lastBackup` が極端に古くない

### 3. 手動バックアップ実行

更新前バックアップは必須です。  
バックアップに失敗した場合は、更新を中止してください。

```bash
curl -X POST http://localhost:8080/api/v1/admin/backup \
  -H "Authorization: Bearer <token>"
```

## 2. 更新実行

### 1. 最新コード取得

```bash
git checkout main
git pull --ff-only
```

タグ運用している場合は、更新対象タグを明示してチェックアウトします。

```bash
git fetch --tags
git checkout v0.1.0
```

### 2. コンテナ再ビルド・再起動

```bash
docker compose up --build -d
```

### 3. ヘルパースクリプトを使う場合

`AUTH_TOKEN` を設定している場合は、以下でも更新できます。  
このスクリプトは「更新前バックアップ -> 再ビルド -> status確認」を順に実行し、途中で失敗したら停止します。

```bash
AUTH_TOKEN=<token> ./scripts/update.sh
```

必要に応じて `APP_URL` を上書きできます。

```bash
APP_URL=http://localhost:8080 AUTH_TOKEN=<token> ./scripts/update.sh
```

## 3. 更新後チェック

### 1. UI表示確認

- ブラウザで `http://localhost:8080` を開く
- ダッシュボードが表示されることを確認する

### 2. `/api/v1/status` 確認

```bash
curl http://localhost:8080/api/v1/status \
  -H "Authorization: Bearer <token>"
```

確認ポイント:

- `db.ok` が `true`
- `config.garbageScheduleLoaded` が `true`
- `serverTime` が更新されている

### 3. 主要操作確認

最低限、以下を確認します。

- 連絡追加
- 買い物追加
- ゴミ表示（今日・明日）

## 4. ロールバック

更新に失敗した場合は、ひとつ前のタグまたはコミットへ戻します。

```bash
docker compose down
git checkout <previous-tag-or-commit>
docker compose up --build -d
```

### `app.db` の復元が必要な場合

アプリ停止後に DB を差し替えます。

```bash
docker compose down
cp ./data/app.db ./data/app.db.before-restore
cp ./data/backups/app-YYYYMMDD-HHMMSS.db ./data/app.db
docker compose up --build -d
```

## 補足

- `scripts/update.sh` は自動ロールバックしません
- 更新前バックアップに失敗した場合は、その時点で更新を止める運用にしてください
- `AUTH_TOKEN` を使わない運用では、更新ヘルパーは使わず手順を手動で実施してください
