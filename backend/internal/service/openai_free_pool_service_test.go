//go:build unit

package service

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type openAIFreePoolSettingRepoStub struct {
	values map[string]string
}

func (s *openAIFreePoolSettingRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *openAIFreePoolSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", ErrSettingNotFound
}

func (s *openAIFreePoolSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *openAIFreePoolSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *openAIFreePoolSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *openAIFreePoolSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	return map[string]string{}, nil
}

func (s *openAIFreePoolSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

type openAIFreePoolAdminServiceStub struct {
	groups            []Group
	proxies           []Proxy
	accounts          []Account
	bulkUpdateInputs  []*BulkUpdateAccountsInput
	bulkUpdateErrByID map[int64]error
}

func (s *openAIFreePoolAdminServiceStub) GetAllGroups(context.Context) ([]Group, error) {
	return s.groups, nil
}

func (s *openAIFreePoolAdminServiceStub) GetAllProxies(context.Context) ([]Proxy, error) {
	return s.proxies, nil
}

func (s *openAIFreePoolAdminServiceStub) ListAccounts(context.Context, int, int, string, string, string, string, int64, string, string, string) ([]Account, int64, error) {
	items := append([]Account(nil), s.accounts...)
	return items, int64(len(items)), nil
}

func (s *openAIFreePoolAdminServiceStub) BulkUpdateAccounts(_ context.Context, input *BulkUpdateAccountsInput) (*BulkUpdateAccountsResult, error) {
	cloned := &BulkUpdateAccountsInput{
		AccountIDs: append([]int64(nil), input.AccountIDs...),
		Extra:      cloneExtraMap(input.Extra),
	}
	if input.ProxyID != nil {
		cloned.ProxyID = int64PtrPool(*input.ProxyID)
	}
	if input.GroupIDs != nil {
		groupIDs := append([]int64(nil), (*input.GroupIDs)...)
		cloned.GroupIDs = &groupIDs
	}
	s.bulkUpdateInputs = append(s.bulkUpdateInputs, cloned)
	if len(input.AccountIDs) > 0 && s.bulkUpdateErrByID != nil {
		if err, ok := s.bulkUpdateErrByID[input.AccountIDs[0]]; ok {
			return nil, err
		}
	}
	return &BulkUpdateAccountsResult{Success: len(input.AccountIDs)}, nil
}

func TestOpenAIFreePoolPreview_NewDefaultAccountFollowsExistingDateAnchor(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, OpenAIFreePoolConfig{
		Enabled:        true,
		DefaultGroupID: 100,
		PlusGroupID:    200,
		LookaheadDays:  14,
		Pools: []OpenAIFreePool{
			{GroupID: 101, ProxyID: 501, Label: "free-1", SortOrder: 1},
			{GroupID: 102, ProxyID: 502, Label: "free-2", SortOrder: 2},
			{GroupID: 103, ProxyID: 503, Label: "free-3", SortOrder: 3},
			{GroupID: 104, ProxyID: 504, Label: "free-4", SortOrder: 4},
			{GroupID: 105, ProxyID: 505, Label: "free-5", SortOrder: 5},
		},
	})
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "default-new", 100, 0, nil, fixedResetAt(2026, time.June, 1)),
				newOpenAIFreePoolAccount(2, "pool-1-a", 101, 501, nil, fixedResetAt(2026, time.June, 2)),
				newOpenAIFreePoolAccount(3, "pool-2-a", 102, 502, nil, fixedResetAt(2026, time.June, 1)),
				newOpenAIFreePoolAccount(4, "pool-2-b", 102, 502, nil, fixedResetAt(2026, time.June, 1)),
				newOpenAIFreePoolAccount(5, "pool-3-a", 103, 503, nil, fixedResetAt(2026, time.June, 3)),
				newOpenAIFreePoolAccount(6, "pool-4-a", 104, 504, nil, fixedResetAt(2026, time.June, 4)),
				newOpenAIFreePoolAccount(7, "pool-5-a", 105, 505, nil, fixedResetAt(2026, time.June, 5)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, preview.Moves, 1)
	require.Equal(t, int64(1), preview.Moves[0].AccountID)
	require.Equal(t, int64(102), preview.Moves[0].TargetGroupID)
	require.Equal(t, int64(502), preview.Moves[0].TargetProxyID)
	require.Equal(t, openAIFreePoolReasonNewFromDefault, preview.Moves[0].Reason)
}

