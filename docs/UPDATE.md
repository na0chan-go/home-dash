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

### 2. アプリ状態の確認

`AUTH_TOKEN` の有無で確認手順が変わります。

#### `AUTH_TOKEN` を設定している場合

`/api/v1/status` を確認します。

```bash
curl http://localhost:8080/api/v1/status \
  -H "Authorization: Bearer <token>"
```

確認ポイント:

- `db.ok` が `true`
- `config.garbageScheduleLoaded` が `true`
- `auth.enabled` が想定どおり
- `lastBackup` が極端に古くない

#### `AUTH_TOKEN` を使わない場合

`/api/v1/status` と `/api/v1/admin/backup` は利用できないため、`/api/v1/health`、`/api/v1/dashboard`、ログで確認します。

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/dashboard
docker compose logs --tail=50 app
```

確認ポイント:

- `/api/v1/health` が `200` を返す
- `/api/v1/dashboard` が `200` を返す
- ログに DB や設定読込のエラーが出ていない

### 3. 手動バックアップ実行

更新前バックアップは必須です。  
バックアップに失敗した場合は、更新を中止してください。

#### `AUTH_TOKEN` を設定している場合

```bash
curl -X POST http://localhost:8080/api/v1/admin/backup \
  -H "Authorization: Bearer <token>"
```

#### `AUTH_TOKEN` を使わない場合

稼働中の SQLite を直接コピーしないため、先にアプリを停止してから `./data/app.db` を退避します。

```bash
docker compose stop app
mkdir -p ./data/backups
cp ./data/app.db ./data/backups/app-$(date +%Y%m%d-%H%M%S)-pre-update.db
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

以下のスクリプトでも更新できます。  
このスクリプトは「更新前バックアップ -> 再ビルド -> 更新後確認」を順に実行し、途中で失敗したら停止します。

```bash
./scripts/update.sh
```

`AUTH_TOKEN` を設定している場合は Bearer 付きで `/api/v1/status` を確認します。`.env` に `AUTH_TOKEN` を書いている場合も自動で読み取ります。  
`AUTH_TOKEN` を使わない場合は、アプリ停止後に `DB_PATH`（`.env` の設定も含む）に対応する DB ファイルを退避したうえで `/api/v1/health` と `/api/v1/dashboard` を確認します。

必要に応じて `APP_URL`、待機回数、待機秒数、`.env` のパスを上書きできます。

```bash
APP_URL=http://localhost:8080 WAIT_RETRIES=12 WAIT_SECONDS=5 AUTH_TOKEN=<token> ./scripts/update.sh
ENV_FILE=.env.production ./scripts/update.sh
```

## 3. 更新後チェック

### 1. UI表示確認

- ブラウザで `http://localhost:8080` を開く
- ダッシュボードが表示されることを確認する

### 2. アプリ状態の再確認

#### `AUTH_TOKEN` を設定している場合

```bash
curl http://localhost:8080/api/v1/status \
  -H "Authorization: Bearer <token>"
```

確認ポイント:

- `db.ok` が `true`
- `config.garbageScheduleLoaded` が `true`
- `serverTime` が更新されている

#### `AUTH_TOKEN` を使わない場合

```bash
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/dashboard
docker compose logs --tail=50 app
```

確認ポイント:

- `/api/v1/health` が `200` を返す
- `/api/v1/dashboard` が `200` を返す
- ログに DB や設定読込のエラーが出ていない

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
- `AUTH_TOKEN` を使わない運用では、`scripts/update.sh` が `docker compose stop app` を行ってから更新前バックアップを作成します
