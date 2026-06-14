package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestInspectAnthropicDesktopProbeCapture_Observed(t *testing.T) {
	body := []byte(`{
		"model":"claude-opus-4-8",
		"max_tokens":8,
		"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]
	}`)
	parsed, err := service.ParseGatewayRequest(service.NewRequestBodyRef(body), domain.PlatformAnthropic)
	require.NoError(t, err)

	info := inspectAnthropicDesktopProbeCapture(parsed)
	require.True(t, info.Observed)
	require.Equal(t, "claude-opus-4-8", info.Model)
	require.Equal(t, 8, info.MaxTokens)
	require.Equal(t, 1, info.MessagesCount)
	require.Equal(t, "user", info.FirstRole)
	require.Equal(t, "hello", info.FirstText)
	require.Equal(t, 5, info.FirstTextLen)
}

func TestInspectAnthropicDesktopProbeCapture_StillObservesEvenIfLaterClassifiedAsClaudeCode(t *testing.T) {
	body := []byte(`{
		"model":"claude-opus-4-8",
		"max_tokens":8,
		"messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]
	}`)
	parsed, err := service.ParseGatewayRequest(service.NewRequestBodyRef(body), domain.PlatformAnthropic)
	require.NoError(t, err)

	info := inspectAnthropicDesktopProbeCapture(parsed)
	require.True(t, info.Observed)
}