func TestOpenAIFreePoolPreview_SameResetDateDefaultsGoToSamePool(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "default-a", 100, 0, nil, fixedResetAt(2026, time.June, 10)),
				newOpenAIFreePoolAccount(2, "default-b", 100, 0, nil, fixedResetAt(2026, time.June, 10)),
				newOpenAIFreePoolAccount(3, "pool-1-a", 101, 501, nil, fixedResetAt(2026, time.June, 1)),
				newOpenAIFreePoolAccount(4, "pool-2-a", 102, 502, nil, fixedResetAt(2026, time.June, 2)),
				newOpenAIFreePoolAccount(5, "pool-3-a", 103, 503, nil, fixedResetAt(2026, time.June, 3)),
				newOpenAIFreePoolAccount(6, "pool-4-a", 104, 504, nil, fixedResetAt(2026, time.June, 4)),
				newOpenAIFreePoolAccount(7, "pool-5-a", 105, 505, nil, fixedResetAt(2026, time.June, 5)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, preview.Moves, 2)
	require.Equal(t, preview.Moves[0].TargetGroupID, preview.Moves[1].TargetGroupID)
	require.Equal(t, preview.Moves[0].TargetProxyID, preview.Moves[1].TargetProxyID)
}

func TestOpenAIFreePoolPreview_UnknownResetUsesUnknownLightestPool(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "default-unknown", 100, 0, nil, ""),
				newOpenAIFreePoolAccount(2, "pool-1-unknown-a", 101, 501, nil, ""),
				newOpenAIFreePoolAccount(3, "pool-1-unknown-b", 101, 501, nil, ""),
				newOpenAIFreePoolAccount(4, "pool-2-unknown-a", 102, 502, nil, ""),
				newOpenAIFreePoolAccount(5, "pool-3-known", 103, 503, nil, fixedResetAt(2026, time.June, 3)),
				newOpenAIFreePoolAccount(6, "pool-4-known", 104, 504, nil, fixedResetAt(2026, time.June, 4)),
				newOpenAIFreePoolAccount(7, "pool-5-known", 105, 505, nil, fixedResetAt(2026, time.June, 5)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, preview.Moves, 1)
	require.Equal(t, int64(103), preview.Moves[0].TargetGroupID)
	require.Equal(t, int64(503), preview.Moves[0].TargetProxyID)
}

func TestOpenAIFreePoolPreview_StableLockedAccountDoesNotMoveByDefault(t *testing.T) {
	lockExtra := map[string]any{
		"auto_pool_source":   OpenAIFreePoolSource,
		"auto_pool_group_id": int64(101),
		"auto_pool_proxy_id": int64(501),
	}
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "free-peru", 101, 501, lockExtra, resetAtWithOffset(1)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), false)
	require.NoError(t, err)
	require.Empty(t, preview.Moves)
	require.Equal(t, 1, preview.Summary.LockedAccounts)
}

func TestOpenAIFreePoolPreview_ForceRebalanceAllowsHistoryMove(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "existing-free", 101, 501, nil, resetAtWithOffset(2)),
				newOpenAIFreePoolAccount(2, "existing-free-2", 101, 501, nil, resetAtWithOffset(3)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), true)
	require.NoError(t, err)
	require.Len(t, preview.Moves, 2)
	for _, move := range preview.Moves {
		require.Equal(t, openAIFreePoolReasonForcedRebalance, move.Reason)
	}
}

