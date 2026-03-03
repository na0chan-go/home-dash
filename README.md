# HomeDash

家庭内LANで動作する、家庭向けダッシュボードアプリです。

## 目的

- リビング〜キッチン導線で使える「家庭ホワイトボード」を提供する
- 外部公開せず、ローカルネットワーク内で運用する
- Docker前提で運用し、データはローカル保存する

## MVP0のスコープ

MVP0では次の3機能のみを実装します。

1. 共有メモ（連絡）
- 追加
- 一覧
- ピン留め
- 削除（またはアーカイブ）

2. 買い物メモ
- 追加
- 一覧
- チェック（done）

3. ゴミ表示（今日・明日）
- 曜日固定ルール
- `config/garbage_schedule.json` から読み込み
- 祝日や特例は扱わない

## 想定アーキテクチャ

依存方向は次を原則とします。

- `app -> usecase -> domain`
- `usecase -> ports <- infra`

主な責務は次のとおりです。

- `internal/domain`: 純粋モデル
- `internal/usecase`: ユースケース実装 + DTO
- `internal/ports`: Repository等のinterface
- `internal/infra`: ports実装（SQLite、設定ファイル読み込み）
- `internal/app`: HTTP、ルーティング、scheduler、DI（配線のみ）
- `internal/ui`: 暫定SSRテンプレ・静的ファイル

## データ方針

- メモ類: SQLite（`/data/app.db`）
- ゴミ表示: `config/garbage_schedule.json`
- `.env` はコミットしない（`.env.example`のみ）

## 開発メモ

- README、コメント、PR、コミットメッセージは日本語
- 新規機能は縦スライス（domain -> ports -> usecase -> infra -> handler -> SSR）で実装

