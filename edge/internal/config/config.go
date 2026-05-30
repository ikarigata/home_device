// Package config はエッジエージェントの設定を環境変数から読み込む。
package config

import (
	"fmt"
	"os"
	"strings"
	"time"
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

	// IoT credentials provider (デバイス証明書 → S3 用一時クレデンシャル)
	IoTCredEndpoint string // 例: xxxxxxxx.credentials.iot.ap-northeast-1.amazonaws.com
	IoTRoleAlias    string // terraform output iot_role_alias
	IoTThingName    string // 任意。x-amzn-iot-thingname ヘッダに使用
}

// CaptureTopic は撮影トリガを受け取る MQTT トピック。
func (c Config) CaptureTopic() string {
	return fmt.Sprintf("home_device/%s/capture", c.DeviceID)
}

// LatestKey は最新画像を保存する S3 オブジェクトキー（毎回上書きされる pointer）。
func (c Config) LatestKey() string {
	return fmt.Sprintf("%s/%s/latest.jpg", strings.Trim(c.S3KeyPrefix, "/"), c.DeviceID)
}

// HistoryKey は履歴画像の S3 オブジェクトキー。
// images/{deviceId}/history/YYYY/MM/DD/HHMMSS.jpg の形式。
// 撮影時刻をキー自身に含めることで、Lambda 側で ListObjectsV2 によりタイムラインを構築できる。
func (c Config) HistoryKey(t time.Time) string {
	t = t.UTC()
	return fmt.Sprintf("%s/%s/history/%s.jpg",
		strings.Trim(c.S3KeyPrefix, "/"),
		c.DeviceID,
		t.Format("2006/01/02/150405"),
	)
}

// UsesIoT は IoT Core への接続が設定されているかを返す。
// 実機なしの開発時にエンドポイント未設定なら false。
func (c Config) UsesIoT() bool {
	return c.IoTEndpoint != ""
}

// UsesIoTCreds は S3 用クレデンシャルを IoT credentials provider から取得するかを返す。
// 未設定なら AWS SDK 標準のクレデンシャルチェーン（~/.aws、環境変数等）にフォールバックする。
func (c Config) UsesIoTCreds() bool {
	return c.IoTCredEndpoint != "" && c.IoTRoleAlias != ""
}

func Load() Config {
	return Config{
		DeviceID:        env("DEVICE_ID", "kitchen-1"),
		CameraType:      env("CAMERA", "fswebcam"),
		AWSRegion:       env("AWS_REGION", "ap-northeast-1"),
		S3Bucket:        env("S3_BUCKET", ""),
		S3KeyPrefix:     env("S3_KEY_PREFIX", "images"),
		IoTEndpoint:     env("IOT_ENDPOINT", ""),
		CertFile:        env("IOT_CERT_FILE", ""),
		KeyFile:         env("IOT_KEY_FILE", ""),
		CACertFile:      env("IOT_CA_FILE", ""),
		IoTCredEndpoint: env("IOT_CRED_ENDPOINT", ""),
		IoTRoleAlias:    env("IOT_ROLE_ALIAS", ""),
		IoTThingName:    env("IOT_THING_NAME", ""),
	}
}

func env(key, def string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return def
}
