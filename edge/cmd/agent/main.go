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
	"syscall"

	"github.com/ikarigata/home_device/edge/internal/camera"
	"github.com/ikarigata/home_device/edge/internal/config"
	"github.com/ikarigata/home_device/edge/internal/mqtt"
	"github.com/ikarigata/home_device/edge/internal/uploader"

	paho "github.com/eclipse/paho.mqtt.golang"
)

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
	up, err := uploader.New(ctx, cfg.AWSRegion, cfg.S3Bucket)
	if err != nil {
		log.Fatalf("uploader init: %v", err)
	}

	capture := func() {
		log.Println("capture triggered")
		data, err := cam.Capture(ctx)
		if err != nil {
			log.Printf("capture error: %v", err)
			return
		}
		if err := up.PutJPEG(ctx, cfg.LatestKey(), data); err != nil {
			log.Printf("upload error: %v", err)
			return
		}
		log.Printf("uploaded %d bytes to s3://%s/%s", len(data), cfg.S3Bucket, cfg.LatestKey())
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
	})
	if err != nil {
		log.Fatalf("mqtt connect: %v", err)
	}
	defer client.Disconnect(250)

	topic := cfg.CaptureTopic()
	token := client.Subscribe(topic, 1, func(_ paho.Client, _ paho.Message) {
		capture()
	})
	if token.Wait(); token.Error() != nil {
		log.Fatalf("subscribe %s: %v", topic, token.Error())
	}
	log.Printf("subscribed to %s, waiting for triggers", topic)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("shutting down")
}
