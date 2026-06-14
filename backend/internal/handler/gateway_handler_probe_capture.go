package handler

import (
	"strings"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

const (
	anthropicDesktopProbeCaptureModel           = "claude-opus-4-8"
	anthropicDesktopProbeCaptureObservedMaxToks = 64
	anthropicDesktopProbeCapturePreviewMaxRunes = 120
)

type anthropicDesktopProbeContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicDesktopProbeMessage struct {
	Role    string                              `json:"role"`
	Content []anthropicDesktopProbeContentBlock `json:"content"`
}

type anthropicDesktopProbeCapture struct {
	Observed      bool
	Model         string
	MaxTokens     int
	MessagesCount int
	FirstRole     string
	FirstText     string
	FirstTextLen  int
}

func inspectAnthropicDesktopProbeCapture(parsed *service.ParsedRequest) anthropicDesktopProbeCapture {
	info := anthropicDesktopProbeCapture{}
	if parsed == nil {
		return info
	}

	info.Model = strings.TrimSpace(parsed.Model)
	if !strings.EqualFold(info.Model, anthropicDesktopProbeCaptureModel) {
		return info
	}

	var messages []anthropicDesktopProbeMessage
	if err := parsed.DecodeMessages(&messages); err != nil {
		return info
	}
	info.MaxTokens = parsed.MaxTokens
	info.MessagesCount = len(messages)
	if len(messages) == 0 {
		return info
	}

	info.FirstRole = strings.TrimSpace(messages[0].Role)
	info.FirstText = extractAnthropicDesktopProbeFirstText(messages[0].Content)
	info.FirstTextLen = utf8.RuneCountInString(info.FirstText)
	if len(messages) == 1 && info.FirstTextLen > 0 && info.MaxTokens > 0 && info.MaxTokens <= anthropicDesktopProbeCaptureObservedMaxToks {
		info.Observed = true
	}
	return info
}

func extractAnthropicDesktopProbeFirstText(content []anthropicDesktopProbeContentBlock) string {
	for _, block := range content {
		if strings.EqualFold(strings.TrimSpace(block.Type), "text") {
			return strings.TrimSpace(block.Text)
		}
	}
	return ""
}

func anthropicDesktopProbeCapturePreview(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= anthropicDesktopProbeCapturePreviewMaxRunes {
		return text
	}
	return string(runes[:anthropicDesktopProbeCapturePreviewMaxRunes]) + "..."
}
