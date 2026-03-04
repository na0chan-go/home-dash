# RELEASE手順

HomeDash のリリース運用は「手動でも再現できること」を優先します。  
MVP0ではローカルビルド運用（`docker compose up --build`）を前提にします。

## 1. 事前準備

1. `main` が最新であることを確認
2. 変更内容を `CHANGELOG.md` の `Unreleased` に追記
3. リリース前バックアップを作成（推奨）

```bash
curl -X POST http://localhost:8080/api/v1/admin/backup \
  -H "Authorization: Bearer <token>"
```

## 2. バージョン確定

1. `CHANGELOG.md` の `Unreleased` 内容を `vX.Y.Z` セクションへ移動
2. `README.md` / `docs` との差分を確認
3. 変更をコミット

```bash
git add CHANGELOG.md README.md docs/
git commit -m "v0.1.1 リリース準備"
```

## 3. タグ作成と共有

```bash
git checkout main
git pull
git tag -a v0.1.1 -m "v0.1.1"
git push origin main
git push origin v0.1.1
```

## 4. 更新（ローカル運用）

対象端末で、動かしたいタグをチェックアウトして再ビルドします。

```bash
git fetch --tags
git checkout v0.1.1
docker compose up --build -d
```

将来、タグ付きイメージ配布（例: GHCR）に切り替えた場合は、以下で更新します。

```bash
docker compose pull
docker compose up -d
```

確認:

```bash
curl http://localhost:8080/api/v1/health
```

## 5. ロールバック手順

不具合時は 1つ前のタグへ戻します。

```bash
docker compose down
git checkout v0.1.0
docker compose up --build -d
```

## 6. DB復元手順（必要時のみ）

アプリ停止後に `app.db` を差し替えて復元します。

```bash
docker compose down
cp ./data/app.db ./data/app.db.before-restore
cp ./data/backups/app-YYYYMMDD-HHMMSS.db ./data/app.db
docker compose up --build -d
```

## Dockerタグ方針（MVP0）

- 現時点では Dockerイメージの外部配布は行いません
- 運用上の版管理は Gitタグ（`v0.1.0` など）で行います
- 将来、GHCR等で `home-dash:vX.Y.Z` を配布する場合は本手順書を更新します（現時点では未実装）
