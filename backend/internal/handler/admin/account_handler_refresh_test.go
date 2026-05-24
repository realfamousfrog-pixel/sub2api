package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

type openaiOAuthClientRefreshFailureStub struct {
	err error
}

func (s *openaiOAuthClientRefreshFailureStub) ExchangeCode(ctx context.Context, code, codeVerifier, redirectURI, proxyURL, clientID string) (*openai.TokenResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *openaiOAuthClientRefreshFailureStub) RefreshToken(ctx context.Context, refreshToken, proxyURL string) (*openai.TokenResponse, error) {
	return nil, s.err
}

func (s *openaiOAuthClientRefreshFailureStub) RefreshTokenWithClientID(ctx context.Context, refreshToken, proxyURL string, clientID string) (*openai.TokenResponse, error) {
	return nil, s.err
}

func TestRefreshSingleAccountMarksNonRetryableManualRefreshFailureAsError(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       42,
			Name:     "claude-oauth",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"access_token": "still-present",
			},
		},
	}

	handler := NewAccountHandler(
		adminSvc,
		service.NewOAuthService(nil, nil),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	account, err := adminSvc.GetAccount(context.Background(), 42)
	require.NoError(t, err)

	updated, warning, refreshErr := handler.refreshSingleAccount(context.Background(), account)
	require.Nil(t, updated)
	require.Empty(t, warning)
	require.Error(t, refreshErr)
	require.Contains(t, refreshErr.Error(), "no refresh token available")

	require.Equal(t, 1, adminSvc.lastSetAccountError.calls)
	require.Equal(t, int64(42), adminSvc.lastSetAccountError.id)
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "Token refresh failed (non-retryable)")
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "no refresh token available")

	latest, err := adminSvc.GetAccount(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, service.StatusError, latest.Status)
	require.Contains(t, latest.ErrorMessage, "no refresh token available")
}

func TestRefreshSingleAccountMarksOpenAINoRefreshTokenAsError(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       77,
			Name:     "openai-oauth",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"access_token": "still-present-access-token",
				"client_id":    "client-id-1",
			},
		},
	}

	handler := NewAccountHandler(
		adminSvc,
		nil,
		service.NewOpenAIOAuthService(nil, nil),
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	account, err := adminSvc.GetAccount(context.Background(), 77)
	require.NoError(t, err)

	updated, warning, refreshErr := handler.refreshSingleAccount(context.Background(), account)
	require.Nil(t, updated)
	require.Empty(t, warning)
	require.Error(t, refreshErr)
	require.Contains(t, refreshErr.Error(), "no refresh token available")

	require.Equal(t, 1, adminSvc.lastSetAccountError.calls)
	require.Equal(t, int64(77), adminSvc.lastSetAccountError.id)
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "Token refresh failed (non-retryable)")
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "no refresh token available")

	latest, err := adminSvc.GetAccount(context.Background(), 77)
	require.NoError(t, err)
	require.Equal(t, service.StatusError, latest.Status)
	require.Contains(t, latest.ErrorMessage, "no refresh token available")
}

func TestRefreshSingleAccountMarksOpenAIRefreshTokenReusedAsError(t *testing.T) {
	adminSvc := newStubAdminService()
	adminSvc.accounts = []service.Account{
		{
			ID:       88,
			Name:     "openai-oauth-reused",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeOAuth,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"access_token":  "still-present-access-token",
				"refresh_token": "refresh-token-1",
				"client_id":     "client-id-1",
			},
		},
	}

	openaiSvc := service.NewOpenAIOAuthService(nil, &openaiOAuthClientRefreshFailureStub{
		err: errors.New("token refresh failed: status 401, body: {\"error\":{\"message\":\"Your refresh token has already been used to generate a new access token. Please try signing in again.\",\"type\":\"invalid_request_error\",\"code\":\"refresh_token_reused\"}}"),
	})

	handler := NewAccountHandler(
		adminSvc,
		nil,
		openaiSvc,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
	)

	account, err := adminSvc.GetAccount(context.Background(), 88)
	require.NoError(t, err)

	updated, warning, refreshErr := handler.refreshSingleAccount(context.Background(), account)
	require.Nil(t, updated)
	require.Empty(t, warning)
	require.Error(t, refreshErr)
	require.Contains(t, refreshErr.Error(), "refresh_token_reused")

	require.Equal(t, 1, adminSvc.lastSetAccountError.calls)
	require.Equal(t, int64(88), adminSvc.lastSetAccountError.id)
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "Token refresh failed (non-retryable)")
	require.Contains(t, adminSvc.lastSetAccountError.errorMsg, "refresh_token_reused")

	latest, err := adminSvc.GetAccount(context.Background(), 88)
	require.NoError(t, err)
	require.Equal(t, service.StatusError, latest.Status)
	require.Contains(t, latest.ErrorMessage, "refresh_token_reused")
}
