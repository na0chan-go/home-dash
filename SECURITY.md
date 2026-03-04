# Security Policy

HomeDash は家庭内LAN利用を想定したアプリです。  
インターネット公開用途では設計されていません。

## Supported Versions

現在サポート対象は `v0.x` 系列のみです。

## Security Recommendations

安全運用のため、以下を推奨します。

- 外部公開しない
- VPN利用時は `AUTH_TOKEN` を設定する
- ポートフォワーディングを行わない
- リバースプロキシ公開は非推奨

## Reporting a Vulnerability

脆弱性を発見した場合は、次の手順で報告してください。

- 公開Issueではなくメンテナへ連絡する
- 公開Issueに詳細を書かない
- 修正後に公開する

連絡先が未整備の場合は、GitHub Issue に `Security report` として最小情報のみ記載してください。
