package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/abdusco/submyvid/ffmpeg"
	"github.com/abdusco/submyvid/openai"
	"github.com/abdusco/submyvid/subtitle"
	"github.com/alecthomas/kong"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type cliArgs struct {
	VideoPath string  `arg:"" help:"Path to the video file" type:"path"`
	Output    string  `short:"o" help:"Output video file path (default: adds suffix to input)"`
	Embed     bool    `help:"Embed subtitles into video"`
	Hardcode  bool    `help:"Hardcode subtitles into video instead of soft embedding"`
	Speed     float64 `help:"Audio speed-up factor for transcription (reduces costs)" default:"2.5"`
	APIKey    string  `env:"OPENAI_API_KEY" help:"OpenAI API key" required:"true"`
	Debug     bool    `help:"Enable debug logging"`
}

func main() {
	var args cliArgs
	cliCtx := kong.Parse(&args)
	_ = cliCtx

	// Set up logger with hh:mm:ss timestamps
	level := zerolog.InfoLevel
	if args.Debug {
		level = zerolog.DebugLevel
	}
	consoleWriter := zerolog.NewConsoleWriter()
	consoleWriter.TimeFormat = time.TimeOnly
	logger := zerolog.New(consoleWriter).Level(level)
	log.Logger = logger
	zerolog.DefaultContextLogger = &logger
	if err := run(context.Background(), &args); err != nil {
		log.Fatal().Err(err).Send()
	}
}

func run(ctx context.Context, args *cliArgs) error {
	log.Ctx(ctx).Debug().Msg("starting video processing")

	// Verify ffmpeg is available
	if err := ffmpeg.VerifyInstalled(); err != nil {
		return err
	}

	tempDir, err := os.MkdirTemp("", "submyvid-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)
	log.Ctx(ctx).Debug().Str("temp_dir", tempDir).Msg("created temp directory")

	log.Ctx(ctx).Info().Str("video_path", args.VideoPath).Msg("extracting audio")

	// Extract and speed up audio
	audioPath := filepath.Join(tempDir, "audio.m4a")
	if err := ffmpeg.ExtractAudio(args.VideoPath, audioPath, args.Speed); err != nil {
		return fmt.Errorf("failed to extract audio: %w", err)
	}

	log.Ctx(ctx).Info().Str("audio_path", audioPath).Msg("extracted audio, transcribing with whisper...")

	// Transcribe audio using OpenAI Whisper
	client := openai.NewClient(args.APIKey)
	result, err := client.Transcribe(ctx, audioPath)
	if err != nil {
		return fmt.Errorf("transcription failed: %w", err)
	}

	// Log processing details
	if result.ProcessingTime > 0 {
		log.Ctx(ctx).Info().Str("processing_time", result.ProcessingTime.String()).Msg("transcription completed")
	}

	// Save intermediate SRT for debugging
	intermediateSRTPath := filepath.Join(tempDir, "raw_transcription.srt")
	if err := os.WriteFile(intermediateSRTPath, []byte(result.SRT), 0644); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msg("failed to save intermediate SRT")
	}
	log.Ctx(ctx).Debug().Str("subtitle_path", intermediateSRTPath).Msg("saved intermediate SRT")

	log.Ctx(ctx).Info().Msg("processing subtitles...")

	// Parse SRT
	sub, err := subtitle.Parse(result.SRT)
	if err != nil {
		return fmt.Errorf("failed to parse SRT: %w", err)
	}
	log.Ctx(ctx).Debug().Int("total_entries", len(sub.Entries)).Msg("parsed SRT entries")

	// Slow down by speed factor (reverse the speedup)
	sub.AdjustSpeed(args.Speed)
	finalSRT, err := sub.String()
	if err != nil {
		return fmt.Errorf("failed to render SRT: %w", err)
	}

	var outDir string
	if args.Output != "" {
		outDir = filepath.Dir(args.Output)
	} else {
		outDir = filepath.Dir(args.VideoPath)
	}

	videoBasename := strings.TrimSuffix(filepath.Base(args.VideoPath), filepath.Ext(args.VideoPath))
	srtSavePath := filepath.Join(outDir, videoBasename+".srt")
	if err := os.WriteFile(srtSavePath, []byte(finalSRT), 0644); err != nil {
		return fmt.Errorf("failed to save final SRT file: %w", err)
	}
	log.Ctx(ctx).Info().Str("srt_path", srtSavePath).Msg("saved subtitle file")

	if args.Embed {
		var videoSavePath string
		if args.Output != "" {
			videoSavePath = args.Output
		} else {
			videoSavePath = filepath.Join(outDir, videoBasename+".embedded.mp4")
		}

		log.Ctx(ctx).Info().Str("output_path", videoSavePath).Bool("hardcode", args.Hardcode).Msg("embedding subtitles into video")

		if err := ffmpeg.EmbedSubtitles(ctx, ffmpeg.EmbedSubtitlesParams{
			VideoPath:  args.VideoPath,
			SRTPath:    srtSavePath,
			OutputPath: videoSavePath,
			Hardcode:   args.Hardcode,
		}); err != nil {
			return fmt.Errorf("failed to embed subtitles: %w", err)
		}
	}

	log.Ctx(ctx).Info().Msg("video processing completed successfully")

	return nil
}
