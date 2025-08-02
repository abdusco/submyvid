package subtitle

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Subtitle struct {
	Entries []Entry
}

// AdjustSpeed adjusts subtitle timing by dividing by the given factor
func (s *Subtitle) AdjustSpeed(factor float64) {
	for i := range s.Entries {
		s.Entries[i].StartTime = time.Duration(float64(s.Entries[i].StartTime) * factor)
		s.Entries[i].EndTime = time.Duration(float64(s.Entries[i].EndTime) * factor)
	}
}

func (s *Subtitle) String() (string, error) {
	var result strings.Builder

	for i, entry := range s.Entries {
		if i > 0 {
			result.WriteString("\n\n")
		}

		result.WriteString(fmt.Sprintf("%d\n", entry.Index))
		result.WriteString(fmt.Sprintf("%s --> %s\n",
			formatTime(entry.StartTime),
			formatTime(entry.EndTime)))
		result.WriteString(entry.Text)
	}

	return result.String(), nil
}

type Entry struct {
	Index     int
	StartTime time.Duration
	EndTime   time.Duration
	Text      string
}

var srtTimeRegex = regexp.MustCompile(`(\d{2}):(\d{2}):(\d{2}),(\d{3}) --> (\d{2}):(\d{2}):(\d{2}),(\d{3})`)

func Parse(content string) (Subtitle, error) {
	var subtitle Subtitle

	// Split into blocks by double newlines
	blocks := strings.Split(strings.TrimSpace(content), "\n\n")

	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}

		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) < 3 {
			continue
		}

		// Parse index
		index, err := strconv.Atoi(strings.TrimSpace(lines[0]))
		if err != nil {
			continue
		}

		// Parse timing
		matches := srtTimeRegex.FindStringSubmatch(lines[1])
		if len(matches) != 9 {
			continue
		}

		startTime, err := parseTime(matches[1], matches[2], matches[3], matches[4])
		if err != nil {
			continue
		}

		endTime, err := parseTime(matches[5], matches[6], matches[7], matches[8])
		if err != nil {
			continue
		}

		// Parse text (remaining lines)
		text := strings.Join(lines[2:], "\n")

		subtitle.Entries = append(subtitle.Entries, Entry{
			Index:     index,
			StartTime: startTime,
			EndTime:   endTime,
			Text:      text,
		})
	}

	return subtitle, nil
}

func parseTime(hours, minutes, seconds, milliseconds string) (time.Duration, error) {
	h, err := strconv.Atoi(hours)
	if err != nil {
		return 0, err
	}

	m, err := strconv.Atoi(minutes)
	if err != nil {
		return 0, err
	}

	s, err := strconv.Atoi(seconds)
	if err != nil {
		return 0, err
	}

	ms, err := strconv.Atoi(milliseconds)
	if err != nil {
		return 0, err
	}

	return time.Duration(h)*time.Hour +
		time.Duration(m)*time.Minute +
		time.Duration(s)*time.Second +
		time.Duration(ms)*time.Millisecond, nil
}

func formatTime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	milliseconds := int(d.Milliseconds()) % 1000

	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, seconds, milliseconds)
}
