# 変更履歴

このファイルは HomeDash のリリース差分を記録します。  
形式は Keep a Changelog の考え方に沿って、`Unreleased` を先に更新します。

## 追記ルール

- 機能追加・修正を行ったら、まず `Unreleased` に追記する
- リリース時に `Unreleased` の内容を `vX.Y.Z` セクションへ移動する
- 日付は `YYYY-MM-DD` で記載する

## [Unreleased]

### 変更予定

- （ここに次回リリース予定の変更を追記）

## [v0.1.0] - 2026-03-04

### 追加

- 準Clean Architecture 構成（`cmd` / `internal` / `web`）を整備
- `SQLite + migrate` の起動基盤を追加（`/data/app.db` 永続化）
- `notes` API（連絡/買い物の追加・一覧・削除・pin・done）を追加
- `garbage` API（today/tomorrow/summary、設定ファイル読み込み）を追加
- `dashboard` 集約API（`/api/v1/dashboard`）を追加
- Vueフロントエンド（単一コンテナ静的配信）を追加
- タブレット向けUX改善（楽観的更新、誤操作防止、状態表示）を追加
- オフライン検出・スリープ復帰再取得・ポーリング制御を追加
- PWA対応（manifest、service worker、ホーム画面追加）を追加
- 簡易トークン認証（`AUTH_TOKEN` / Bearer）を追加
- SQLiteバックアップ/復元運用（定期・手動・世代管理）を追加
- 運用ステータスAPI（`/api/v1/status`）を追加

### 変更

- APIエラー形式を統一（`error.code` / `requestId` / `timestamp`）
- API契約・運用ドキュメントを日本語で整備

