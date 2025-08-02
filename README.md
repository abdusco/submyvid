# submyvid

Generate subtitles for videos using OpenAI Whisper API.

## Features

- Extract and transcribe audio from video files
- Generate SRT subtitle files
- Embed subtitles into videos (soft- or hardcoded)
- Speed up audio processing to reduce API costs
- Support for various video formats

## Requirements

- ffmpeg installed and available in `$PATH`
- OpenAI API key

## Installation

Download the latest release from the [releases page](https://github.com/abdusco/submyvid/releases).

```bash
# Generate subtitles only
submyvid video.mp4

# Embed soft subtitles into video
submyvid video.mp4 --embed

# Hardcoded subtitles
submyvid video.mp4 --embed --hardcode

# Custom output path
submyvid video.mp4 -o output.mp4 --embed

# Adjust audio speed (default: 2.5x for cost reduction), at the cost of accuracy
submyvid video.mp4 --speed 3.0
```

## Environment

Set your OpenAI API key:
```bash
export OPENAI_API_KEY=your_api_key_here
```
