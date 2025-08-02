package ffmpeg

import (
	"bytes"
	"fmt"
	"os/exec"
)

// VerifyInstalled checks if ffmpeg is available in PATH
func VerifyInstalled() error {
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH. Please install ffmpeg")
	}
	return nil
}

// ExtractAudio extracts audio from video with speed adjustment
func ExtractAudio(videoPath, audioPath string, speed float64) error {
	cmd := exec.Command("ffmpeg",
		"-y", // overwrite output
		"-i", videoPath,
		"-vn", // no video
		"-acodec", "aac",
		"-ar", "16000", // 16kHz sample rate
		"-ac", "1", // mono
		"-b:a", "32k", // 32k bitrate
		"-filter:a", fmt.Sprintf("atempo=%.1f", speed), // speed up
		audioPath,
	)

	var buf bytes.Buffer
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		errMsg := buf.String()
		return fmt.Errorf("failed to extract audio: %w. stderr: %s", err, errMsg)
	}

	return nil
}
