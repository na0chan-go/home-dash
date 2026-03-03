# AGENT.md

このファイルは HomeDash における AI（Codex 等）向けの開発ルールです。
必ず遵守してください。

---

## 🎯 プロジェクト目的

- 家庭内LANで動作するローカルWebアプリ
- リビング〜キッチン導線に置く「家庭ホワイトボード」を目指す
- 外部公開はしない（必要ならVPNのみ）
- Docker前提、データはローカル保存（SQLite / 設定ファイル）

---

## ✅ MVP0（最重要：この範囲だけを実装する）

MVP0は以下のみ実装する。勝手に機能を増やさない。

1. 共有メモ（連絡）

- 追加 / 一覧 / ピン留め / 削除（またはアーカイブ）

2. 買い物メモ

- 追加 / 一覧 / チェック（done）

3. ゴミ表示（今日・明日）

- 曜日固定ルール
- `config/garbage_schedule.json` から読み込む（DB化しない）
- 祝日や特例はMVP0では扱わない

---

## 🌐 フロント分離方針（将来Vue化）

- 現時点は薄いSSR（暫定UI）でも良いが、APIを必ず提供する
- APIは原則 `/api/v1/*` に集約する
- 画面表示用の整形は usecase のDTOで行い、handlerはDTOを返すだけにする
- 将来的に `web/` 配下へ Vue.js を追加してUIを置き換える前提とする

---

## 🏗 準Clean Architecture（依存方向）

依存関係の原則：

app → usecase → domain  
usecase → ports ← infra

### ルール

1. handler にビジネスロジックを書かない（呼び出しと入出力のみ）
2. usecase は ports（interface）にのみ依存する
3. infra は ports を実装するだけ（SQLite、設定ファイル読み込み等）
4. domain は外部パッケージに依存しない
5. HTTPアクセス時に外部APIを直接叩かない（MVP0では外部API自体を扱わない）

---

## 📂 ディレクトリ責務

- `internal/domain`：純粋モデル（メモ等）
- `internal/usecase`：ユースケース実装 + DTO
- `internal/ports`：Repository等のinterface
- `internal/infra`：ports実装（SQLite、設定ファイル）
- `internal/app`：HTTP、ルーティング、scheduler、DI（配線のみ）
- `internal/ui`：暫定SSRテンプレ・静的ファイル

---

## 💾 データ方針

- メモ類は SQLite（/data/app.db）に保存
- ゴミは `config/garbage_schedule.json` から読み込む（固定）
- .env はコミットしない（.env.exampleのみ）

---

## 🧠 実装手順（新規機能は縦スライス）

1. domain（モデル）
2. ports（interface）
3. usecase（ユースケース + DTO）
4. infra（実装）
5. handler（API）
6. SSR（必要なら）

---

## 📝 言語・記述方針

- README、コメント、PR、コミットメッセージは日本語
- ただし、Goの識別子（関数名・変数名・パッケージ名）は英語（慣習に従う）

---

## 🧾 コミット方針

- コミットメッセージは日本語
- prefixは任意で以下を使用可：feat / fix / refactor / docs / chore

例：

- `feat: 共有メモAPIを追加`
- `fix: ゴミ表示の曜日計算を修正`

---

## 🚫 禁止事項

- handler内にSQLを書く
- usecaseがinfraをimportする
- domainにDBタグ（ORM前提のタグ等）を書く
- 過剰設計（CQRS / Event Sourcing / マイクロサービス / DIフレームワーク導入）
- MVP0範囲外の機能追加（天気・室温・株価・献立・在庫・IoT等）

---

## 最重要原則

「動く」より「依存関係とMVP範囲」を守ることを優先する。
