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

## ラズパイ側の前提（実機到着後）

- `sudo apt install fswebcam`
- IoT デバイス証明書一式を `IOT_CERT_FILE` / `IOT_KEY_FILE` / `IOT_CA_FILE` のパスに配置
- systemd サービス化して常駐（アウトバウンドのみ、SSH ポートは非公開）
