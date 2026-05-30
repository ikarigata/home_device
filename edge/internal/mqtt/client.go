// Package mqtt は AWS IoT Core への MQTT 接続(デバイス証明書による相互TLS認証)を扱う。
package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"

	paho "github.com/eclipse/paho.mqtt.golang"
)

// Params は IoT Core 接続に必要な値。
type Params struct {
	Endpoint   string // xxxx-ats.iot.<region>.amazonaws.com
	ClientID   string // 一意なクライアントID（DeviceID を流用）
	CertFile   string
	KeyFile    string
	CACertFile string

	Topic     string              // 購読する撮影トリガトピック
	QoS       byte                // 購読 QoS
	OnMessage paho.MessageHandler // トリガ受信時に呼ばれるハンドラ
}

// Connect は相互TLSで IoT Core に接続したクライアントを返す。
// 購読は OnConnect ハンドラ内で行うため、自動再接続のたびに購読が復元される。
func Connect(p Params) (paho.Client, error) {
	tlsCfg, err := newTLSConfig(p)
	if err != nil {
		return nil, err
	}

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("ssl://%s:8883", p.Endpoint)).
		SetClientID(p.ClientID).
		SetTLSConfig(tlsCfg).
		SetAutoReconnect(true).
		// CleanSession=true: 切断中のトリガはキューせず破棄する（オンデマンド撮影では
		// 過去のボタン押下が再接続時に蘇るのは望ましくない）。その代わり再接続後は
		// 購読が失われるため、下の OnConnect で毎回購読し直す。
		SetCleanSession(true).
		SetConnectionLostHandler(func(_ paho.Client, err error) {
			log.Printf("mqtt connection lost: %v", err)
		}).
		SetOnConnectHandler(func(c paho.Client) {
			token := c.Subscribe(p.Topic, p.QoS, p.OnMessage)
			if token.Wait(); token.Error() != nil {
				log.Printf("subscribe %s failed: %v", p.Topic, token.Error())
				return
			}
			log.Printf("subscribed to %s", p.Topic)
		})

	client := paho.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}
	return client, nil
}

func newTLSConfig(p Params) (*tls.Config, error) {
	ca, err := os.ReadFile(p.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("read CA cert: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to parse CA cert")
	}

	cert, err := tls.LoadX509KeyPair(p.CertFile, p.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load device key pair: %w", err)
	}

	return &tls.Config{
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert},
	}, nil
}
