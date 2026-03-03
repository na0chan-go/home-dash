# ARCHITECTURE

## 依存方向（準Clean）

- `app -> usecase -> domain`
- `usecase -> ports <- infra`

## レイヤ責務

- `internal/app`: ルーティング・handler・DI（配線）
- `internal/usecase`: ユースケース本体・DTO
- `internal/domain`: 純粋モデル
- `internal/ports`: インターフェース
- `internal/infra`: ports実装（SQLite、設定ファイル読み込み）

## ルール

- handler は薄く保ち、ビジネスロジックを書かない
- DTO は usecase 側に置く
- 設定ファイル読み込みは infra に閉じ込める

## MVP0スコープ

次の3機能のみ実装対象です。

- 共有メモ（notice）
- 買い物メモ（shopping）
- ゴミ表示（today / tomorrow）

MVP0範囲外の機能は追加しません。
