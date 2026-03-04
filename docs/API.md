# API契約（v1）

## バージョニングと互換性

- APIは `/api/v1` を維持する
- 破壊的変更は `v2` で行う
- フィールド追加は互換性あり（クライアントは未知フィールドを無視する）
- フィールド削除・型変更は破壊的変更（`v1` では禁止）
- 日時は原則 ISO 8601 文字列
- エラーレスポンスの `timestamp` は Asia/Tokyo で返す

## 共通レスポンス方針

### 成功時

- `Content-Type: application/json`
- `Cache-Control: no-store`

### エラー時

全エンドポイント共通で以下の形式を返す。

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "..."
  },
  "requestId": "...",
  "timestamp": "2026-03-04T01:00:00+09:00"
}
```

エラーコード:

- `VALIDATION_ERROR`（400）
- `NOT_FOUND`（404）
- `UNAUTHORIZED`（401）
- `FORBIDDEN`（403）
- `INTERNAL_ERROR`（500）
- `CONFIG_ERROR`（500、設定ファイル起因）

## CORS方針

- デフォルトは無効（同一オリジン前提）
- `CORS_ALLOW_ORIGINS` 設定時のみ有効化
  - 例: `CORS_ALLOW_ORIGINS=http://localhost:5173`
- 許可メソッド: `GET, POST, PATCH, DELETE, OPTIONS`

## MVP0 エンドポイント一覧

### health

- `GET /api/v1/health`

### status

- `GET /api/v1/status`（Bearer必須）

### notes

- `GET /api/v1/notes?kind=notice|shopping&limit=50`
- `POST /api/v1/notes`
- `DELETE /api/v1/notes/:id`
- `PATCH /api/v1/notes/:id/pin`
- `PATCH /api/v1/notes/:id/done`

### garbage

- `GET /api/v1/garbage/today`
- `GET /api/v1/garbage/tomorrow`
- `GET /api/v1/garbage/summary`

### dashboard

- `GET /api/v1/dashboard`

## レスポンス例

### GET /api/v1/dashboard

```json
{
  "generatedAt": "2026-03-04T01:10:00+09:00",
  "notes": {
    "notice": [],
    "shopping": []
  },
  "garbage": {
    "today": {
      "date": "2026-03-04",
      "weekday": "wednesday",
      "items": [],
      "label": "なし"
    },
    "tomorrow": {
      "date": "2026-03-05",
      "weekday": "thursday",
      "items": ["燃えるゴミ"],
      "label": "燃えるゴミ"
    }
  }
}
```

### GET /api/v1/status

```json
{
  "appVersion": "unknown",
  "uptimeSeconds": 1200,
  "serverTime": "2026-03-05T09:00:00+09:00",
  "db": {
    "path": "/data/app.db",
    "ok": true
  },
  "config": {
    "garbageScheduleLoaded": true
  },
  "auth": {
    "enabled": true
  },
  "lastBackup": "2026-03-05T08:30:00+09:00"
}
```
