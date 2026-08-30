package mediautil

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

// GenerateVideoThumb generates a thumbnail for a video and returns the temporary file path.
// The caller is responsible for deleting the file after use.
func GenerateVideoThumb(ctx context.Context, videoPath string, durSec int, w int, h int) (string, error) {
	if videoPath == "" {
		return "", fmt.Errorf("videoPath is empty")
	}

	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return "", fmt.Errorf("ffmpeg not found in PATH")
	}

	outFile, err := os.CreateTemp("", "tdl_thumb_*.jpg")
	if err != nil {
		return "", err
	}
	outPath := outFile.Name()
	outFile.Close()
	os.Remove(outPath) // Let ffmpeg create it

	halfTime := durSec / 2

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-ss", fmt.Sprintf("%d", halfTime),
		"-i", videoPath,
		"-vframes", "1",
		"-vf", "scale=320:-1",
		"-f", "mjpeg",
		outPath,
	)

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return outPath, nil
}
