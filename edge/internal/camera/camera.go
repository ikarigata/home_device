// Package camera は静止画の撮影を抽象化する。
// 実機(fswebcam)とモック(実機なし開発用)の2実装を切り替えられる。
package camera

import (
	"context"
	"fmt"
)

// Camera は静止画を撮影して JPEG バイト列を返す。
type Camera interface {
	Capture(ctx context.Context) ([]byte, error)
}

// New は種別に応じた Camera を生成する。
//
//	"fswebcam" / "" : 実機(USBカメラ + fswebcam)
//	"mock"          : 実機なし。ダミー JPEG を生成
func New(kind string) (Camera, error) {
	switch kind {
	case "", "fswebcam":
		return FsWebcam{Device: "/dev/video0", Resolution: "1280x720"}, nil
	case "mock":
		return Mock{}, nil
	default:
		return nil, fmt.Errorf("unknown camera type: %q", kind)
	}
}
