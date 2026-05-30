// Command agent はラズパイ上で常駐し、AWS IoT Core からの撮影トリガを待ち受けて
// 静止画を撮影し、S3 へアップロードする。
//
// 実機なしの開発時は CAMERA=mock を指定するとダミー画像で動作を確認できる。
// IOT_ENDPOINT が未設定の場合はローカルモードとなり、起動時に1回だけ撮影→アップロードして終了する。
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ikarigata/home_device/edge/internal/camera"
	"github.com/ikarigata/home_device/edge/internal/config"
	"github.com/ikarigata/home_device/edge/internal/iotcreds"
	"github.com/ikarigata/home_device/edge/internal/mqtt"
	"github.com/ikarigata/home_device/edge/internal/uploader"

	"github.com/aws/aws-sdk-go-v2/aws"
	paho "github.com/eclipse/paho.mqtt.golang"
)

// captureTimeout は1回の撮影+アップロードに許す上限。
// fswebcam のウォームアップと S3 アップロードがハングしても無限に待たないようにする。
const captureTimeout = 30 * time.Second

func main() {
	cfg := config.Load()
	log.Printf("starting agent: device=%s camera=%s bucket=%s", cfg.DeviceID, cfg.CameraType, cfg.S3Bucket)

	ctx := context.Background()

	cam, err := camera.New(cfg.CameraType)
	if err != nil {
		log.Fatalf("camera init: %v", err)
	}

	if cfg.S3Bucket == "" {
		log.Fatal("S3_BUCKET is required")
	}

	// S3 用クレデンシャル: IoT credentials provider が設定されていれば
	// デバイス証明書から一時クレデンシャルを取得する。未設定なら SDK 標準チェーン。
	var creds aws.CredentialsProvider
	if cfg.UsesIoTCreds() {
		p, err := iotcreds.NewProvider(iotcreds.Params{
			Endpoint:   cfg.IoTCredEndpoint,
			RoleAlias:  cfg.IoTRoleAlias,
			ThingName:  cfg.IoTThingName,
			CertFile:   cfg.CertFile,
			KeyFile:    cfg.KeyFile,
			CACertFile: cfg.CACertFile,
		})
		if err != nil {
			log.Fatalf("iot credentials provider init: %v", err)
		}
		creds = aws.NewCredentialsCache(p)
		log.Printf("using IoT credentials provider: role_alias=%s", cfg.IoTRoleAlias)
	} else {
		log.Println("IOT_CRED_ENDPOINT/IOT_ROLE_ALIAS not set: using default AWS credential chain")
	}

	up, err := uploader.New(ctx, cfg.AWSRegion, cfg.S3Bucket, creds)
	if err != nil {
		log.Fatalf("uploader init: %v", err)
	}

	// 撮影が走っている間に来た重複トリガは捨てる（MQTT ハンドラは単一 goroutine で
	// 動くため、撮影が長引くと後続メッセージを詰まらせてしまうのを防ぐ）。
	var captureMu sync.Mutex
	capture := func() {
		if !captureMu.TryLock() {
			log.Println("capture already in progress, skipping trigger")
			return
		}
		defer captureMu.Unlock()

		runCtx, cancel := context.WithTimeout(ctx, captureTimeout)
		defer cancel()

		log.Println("capture triggered")
		data, err := cam.Capture(runCtx)
		if err != nil {
			log.Printf("capture error: %v", err)
			return
		}

		// 履歴と latest 双方に保存（dual write）。履歴は撮影時刻入りキー、latest は pointer。
		historyKey := cfg.HistoryKey(time.Now())
		if err := up.PutJPEG(runCtx, historyKey, data); err != nil {
			log.Printf("history upload error: %v", err)
			return
		}
		if err := up.PutJPEG(runCtx, cfg.LatestKey(), data); err != nil {
			log.Printf("latest upload error: %v", err)
			return
		}
		log.Printf("uploaded %d bytes to s3://%s/%s and %s", len(data), cfg.S3Bucket, historyKey, cfg.LatestKey())
	}

	if !cfg.UsesIoT() {
		log.Println("IOT_ENDPOINT not set: running one-shot local mode")
		capture()
		return
	}

	client, err := mqtt.Connect(mqtt.Params{
		Endpoint:   cfg.IoTEndpoint,
		ClientID:   cfg.DeviceID,
		CertFile:   cfg.CertFile,
		KeyFile:    cfg.KeyFile,
		CACertFile: cfg.CACertFile,
		Topic:      cfg.CaptureTopic(),
		QoS:        1,
		OnMessage:  func(_ paho.Client, _ paho.Message) { capture() },
	})
	if err != nil {
		log.Fatalf("mqtt connect: %v", err)
	}
	defer client.Disconnect(250)

	log.Printf("connected, waiting for capture triggers on %s", cfg.CaptureTopic())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}
