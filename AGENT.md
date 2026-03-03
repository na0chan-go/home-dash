# AGENT.md

このファイルは HomeDash における AI（Codex 等）向けの開発ルールです。
必ず遵守してください。

## プロジェクト目的

- 家庭内LANで動作するローカルWebアプリ
- リビング〜キッチン導線に置く「家庭ホワイトボード」を目指す
- 外部公開はしない（必要ならVPNのみ）
- Docker前提、データはローカル保存（SQLite / 設定ファイル）

## 目的

MVP0（共有メモ/買い物メモ/ゴミ表示）を安全に積み上げるための土台を作る。
この時点では、機能実装は最小（`/health`、起動、DB、設定読込、配線）に留める。

## MVP0（最重要：この範囲だけを実装する）

1. 共有メモ（連絡）
- 追加 / 一覧 / ピン留め / 削除（またはアーカイブ）

2. 買い物メモ
- 追加 / 一覧 / チェック（done）

3. ゴミ表示（今日・明日）
- 曜日固定ルール
- `config/garbage_schedule.json` から読み込む（DB化しない）
- 祝日や特例はMVP0では扱わない

MVP0範囲外（天気/室温/株価/献立/在庫/IoT等）は実装しないこと。

## 準Clean Architecture（依存方向）

依存関係の原則：

- `app -> usecase -> domain`
- `usecase -> ports <- infra`

### ルール

1. handler にビジネスロジックを書かない（呼び出しと入出力のみ）
2. usecase は ports（interface）にのみ依存する
3. infra は ports を実装するだけ（SQLite、設定ファイル読み込み等）
4. domain は外部パッケージに依存しない
5. HTTPアクセス時に外部APIを直接叩かない（MVP0では外部API自体を扱わない）

## ディレクトリ責務

- `cmd/server`：起動点
- `internal/app`：HTTP、ルーティング、DI、config
- `internal/usecase`：ユースケース実装 + DTO
- `internal/domain`：純粋モデル
- `internal/ports`：Repository等のinterface
- `internal/infra`：ports実装（SQLite、設定ファイル）
- `internal/ui`：暫定SSRテンプレ・静的ファイル（必要最低限）

## データ方針

- メモ類は SQLite（`/data/app.db`）に保存
- ゴミは `config/garbage_schedule.json` から読み込む（固定）
- `.env` はコミットしない（`.env.example`のみ）

## DB基盤

- SQLite + migrate の仕組みを初期段階で確定する
- 起動時にマイグレーションを必ず実行する
- MVP0で必要になるテーブルはこの段階で未作成でも可
- DB接続は `infra` 側に閉じ込め、各所に散らさない

## API方針

- APIファーストで作る
- `/api/v1/health` はJSONで200を返す
- エラーレスポンスはJSONで統一する（簡易な枠でよい）
- SSRは暫定扱いとし、この段階では最小（無くても良い）

## 品質の最低ライン

- ハンドラにビジネスロジックを置かない
- usecaseはportsにのみ依存し、infraをimportしない
- domainは外部依存なし
- グローバル変数で依存を保持しない（手動DIで配線）

## 実装手順（新規機能は縦スライス）

1. domain（モデル）
2. ports（interface）
3. usecase（ユースケース + DTO）
4. infra（実装）
5. handler（API）
6. SSR（必要なら）

## 言語・記述方針

- README、コメント、PR、コミットメッセージは日本語
- Goの識別子（関数名・変数名・パッケージ名）は英語

## コミット方針

- コミットメッセージは日本語
- 成果は1コミットにまとめる

## 禁止事項

- handler内にSQLを書く
- usecaseがinfraをimportする
- domainにDBタグ（ORM前提タグ等）を書く
- 過剰設計（CQRS / Event Sourcing / マイクロサービス / DIフレームワーク導入）
- MVP0範囲外の機能追加

## 最重要原則

「動く」より「依存関係とMVP範囲」を守ることを優先する。
