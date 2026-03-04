# CONTRIBUTING

## 目的

外部コントリビューターが参加しやすい入口を維持しつつ、
HomeDash のスコープ（家庭ホワイトボード）を守ることを目的にします。

## Issue / PR の基本

- Issue はテンプレート（バグ / 改善 / 質問）を利用してください
- PR は `.github/pull_request_template.md` に沿って記載してください
- MVP0範囲外（天気/室温/株価/献立/在庫/IoT等）の追加は、まずIssueで合意してください

## 推奨ラベル設計

### type

- `type: bug`
- `type: enhancement`
- `type: docs`

### area

- `area: ui`
- `area: api`
- `area: ops`

### priority

- `priority: p0`
- `priority: p1`
- `priority: p2`

### onboarding

- `good first issue`

## ラベル運用ルール

- 1Issueにつき `type` は1つ
- `area` は主領域を1つ（必要なら複数可）
- `priority` は保守者が最終判断

## ラベル作成コマンド例（任意）

GitHub CLI が使える場合は以下で作成できます。

```bash
gh label create "type: bug" --color "d73a4a" --description "不具合"
gh label create "type: enhancement" --color "a2eeef" --description "改善提案"
gh label create "type: docs" --color "0075ca" --description "ドキュメント"
gh label create "area: ui" --color "5319e7" --description "UI関連"
gh label create "area: api" --color "1d76db" --description "API関連"
gh label create "area: ops" --color "fbca04" --description "運用関連"
gh label create "priority: p0" --color "b60205" --description "最優先"
gh label create "priority: p1" --color "d93f0b" --description "高"
gh label create "priority: p2" --color "f9d0c4" --description "中"
gh label create "good first issue" --color "7057ff" --description "初めて向け"
```
