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

ラズパイには長期 IAM アクセスキーを置かず、**デバイス証明書を一時 AWS クレデンシャルに交換**して S3 にアップロードする。`IOT_CRED_ENDPOINT` と `IOT_ROLE_ALIAS` を設定すると有効化され、未設定なら AWS SDK 標準のクレデンシャルチェーン（`~/.aws` 等）にフォールバックする（ローカル開発時はこちら）。

```sh
# terraform apply 後に取得して .env / systemd EnvironmentFile に設定
terraform output -raw iot_cred_endpoint   # → IOT_CRED_ENDPOINT
terraform output -raw iot_role_alias      # → IOT_ROLE_ALIAS
terraform output -raw device_thing_name   # → IOT_THING_NAME
```

> credentials provider のエンドポイント（`xxxx.credentials.iot...`）は MQTT のデータエンドポイント（`xxxx-ats.iot...`）とは別物。証明書・鍵・CA は MQTT と同じものを使う。

## ラズパイ側の前提（実機到着後）

- `sudo apt install fswebcam`
- 時刻同期(NTP)を有効化: `sudo timedatectl set-ntp true`（TLS / SigV4 署名が時計に依存するため必須）
- IoT デバイス証明書一式を `IOT_CERT_FILE` / `IOT_KEY_FILE` / `IOT_CA_FILE` のパスに配置
  - 証明書・鍵: `terraform output -raw device_certificate_pem` / `device_private_key`
  - CA: `curl -o AmazonRootCA1.pem https://www.amazontrust.com/repository/AmazonRootCA1.pem`
- S3 認証用に `IOT_CRED_ENDPOINT` / `IOT_ROLE_ALIAS` / `IOT_THING_NAME` を設定（上記）

## systemd で常駐させる

unit ファイルは `deploy/home_device.service` に同梱。OS 起動時の自動開始・クラッシュ時の自動再起動・NTP 後の起動を行う。

```sh
# 1. 専用ユーザー作成（カメラアクセス用に video グループへ）
sudo useradd --system --no-create-home --shell /usr/sbin/nologin -G video home_device

# 2. バイナリ・設定・証明書を配置
sudo install -m 0755 agent /usr/local/bin/home_device-agent
sudo mkdir -p /etc/home_device/certs
sudo cp .env /etc/home_device/agent.env          # edge/.env.example をもとに作成
sudo cp device.pem.crt device.pem.key AmazonRootCA1.pem /etc/home_device/certs/
sudo chown -R root:home_device /etc/home_device
sudo chmod 0640 /etc/home_device/agent.env /etc/home_device/certs/*

# 3. unit を登録して起動
sudo cp deploy/home_device.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now home_device

# 4. ログ確認
sudo journalctl -u home_device -f
```

> アウトバウンド通信のみで動作（SSH ポートは非公開のまま）。`agent.env` には認証情報が入るためパーミッションは 0640・所有グループ `home_device` に限定する。
