# edge — ラズパイ常駐エージェント

AWS IoT Core からの撮影トリガ（MQTT）を待ち受け、`fswebcam` で静止画を撮影して S3 にアップロードする Go 製エージェント。

## 動作モード

| モード | 条件 | 挙動 |
| --- | --- | --- |
| 本番 | `IOT_ENDPOINT` 設定あり | IoT Core に常時接続し `home_device/{DEVICE_ID}/capture` を subscribe。トリガ受信のたびに撮影→S3 アップロード |
| ローカル | `IOT_ENDPOINT` 未設定 | 起動時に1回だけ撮影→S3 アップロードして終了（S3 経路の動作確認用） |

`CAMERA=mock` を指定するとカメラ実機なしでダミー JPEG を生成する。実機がまだ無い段階で、Terraform で作った S3 への保存経路をローカルから検証できる。

## ローカル実行（実機・IoT なしで S3 経路だけ確認）

```sh
cp .env.example .env   # S3_BUCKET を terraform output で埋める
export $(grep -v '^#' .env | xargs)
CAMERA=mock IOT_ENDPOINT= go run ./cmd/agent
```

## ビルド（ラズパイ向けクロスコンパイル例）

```sh
# Raspberry Pi (64bit OS / arm64)
GOOS=linux GOARCH=arm64 go build -o agent ./cmd/agent
# Raspberry Pi (32bit OS / armv7)
GOOS=linux GOARCH=arm GOARM=7 go build -o agent ./cmd/agent
```

## S3 への認証（IoT credentials provider）

ラズパイには長期 IAM アクセスキーを置かず、**デバイス証明書を一時 AWS クレデンシャルに交換**して S3 にアップロードする。`bootstrap-edge.sh` が `IOT_CRED_ENDPOINT` / `IOT_ROLE_ALIAS` / `IOT_THING_NAME` を `agent.env` に自動で書き込む。

未設定で起動した場合は AWS SDK 標準のクレデンシャルチェーン（`~/.aws` 等）にフォールバックする（ローカル開発時はこちら）。

> credentials provider のエンドポイント（`xxxx.credentials.iot...`）は MQTT のデータエンドポイント（`xxxx-ats.iot...`）とは別物。証明書・鍵・CA は MQTT と同じものを使う。

## ラズパイ側の前提（実機到着後）

- `sudo apt install fswebcam openssl curl`
- 時刻同期(NTP)を有効化: `sudo timedatectl set-ntp true`（TLS / SigV4 署名が時計に依存するため必須）
- 開発機から `ssh ika` で到達できる状態にする（ラズパイ側で sudo はパスワードなしで使える前提）

## 初回セットアップ: `./bootstrap-edge.sh`（リポジトリ直下）

CSR ベースのデバイス証明書プロビジョニングを含む、初回フルセットアップを 1 コマンドで行う。

**ポイント**: 秘密鍵 (`device.pem.key`) はラズパイ上で生成され、その後一度も外に出ない。AWS へ送るのは公開情報である CSR のみ。tfstate にも秘密鍵は入らない。

```sh
./bootstrap-edge.sh
```

これが順に実行されるので、初回はこの 1 コマンドで以下まで完了する:

1. ラズパイに `home_device` ユーザー作成 + `/etc/home_device/certs` 準備
2. ラズパイ上で秘密鍵生成 → CSR 作成（既存鍵があれば再利用）
3. CSR を `terraform/device.csr` に取得
4. Lambda ビルド + `terraform apply` で AWS が CSR を署名 → cert 発行
5. 署名済み cert をラズパイへ配置（`/etc/home_device/certs/device.pem.crt`）
6. Amazon Root CA をラズパイへ配置
7. `terraform output` から `agent.env` を生成してラズパイへ配置
8. agent バイナリを arm64 でビルドして配置
9. systemd unit を登録 → `enable --now`

冪等なので何度実行しても安全（既存鍵・既存 CA はスキップされ、cert と env は最新化される）。

> CloudFront + frontend は別途 `./deploy.sh` でデプロイする。

## 2回目以降の更新

| 何を更新したい | コマンド |
| --- | --- |
| Lambda / Terraform / frontend | `./deploy.sh` |
| エッジバイナリのみ | `./deploy-edge.sh` |
| cert ローテーション | `./bootstrap-edge.sh`（鍵は維持、新しい CSR を投げる場合は `device.pem.key` を消してから） |
