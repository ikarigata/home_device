// Package config はエッジエージェントの設定を環境変数から読み込む。
package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DeviceID   string // この機器の識別子（MQTTトピック/ S3キーに使用）
	CameraType string // "fswebcam"（実機）または "mock"（実機なし）
	AWSRegion  string

	// S3
	S3Bucket    string
	S3KeyPrefix string // 例: images

	// AWS IoT Core (MQTT over TLS, デバイス証明書による相互認証)
	IoTEndpoint string // 例: xxxxxxxx-ats.iot.ap-northeast-1.amazonaws.com
	CertFile    string // デバイス証明書 (.pem.crt)
	KeyFile     string // 秘密鍵 (.pem.key)
	CACertFile  string // Amazon Root CA
}

// CaptureTopic は撮影トリガを受け取る MQTT トピック。
func (c Config) CaptureTopic() string {
	return fmt.Sprintf("home_device/%s/capture", c.DeviceID)
}

// LatestKey は最新画像を保存する S3 オブジェクトキー。
func (c Config) LatestKey() string {
	return fmt.Sprintf("%s/%s/latest.jpg", strings.Trim(c.S3KeyPrefix, "/"), c.DeviceID)
}

// UsesIoT は IoT Core への接続が設定されているかを返す。
// 実機なしの開発時にエンドポイント未設定なら false。
func (c Config) UsesIoT() bool {
	return c.IoTEndpoint != ""
}

func Load() Config {
	return Config{
		DeviceID:    env("DEVICE_ID", "kitchen-1"),
		CameraType:  env("CAMERA", "fswebcam"),
		AWSRegion:   env("AWS_REGION", "ap-northeast-1"),
		S3Bucket:    env("S3_BUCKET", ""),
		S3KeyPrefix: env("S3_KEY_PREFIX", "images"),
		IoTEndpoint: env("IOT_ENDPOINT", ""),
		CertFile:    env("IOT_CERT_FILE", ""),
		KeyFile:     env("IOT_KEY_FILE", ""),
		CACertFile:  env("IOT_CA_FILE", ""),
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
