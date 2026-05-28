package camera

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// FsWebcam は fswebcam コマンドで USB カメラから静止画を撮影する。
// ラズパイ上で動作させる本番用の実装。
type FsWebcam struct {
	Device     string // 例: /dev/video0
	Resolution string // 例: 1280x720
}

func (f FsWebcam) Capture(ctx context.Context) ([]byte, error) {
	tmp, err := os.CreateTemp("", "capture-*.jpg")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	tmp.Close()

	// --no-banner: タイムスタンプ等のバナーを焼き込まない
	cmd := exec.CommandContext(ctx, "fswebcam",
		"-d", f.Device,
		"-r", f.Resolution,
		"--no-banner",
		tmp.Name(),
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("fswebcam failed: %w: %s", err, out)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		return nil, fmt.Errorf("read captured image: %w", err)
	}
	return data, nil
}
