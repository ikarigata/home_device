package camera

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"time"
)

// Mock は実機なしで開発するためのダミーカメラ。
// 撮影時刻によって色味が変わる単色 JPEG を返し、撮影フローを通しで検証できる。
type Mock struct{}

func (Mock) Capture(ctx context.Context) ([]byte, error) {
	const w, h = 1280, 720
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	now := time.Now()
	c := color.RGBA{
		R: uint8(now.Hour() * 10 % 256),
		G: uint8(now.Minute() * 4 % 256),
		B: uint8(now.Second() * 4 % 256),
		A: 255,
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("encode mock jpeg: %w", err)
	}
	return buf.Bytes(), nil
}
