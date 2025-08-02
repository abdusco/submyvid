package openai

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/imroc/req/v3"
	"github.com/rs/zerolog/log"
)

type Client struct {
	apiKey string
	client *req.Client
}

type TranscribeResult struct {
	SRT            string
	ProcessingTime time.Duration
}

func NewClient(apiKey string) *Client {
	client := req.C().
		SetBaseURL("https://api.openai.com/v1").
		SetCommonHeader("Authorization", "Bearer "+apiKey)

	return &Client{
		apiKey: apiKey,
		client: client,
	}
}

func (c *Client) Transcribe(ctx context.Context, audioPath string) (TranscribeResult, error) {
	res, err := c.client.R().
		SetContext(ctx).
		SetFile("file", audioPath).
		SetFormData(map[string]string{
			"model":           "whisper-1",
			"response_format": "srt",
		}).
		Post("/audio/transcriptions")

	if err != nil {
		log.Ctx(ctx).Error().Err(err).Msg("failed to send POST request")
		return TranscribeResult{}, fmt.Errorf("request failed: %w", err)
	}

	log.Ctx(ctx).Debug().Int("status_code", res.GetStatusCode()).Msg("received response")

	if res.IsErrorState() {
		log.Ctx(ctx).Error().Int("status_code", res.GetStatusCode()).Str("response", res.String()).Msg("api returned error")
		return TranscribeResult{}, fmt.Errorf("api request failed with status %d: %s", res.GetStatusCode(), res.String())
	}

	result := TranscribeResult{
		SRT: res.String(),
	}

	// Parse processing time from headers
	if processingMs := res.Header.Get("openai-processing-ms"); processingMs != "" {
		if ms, err := strconv.Atoi(processingMs); err == nil {
			result.ProcessingTime = time.Duration(ms) * time.Millisecond
		}
	}

	log.Ctx(ctx).Debug().Int("srt_length", len(result.SRT)).Msg("transcription completed successfully")
	return result, nil
}