func TestOpenAIFreePoolForecast_ExcludesPlusAndCountsUnknownWithinLookahead(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "known", 101, 501, nil, resetAtWithOffset(1)),
				newOpenAIFreePoolAccount(2, "unknown", 102, 502, map[string]any{}, ""),
				newOpenAIFreePoolAccount(3, "too-far", 103, 503, nil, resetAtWithOffset(30)),
				{
					ID:          4,
					Name:        "plus",
					Platform:    PlatformOpenAI,
					Type:        AccountTypeOAuth,
					Credentials: map[string]any{"plan_type": "free"},
					Extra:       map[string]any{"codex_7d_reset_at": resetAtWithOffset(1)},
					GroupIDs:    []int64{200},
					ProxyID:     int64PtrPool(504),
				},
			},
		},
	)

	forecast, err := svc.Forecast(context.Background())
	require.NoError(t, err)
	require.Len(t, forecast.Days, 0)
	require.Equal(t, 0, forecast.UnknownCount)
}

func TestOpenAIFreePoolForecast_AllowsDefaultOnlyWithoutPools(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, OpenAIFreePoolConfig{
		Enabled:        false,
		DefaultGroupID: 100,
		PlusGroupID:    200,
		LookaheadDays:  14,
		Pools:          []OpenAIFreePool{},
	})
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "default-free", 100, 0, nil, resetAtWithOffset(1)),
				newOpenAIFreePoolAccount(2, "default-free-unknown", 100, 0, nil, ""),
				newOpenAIFreePoolAccount(3, "plus-free", 200, 501, nil, resetAtWithOffset(1)),
				newOpenAIFreePoolAccount(4, "other-group-free", 101, 501, nil, resetAtWithOffset(1)),
			},
		},
	)

	forecast, err := svc.Forecast(context.Background())
	require.NoError(t, err)
	require.Len(t, forecast.Days, 1)
	require.Equal(t, 1, forecast.Days[0].Count)
	require.Equal(t, int64(1), forecast.Days[0].Accounts[0].AccountID)
	require.True(t, forecast.Days[0].Accounts[0].InDefaultGroup)
	require.Equal(t, 1, forecast.UnknownCount)
}

func TestOpenAIFreePoolApply_MergesExistingExtraFields(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	adminStub := &openAIFreePoolAdminServiceStub{
		groups:  buildOpenAIFreePoolGroups(),
		proxies: buildOpenAIFreePoolProxies(),
		accounts: []Account{
			newOpenAIFreePoolAccount(1, "default-new", 100, 0, map[string]any{"existing_flag": "keep"}, resetAtWithOffset(1)),
		},
	}
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		adminStub,
	)

	result, err := svc.Apply(context.Background(), OpenAIFreePoolApplyRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, result.Applied)
	require.Len(t, adminStub.bulkUpdateInputs, 1)
	require.Equal(t, "keep", adminStub.bulkUpdateInputs[0].Extra["existing_flag"])
	require.Equal(t, OpenAIFreePoolSource, adminStub.bulkUpdateInputs[0].Extra["auto_pool_source"])
	require.Equal(t, OpenAIFreePoolLockModeAuto, adminStub.bulkUpdateInputs[0].Extra["auto_pool_lock_mode"])
}

func TestOpenAIFreePoolApply_PreservesManualLockModeWhenApplyingManualTarget(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	adminStub := &openAIFreePoolAdminServiceStub{
		groups:  buildOpenAIFreePoolGroups(),
		proxies: buildOpenAIFreePoolProxies(),
		accounts: []Account{
			newOpenAIFreePoolAccount(1, "default-new", 100, 0, map[string]any{
				"auto_pool_source":    OpenAIFreePoolSource,
				"auto_pool_group_id":  int64(104),
				"auto_pool_proxy_id":  int64(504),
				"auto_pool_lock_mode": OpenAIFreePoolLockModeManual,
			}, resetAtWithOffset(1)),
		},
	}
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		adminStub,
	)

	result, err := svc.Apply(context.Background(), OpenAIFreePoolApplyRequest{})
	require.NoError(t, err)
	require.Equal(t, 1, result.Applied)
	require.Len(t, adminStub.bulkUpdateInputs, 1)
	require.Equal(t, OpenAIFreePoolLockModeManual, adminStub.bulkUpdateInputs[0].Extra["auto_pool_lock_mode"])
}

