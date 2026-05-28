// Package mqtt は AWS IoT Core への MQTT 接続(デバイス証明書による相互TLS認証)を扱う。
package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
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
}

// Connect は相互TLSで IoT Core に接続したクライアントを返す。
func Connect(p Params) (paho.Client, error) {
	tlsCfg, err := newTLSConfig(p)
	if err != nil {
		return nil, err
	}

	opts := paho.NewClientOptions().
		AddBroker(fmt.Sprintf("ssl://%s:8883", p.Endpoint)).
		SetClientID(p.ClientID).
		SetTLSConfig(tlsCfg).
		SetAutoReconnect(true)

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
