package ffmpeg

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

type EmbedSubtitlesParams struct {
	VideoPath  string
	SRTPath    string
	OutputPath string
	Hardcode   bool
}

// EmbedSubtitles embeds subtitles into video, either hardcoded or soft
func EmbedSubtitles(ctx context.Context, params EmbedSubtitlesParams) error {
	var cmd *exec.Cmd

	if params.Hardcode {
		// Hardcode subtitles (burn into video)
		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-hwaccel", "auto",
			"-i", params.VideoPath,
			"-vf", fmt.Sprintf("subtitles=%s", params.SRTPath),
			"-c:a", "copy",
			"-c:v", "libx264",
			"-preset", "fast",
			"-f", "mp4",
			"-y",
			params.OutputPath,
		)
	} else {
		// Soft embed subtitles
		cmd = exec.CommandContext(ctx, "ffmpeg",
			"-hwaccel", "auto",
			"-i", params.VideoPath,
			"-i", params.SRTPath,
			"-c:v", "copy",
			"-c:a", "copy",
			"-c:s", "mov_text", // subtitle codec for mp4
			"-f", "mp4",
			"-y",
			params.OutputPath,
		)
	}

	var buf bytes.Buffer
	cmd.Stderr = &buf
	if err := cmd.Run(); err != nil {
		errMsg := buf.String()
		return fmt.Errorf("failed to embed subs: %w. stderr: %s", err, errMsg)
	}

	return nil
}