func TestOpenAIFreePoolPreview_ManualLockKeepsTargetPool(t *testing.T) {
	lockExtra := map[string]any{
		"auto_pool_source":    OpenAIFreePoolSource,
		"auto_pool_group_id":  int64(105),
		"auto_pool_proxy_id":  int64(505),
		"auto_pool_lock_mode": OpenAIFreePoolLockModeManual,
	}
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "manual-locked", 101, 501, lockExtra, resetAtWithOffset(1)),
			},
		},
	)

	preview, err := svc.Preview(context.Background(), false)
	require.NoError(t, err)
	require.Len(t, preview.Moves, 1)
	require.Equal(t, int64(105), preview.Moves[0].TargetGroupID)
	require.Equal(t, int64(505), preview.Moves[0].TargetProxyID)
}

func TestOpenAIFreePoolLockAccount_UsesManualModeAndTargetProxy(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	adminStub := &openAIFreePoolAdminServiceStub{
		groups:  buildOpenAIFreePoolGroups(),
		proxies: buildOpenAIFreePoolProxies(),
		accounts: []Account{
			newOpenAIFreePoolAccount(1, "default-new", 100, 0, nil, resetAtWithOffset(1)),
		},
	}
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		adminStub,
	)

	err := svc.LockAccount(context.Background(), 1, 104)
	require.NoError(t, err)
	require.Len(t, adminStub.bulkUpdateInputs, 1)
	require.NotNil(t, adminStub.bulkUpdateInputs[0].ProxyID)
	require.Equal(t, int64(504), *adminStub.bulkUpdateInputs[0].ProxyID)
	require.NotNil(t, adminStub.bulkUpdateInputs[0].GroupIDs)
	require.Equal(t, []int64{104}, *adminStub.bulkUpdateInputs[0].GroupIDs)
	require.Equal(t, OpenAIFreePoolLockModeManual, adminStub.bulkUpdateInputs[0].Extra["auto_pool_lock_mode"])
}

func TestOpenAIFreePoolUnlockAccount_PreservesCurrentPlacement(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	adminStub := &openAIFreePoolAdminServiceStub{
		groups:  buildOpenAIFreePoolGroups(),
		proxies: buildOpenAIFreePoolProxies(),
		accounts: []Account{
			newOpenAIFreePoolAccount(1, "free-peru", 101, 501, map[string]any{
				"auto_pool_source":    OpenAIFreePoolSource,
				"auto_pool_group_id":  int64(105),
				"auto_pool_proxy_id":  int64(505),
				"auto_pool_lock_mode": OpenAIFreePoolLockModeManual,
			}, resetAtWithOffset(1)),
		},
	}
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		adminStub,
	)

	err := svc.UnlockAccount(context.Background(), 1)
	require.NoError(t, err)
	require.Len(t, adminStub.bulkUpdateInputs, 1)
	require.Nil(t, adminStub.bulkUpdateInputs[0].ProxyID)
	require.Nil(t, adminStub.bulkUpdateInputs[0].GroupIDs)
	require.Equal(t, int64(101), adminStub.bulkUpdateInputs[0].Extra["auto_pool_group_id"])
	require.Equal(t, int64(501), adminStub.bulkUpdateInputs[0].Extra["auto_pool_proxy_id"])
	require.Equal(t, OpenAIFreePoolLockModeAuto, adminStub.bulkUpdateInputs[0].Extra["auto_pool_lock_mode"])
}

