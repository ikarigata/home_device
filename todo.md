# TODO

## edge（ラズパイ）側の残課題

レビュー日: 2026-05-29 / 対象: `edge/`

### 🔴 直すべき（実害あり）

- [x] **再接続時にサブスクライブが復活しない**（最重要） — 対応済み 2026-05-29
  - 場所: `edge/internal/mqtt/client.go` / `edge/cmd/agent/main.go`
  - 対応: 購読を `SetOnConnectHandler` 内に移し、接続/再接続のたびに再購読するようにした。`CleanSession=true` のまま（切断中のトリガは破棄＝オンデマンド撮影に適切）。`SetConnectionLostHandler` で切断ログも追加。

- [x] **capture にタイムアウトと多重実行ガードが無い** — 対応済み 2026-05-29
  - 場所: `edge/cmd/agent/main.go` の `capture`
  - 対応: capture+upload を `context.WithTimeout`（30秒）でラップ。`sync.Mutex.TryLock` で実行中の重複トリガをドロップ。

### 🟡 あると安心（運用面）

- [x] **systemd unit ファイルを同梱する** — 対応済み 2026-05-29
  - 追加: `edge/deploy/home_device.service`（`Restart=always` / `After=network-online.target time-sync.target` / `EnvironmentFile=` / セキュリティ・ハードニング）。
  - README にインストール手順（専用ユーザー作成・配置・enable・journalctl）を追記。

- [x] **時刻同期（NTP）依存への対応** — 対応済み 2026-05-29
  - unit に `After=time-sync.target` を設定。README に `timedatectl set-ntp true` を案内。

### 🟢 任意（将来 / 軽微）

- [x] **撮影完了の通知**（パターンA: capturedAt ポーリング） — 対応済み 2026-05-30
  - lambda image が `HeadObject` で `capturedAt` (S3 LastModified) を返すように拡張
  - frontend は撮影リクエスト後、0.5秒間隔×最大10秒のポーリングで「ボタン押下時刻より新しい capturedAt」を待つ（時計ズレ吸収のため5秒マージン）
  - 最新モードのタイムスタンプ表示も実撮影時刻に統一（履歴と表記揃え）
- [x] **アップロードの一過性失敗にリトライ** — 対応済み 2026-05-30
  - `edge/cmd/agent/main.go` に `putWithRetry` を追加。`PutJPEG` 失敗時に1秒待って1回だけ再試行（`runCtx` キャンセル時は即抜け）。history / latest 双方の PutObject 呼び出しに適用。