func TestOpenAIFreePoolListManagedAccounts_ReturnsLockMetadata(t *testing.T) {
	cfg := mustMarshalOpenAIFreePoolConfig(t, defaultOpenAIFreePoolConfig())
	svc := NewOpenAIFreePoolService(
		&openAIFreePoolSettingRepoStub{values: map[string]string{SettingKeyOpenAIFreePoolConfig: cfg}},
		&openAIFreePoolAdminServiceStub{
			groups:  buildOpenAIFreePoolGroups(),
			proxies: buildOpenAIFreePoolProxies(),
			accounts: []Account{
				newOpenAIFreePoolAccount(1, "default-new", 100, 0, nil, resetAtWithOffset(1)),
				newOpenAIFreePoolAccount(2, "manual-locked", 101, 501, map[string]any{
					"auto_pool_source":    OpenAIFreePoolSource,
					"auto_pool_group_id":  int64(101),
					"auto_pool_proxy_id":  int64(501),
					"auto_pool_lock_mode": OpenAIFreePoolLockModeManual,
				}, resetAtWithOffset(2)),
			},
		},
	)

	items, err := svc.ListManagedAccounts(context.Background())
	require.NoError(t, err)
	require.Len(t, items, 2)
	require.True(t, items[0].InDefaultGroup)
	require.Equal(t, OpenAIFreePoolLockModeUnlocked, items[0].LockMode)
	require.Equal(t, OpenAIFreePoolLockModeManual, items[1].LockMode)
}

func TestValidateOpenAIFreePoolConfig_RejectsDefaultPlusReuse(t *testing.T) {
	err := validateOpenAIFreePoolConfig(&OpenAIFreePoolConfig{
		Enabled:        true,
		DefaultGroupID: 100,
		PlusGroupID:    100,
		LookaheadDays:  14,
		Pools: []OpenAIFreePool{
			{GroupID: 101, ProxyID: 501},
			{GroupID: 102, ProxyID: 502},
			{GroupID: 103, ProxyID: 503},
			{GroupID: 104, ProxyID: 504},
			{GroupID: 105, ProxyID: 505},
		},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be different")
}

func defaultOpenAIFreePoolConfig() OpenAIFreePoolConfig {
	return OpenAIFreePoolConfig{
		Enabled:        true,
		DefaultGroupID: 100,
		PlusGroupID:    200,
		LookaheadDays:  14,
		Pools: []OpenAIFreePool{
			{GroupID: 101, ProxyID: 501, Label: "free-1", SortOrder: 1},
			{GroupID: 102, ProxyID: 502, Label: "free-2", SortOrder: 2},
			{GroupID: 103, ProxyID: 503, Label: "free-3", SortOrder: 3},
			{GroupID: 104, ProxyID: 504, Label: "free-4", SortOrder: 4},
			{GroupID: 105, ProxyID: 505, Label: "free-5", SortOrder: 5},
		},
	}
}

func buildOpenAIFreePoolGroups() []Group {
	return []Group{
		{ID: 100, Name: "default"},
		{ID: 101, Name: "free-1"},
		{ID: 102, Name: "free-2"},
		{ID: 103, Name: "free-3"},
		{ID: 104, Name: "free-4"},
		{ID: 105, Name: "free-5"},
		{ID: 200, Name: "plus"},
	}
}

func buildOpenAIFreePoolProxies() []Proxy {
	return []Proxy{
		{ID: 501, Name: "proxy-1"},
		{ID: 502, Name: "proxy-2"},
		{ID: 503, Name: "proxy-3"},
		{ID: 504, Name: "proxy-4"},
		{ID: 505, Name: "proxy-5"},
	}
}

func newOpenAIFreePoolAccount(id int64, name string, groupID int64, proxyID int64, extra map[string]any, resetAt string) Account {
	clone := cloneExtraMap(extra)
	if resetAt != "" {
		clone["codex_7d_reset_at"] = resetAt
	}
	account := Account{
		ID:          id,
		Name:        name,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"plan_type": "free"},
		Extra:       clone,
		GroupIDs:    []int64{groupID},
	}
	if proxyID > 0 {
		account.ProxyID = int64PtrPool(proxyID)
	}
	return account
}

func resetAtWithOffset(days int) string {
	return time.Now().UTC().AddDate(0, 0, days).Format(time.RFC3339)
}

func fixedResetAt(year int, month time.Month, day int) string {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
}

func mustMarshalOpenAIFreePoolConfig(t *testing.T, cfg OpenAIFreePoolConfig) string {
	t.Helper()
	payload, err := json.Marshal(cfg)
	require.NoError(t, err)
	return string(payload)
}
