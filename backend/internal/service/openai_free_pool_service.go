package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	OpenAIFreePoolSource                = "openai_free_pool_v1"
	OpenAIFreePoolLockModeAuto          = "auto"
	OpenAIFreePoolLockModeManual        = "manual"
	OpenAIFreePoolLockModeUnlocked      = "unlocked"
	openAIFreePoolReasonNewFromDefault  = "new_from_default"
	openAIFreePoolReasonInvalidLock     = "invalid_lock"
	openAIFreePoolReasonInvalidMapping  = "invalid_mapping"
	openAIFreePoolReasonForcedRebalance = "forced_rebalance"
	openAIFreePoolUnknownResetDate      = "unknown"
	openAIFreePoolDefaultLookaheadDays  = 14
)

type OpenAIFreePool struct {
	GroupID   int64  `json:"group_id"`
	ProxyID   int64  `json:"proxy_id"`
	Label     string `json:"label"`
	SortOrder int    `json:"sort_order"`
}

type OpenAIFreePoolConfig struct {
	Enabled        bool             `json:"enabled"`
	DefaultGroupID int64            `json:"default_group_id"`
	PlusGroupID    int64            `json:"plus_group_id"`
	LookaheadDays  int              `json:"lookahead_days"`
	Pools          []OpenAIFreePool `json:"pools"`
}

type OpenAIFreePoolSummary struct {
	ManagedAccounts      int `json:"managed_accounts"`
	DefaultAccounts      int `json:"default_accounts"`
	LockedAccounts       int `json:"locked_accounts"`
	UnknownResetAccounts int `json:"unknown_reset_accounts"`
	PendingMoves         int `json:"pending_moves"`
}

type OpenAIFreePoolState struct {
	GroupID              int64  `json:"group_id"`
	GroupName            string `json:"group_name"`
	ProxyID              int64  `json:"proxy_id"`
	ProxyName            string `json:"proxy_name"`
	Accounts             int    `json:"accounts"`
	LockedAccounts       int    `json:"locked_accounts"`
	UnknownResetAccounts int    `json:"unknown_reset_accounts"`
}

type OpenAIFreePoolMove struct {
	AccountID        int64   `json:"account_id"`
	AccountName      string  `json:"account_name"`
	CurrentGroupID   *int64  `json:"current_group_id,omitempty"`
	CurrentGroupName string  `json:"current_group_name,omitempty"`
	TargetGroupID    int64   `json:"target_group_id"`
	TargetGroupName  string  `json:"target_group_name"`
	CurrentProxyID   *int64  `json:"current_proxy_id,omitempty"`
	CurrentProxyName string  `json:"current_proxy_name,omitempty"`
	TargetProxyID    int64   `json:"target_proxy_id"`
	TargetProxyName  string  `json:"target_proxy_name"`
	ResetAt          *string `json:"reset_at,omitempty"`
	ResetDate        string  `json:"reset_date"`
	Locked           bool    `json:"locked"`
	Reason           string  `json:"reason"`
}

type OpenAIFreePoolPreview struct {
	Config             OpenAIFreePoolConfig  `json:"config"`
	Summary            OpenAIFreePoolSummary `json:"summary"`
	Pools              []OpenAIFreePoolState `json:"pools"`
	Moves              []OpenAIFreePoolMove  `json:"moves"`
	UnknownResetIDs    []int64               `json:"unknown_reset_ids"`
	GeneratedAt        string                `json:"generated_at"`
	ForceRebalance     bool                  `json:"force_rebalance"`
}

type OpenAIFreePoolApplyRequest struct {
	ForceRebalance bool `json:"force_rebalance"`
}

type OpenAIFreePoolApplyResult struct {
	Applied    int                    `json:"applied"`
	Skipped    int                    `json:"skipped"`
	Failed     int                    `json:"failed"`
	Preview    *OpenAIFreePoolPreview `json:"preview,omitempty"`
	GeneratedAt string                `json:"generated_at"`
}

type OpenAIFreePoolManagedAccount struct {
	AccountID        int64   `json:"account_id"`
	AccountName      string  `json:"account_name"`
	CurrentGroupID   *int64  `json:"current_group_id,omitempty"`
	CurrentGroupName string  `json:"current_group_name,omitempty"`
	CurrentProxyID   *int64  `json:"current_proxy_id,omitempty"`
	CurrentProxyName string  `json:"current_proxy_name,omitempty"`
	ResetAt          *string `json:"reset_at,omitempty"`
	InDefaultGroup   bool    `json:"in_default_group"`
	LockMode         string  `json:"lock_mode"`
	LockGroupID      *int64  `json:"lock_group_id,omitempty"`
	LockGroupName    string  `json:"lock_group_name,omitempty"`
	LockProxyID      *int64  `json:"lock_proxy_id,omitempty"`
	LockProxyName    string  `json:"lock_proxy_name,omitempty"`
}

type OpenAIFreePoolLockRequest struct {
	AccountID     int64 `json:"account_id"`
	TargetGroupID int64 `json:"target_group_id"`
}

type OpenAIFreeResetForecastAccount struct {
	AccountID      int64   `json:"account_id"`
	AccountName    string  `json:"account_name"`
	GroupID        *int64  `json:"group_id,omitempty"`
	GroupName      string  `json:"group_name,omitempty"`
	ProxyID        *int64  `json:"proxy_id,omitempty"`
	ProxyName      string  `json:"proxy_name,omitempty"`
	InDefaultGroup bool    `json:"in_default_group"`
	UsagePercent   *float64 `json:"usage_percent,omitempty"`
	ResetAt        *string `json:"reset_at,omitempty"`
}

type OpenAIFreeResetForecastDay struct {
	Date     string                            `json:"date"`
	Count    int                               `json:"count"`
	Accounts []OpenAIFreeResetForecastAccount  `json:"accounts"`
}

type OpenAIFreeResetForecast struct {
	Days         []OpenAIFreeResetForecastDay `json:"days"`
	UnknownCount int                          `json:"unknown_count"`
	GeneratedAt  string                       `json:"generated_at"`
}

type openAIFreePoolAdminService interface {
	GetAllGroups(ctx context.Context) ([]Group, error)
	GetAllProxies(ctx context.Context) ([]Proxy, error)
	ListAccounts(ctx context.Context, page, pageSize int, platform, accountType, status, search string, groupID int64, privacyMode string, sortBy, sortOrder string) ([]Account, int64, error)
	BulkUpdateAccounts(ctx context.Context, input *BulkUpdateAccountsInput) (*BulkUpdateAccountsResult, error)
}

type OpenAIFreePoolService struct {
	settingRepo  SettingRepository
	adminService openAIFreePoolAdminService
}

func NewOpenAIFreePoolService(settingRepo SettingRepository, adminService openAIFreePoolAdminService) *OpenAIFreePoolService {
	return &OpenAIFreePoolService{
		settingRepo: settingRepo,
		adminService: adminService,
	}
}

func (s *OpenAIFreePoolService) GetConfig(ctx context.Context) (*OpenAIFreePoolConfig, error) {
	if s == nil || s.settingRepo == nil {
		return &OpenAIFreePoolConfig{LookaheadDays: openAIFreePoolDefaultLookaheadDays}, nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyOpenAIFreePoolConfig)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return &OpenAIFreePoolConfig{LookaheadDays: openAIFreePoolDefaultLookaheadDays}, nil
		}
		return nil, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return &OpenAIFreePoolConfig{LookaheadDays: openAIFreePoolDefaultLookaheadDays}, nil
	}
	var cfg OpenAIFreePoolConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("parse openai free pool config: %w", err)
	}
	normalizeOpenAIFreePoolConfig(&cfg)
	return &cfg, nil
}

func (s *OpenAIFreePoolService) UpdateConfig(ctx context.Context, cfg *OpenAIFreePoolConfig) (*OpenAIFreePoolConfig, error) {
	if cfg == nil {
		return nil, fmt.Errorf("openai free pool config is required")
	}
	normalized := *cfg
	normalizeOpenAIFreePoolConfig(&normalized)
	if err := validateOpenAIFreePoolConfig(&normalized); err != nil {
		return nil, err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return nil, fmt.Errorf("marshal openai free pool config: %w", err)
	}
	if err := s.settingRepo.Set(ctx, SettingKeyOpenAIFreePoolConfig, string(payload)); err != nil {
		return nil, err
	}
	return &normalized, nil
}

func (s *OpenAIFreePoolService) Preview(ctx context.Context, forceRebalance bool) (*OpenAIFreePoolPreview, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateOpenAIFreePoolConfig(cfg); err != nil {
		return nil, err
	}
	groups, err := s.adminService.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}
	proxies, err := s.adminService.GetAllProxies(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateOpenAIFreePoolRuntimeConfig(cfg, groups, proxies); err != nil {
		return nil, err
	}
	accounts, err := s.collectManagedAccounts(ctx, cfg)
	if err != nil {
		return nil, err
	}
	preview := planOpenAIFreePoolPreview(cfg, groups, proxies, accounts, forceRebalance)
	return preview, nil
}

func (s *OpenAIFreePoolService) Apply(ctx context.Context, req OpenAIFreePoolApplyRequest) (*OpenAIFreePoolApplyResult, error) {
	preview, err := s.Preview(ctx, req.ForceRebalance)
	if err != nil {
		return nil, err
	}
	result := &OpenAIFreePoolApplyResult{
		Preview:     preview,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}
	if len(preview.Moves) == 0 {
		result.Skipped = preview.Summary.ManagedAccounts
		return result, nil
	}
	accounts, err := s.collectManagedAccounts(ctx, &preview.Config)
	if err != nil {
		return nil, err
	}
	accountByID := make(map[int64]*Account, len(accounts))
	for _, account := range accounts {
		if account != nil {
			accountByID[account.ID] = account
		}
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for _, move := range preview.Moves {
		account := accountByID[move.AccountID]
		if account == nil {
			result.Skipped++
			continue
		}
		extra := cloneExtraMap(account.Extra)
		lockMode := OpenAIFreePoolLockModeAuto
		lockState := extractOpenAIFreeLockState(account)
		if !req.ForceRebalance &&
			lockState.mode == OpenAIFreePoolLockModeManual &&
			lockState.groupID != nil &&
			lockState.proxyID != nil &&
			*lockState.groupID == move.TargetGroupID &&
			*lockState.proxyID == move.TargetProxyID {
			lockMode = OpenAIFreePoolLockModeManual
		}
		extra["auto_pool_source"] = OpenAIFreePoolSource
		extra["auto_pool_group_id"] = move.TargetGroupID
		extra["auto_pool_proxy_id"] = move.TargetProxyID
		extra["auto_pool_lock_mode"] = lockMode
		extra["auto_pool_last_reason"] = move.Reason
		extra["auto_pool_last_planned_at"] = now
		extra["auto_pool_locked_at"] = now
		updates := &BulkUpdateAccountsInput{
			AccountIDs: []int64{move.AccountID},
			ProxyID:    int64PtrPool(move.TargetProxyID),
			GroupIDs:   &[]int64{move.TargetGroupID},
			Extra:      extra,
		}
		if _, err := s.adminService.BulkUpdateAccounts(ctx, updates); err != nil {
			result.Failed++
			continue
		}
		result.Applied++
	}
	return result, nil
}

func (s *OpenAIFreePoolService) Forecast(ctx context.Context) (*OpenAIFreeResetForecast, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, err
	}
	groups, err := s.adminService.GetAllGroups(ctx)
	if err != nil {
		return nil, err
	}
	proxies, err := s.adminService.GetAllProxies(ctx)
	if err != nil {
		return nil, err
	}
	if err := validateOpenAIFreeResetForecastRuntimeConfig(cfg, groups); err != nil {
		return nil, err
	}
	accounts, err := s.collectForecastAccounts(ctx, cfg)
	if err != nil {
		return nil, err
	}
	return buildOpenAIFreeResetForecast(cfg, groups, proxies, accounts), nil
}

func (s *OpenAIFreePoolService) ListManagedAccounts(ctx context.Context) ([]OpenAIFreePoolManagedAccount, error) {
	cfg, groups, proxies, accounts, err := s.loadRuntime(ctx)
	if err != nil {
		return nil, err
	}
	groupByID := make(map[int64]Group, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = group
	}
	proxyByID := make(map[int64]Proxy, len(proxies))
	for _, proxy := range proxies {
		proxyByID[proxy.ID] = proxy
	}
	poolIndexByGroupID := managedPoolIndex(cfg)

	items := make([]OpenAIFreePoolManagedAccount, 0, len(accounts))
	for _, account := range accounts {
		if account == nil {
			continue
		}
		currentGroupID, currentGroupName := detectManagedOpenAIFreeGroup(account, cfg, groupByID, poolIndexByGroupID)
		currentProxyID, currentProxyName := detectProxy(account, proxyByID)
		resetAt, _ := extractOpenAIFreeReset(account)
		lockState := extractOpenAIFreeLockState(account)
		lockGroupName := ""
		if lockState.groupID != nil {
			lockGroupName = groupNameByID(groupByID, *lockState.groupID)
		}
		lockProxyName := ""
		if lockState.proxyID != nil {
			lockProxyName = proxyNameByID(proxyByID, *lockState.proxyID)
		}
		items = append(items, OpenAIFreePoolManagedAccount{
			AccountID:        account.ID,
			AccountName:      account.Name,
			CurrentGroupID:   currentGroupID,
			CurrentGroupName: currentGroupName,
			CurrentProxyID:   currentProxyID,
			CurrentProxyName: currentProxyName,
			ResetAt:          resetAt,
			InDefaultGroup:   currentGroupID != nil && *currentGroupID == cfg.DefaultGroupID,
			LockMode:         lockState.mode,
			LockGroupID:      lockState.groupID,
			LockGroupName:    lockGroupName,
			LockProxyID:      lockState.proxyID,
			LockProxyName:    lockProxyName,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		left, right := items[i], items[j]
		if left.InDefaultGroup != right.InDefaultGroup {
			return left.InDefaultGroup
		}
		if left.CurrentGroupName != right.CurrentGroupName {
			return left.CurrentGroupName < right.CurrentGroupName
		}
		return left.AccountName < right.AccountName
	})
	return items, nil
}

func (s *OpenAIFreePoolService) LockAccount(ctx context.Context, accountID, targetGroupID int64) error {
	cfg, _, _, accounts, err := s.loadRuntime(ctx)
	if err != nil {
		return err
	}
	poolIndexByGroupID := managedPoolIndex(cfg)
	poolIdx, ok := poolIndexByGroupID[targetGroupID]
	if !ok {
		return fmt.Errorf("target group %d is not a configured free pool", targetGroupID)
	}
	account := findManagedOpenAIFreeAccount(accounts, accountID)
	if account == nil {
		return fmt.Errorf("managed openai free account %d not found", accountID)
	}
	targetPool := cfg.Pools[poolIdx]
	now := time.Now().UTC().Format(time.RFC3339)
	extra := cloneExtraMap(account.Extra)
	extra["auto_pool_source"] = OpenAIFreePoolSource
	extra["auto_pool_group_id"] = targetPool.GroupID
	extra["auto_pool_proxy_id"] = targetPool.ProxyID
	extra["auto_pool_lock_mode"] = OpenAIFreePoolLockModeManual
	extra["auto_pool_locked_at"] = now
	extra["auto_pool_last_planned_at"] = now
	extra["auto_pool_last_reason"] = openAIFreePoolReasonInvalidLock
	updates := &BulkUpdateAccountsInput{
		AccountIDs: []int64{account.ID},
		ProxyID:    int64PtrPool(targetPool.ProxyID),
		GroupIDs:   &[]int64{targetPool.GroupID},
		Extra:      extra,
	}
	_, err = s.adminService.BulkUpdateAccounts(ctx, updates)
	return err
}

func (s *OpenAIFreePoolService) UnlockAccount(ctx context.Context, accountID int64) error {
	cfg, groups, proxies, accounts, err := s.loadRuntime(ctx)
	if err != nil {
		return err
	}
	account := findManagedOpenAIFreeAccount(accounts, accountID)
	if account == nil {
		return fmt.Errorf("managed openai free account %d not found", accountID)
	}
	groupByID := make(map[int64]Group, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = group
	}
	proxyByID := make(map[int64]Proxy, len(proxies))
	for _, proxy := range proxies {
		proxyByID[proxy.ID] = proxy
	}
	poolIndexByGroupID := managedPoolIndex(cfg)
	currentGroupID, _ := detectManagedOpenAIFreeGroup(account, cfg, groupByID, poolIndexByGroupID)
	currentProxyID, _ := detectProxy(account, proxyByID)
	now := time.Now().UTC().Format(time.RFC3339)
	extra := cloneExtraMap(account.Extra)
	extra["auto_pool_source"] = OpenAIFreePoolSource
	if currentGroupID != nil {
		extra["auto_pool_group_id"] = *currentGroupID
	}
	if currentProxyID != nil {
		extra["auto_pool_proxy_id"] = *currentProxyID
	}
	lockMode := OpenAIFreePoolLockModeUnlocked
	if currentGroupID != nil {
		if idx, ok := poolIndexByGroupID[*currentGroupID]; ok && currentProxyID != nil && *currentProxyID == cfg.Pools[idx].ProxyID {
			lockMode = OpenAIFreePoolLockModeAuto
			extra["auto_pool_group_id"] = cfg.Pools[idx].GroupID
			extra["auto_pool_proxy_id"] = cfg.Pools[idx].ProxyID
		}
	}
	extra["auto_pool_lock_mode"] = lockMode
	extra["auto_pool_last_planned_at"] = now
	updates := &BulkUpdateAccountsInput{
		AccountIDs: []int64{account.ID},
		Extra:      extra,
	}
	_, err = s.adminService.BulkUpdateAccounts(ctx, updates)
	return err
}

func (s *OpenAIFreePoolService) loadRuntime(ctx context.Context) (*OpenAIFreePoolConfig, []Group, []Proxy, []*Account, error) {
	cfg, err := s.GetConfig(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := validateOpenAIFreePoolConfig(cfg); err != nil {
		return nil, nil, nil, nil, err
	}
	groups, err := s.adminService.GetAllGroups(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	proxies, err := s.adminService.GetAllProxies(ctx)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	if err := validateOpenAIFreePoolRuntimeConfig(cfg, groups, proxies); err != nil {
		return nil, nil, nil, nil, err
	}
	accounts, err := s.collectManagedAccounts(ctx, cfg)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return cfg, groups, proxies, accounts, nil
}

func (s *OpenAIFreePoolService) collectManagedAccounts(ctx context.Context, cfg *OpenAIFreePoolConfig) ([]*Account, error) {
	freeGroupIDs := make(map[int64]struct{}, len(cfg.Pools))
	for _, pool := range cfg.Pools {
		freeGroupIDs[pool.GroupID] = struct{}{}
	}
	accounts, _, err := s.adminService.ListAccounts(ctx, 1, 10000, PlatformOpenAI, "", "", "", 0, "", "name", "asc")
	if err != nil {
		return nil, err
	}
	result := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if !isManagedOpenAIFreeAccount(&account, cfg, freeGroupIDs) {
			continue
		}
		accCopy := account
		result = append(result, &accCopy)
	}
	return result, nil
}

func (s *OpenAIFreePoolService) collectForecastAccounts(ctx context.Context, cfg *OpenAIFreePoolConfig) ([]*Account, error) {
	accounts, _, err := s.adminService.ListAccounts(ctx, 1, 10000, PlatformOpenAI, "", "", "", 0, "", "name", "asc")
	if err != nil {
		return nil, err
	}
	result := make([]*Account, 0, len(accounts))
	for i := range accounts {
		account := accounts[i]
		if !isForecastOpenAIFreeAccount(&account, cfg) {
			continue
		}
		accCopy := account
		result = append(result, &accCopy)
	}
	return result, nil
}

func isManagedOpenAIFreeAccount(account *Account, cfg *OpenAIFreePoolConfig, freeGroupIDs map[int64]struct{}) bool {
	if account == nil || account.Platform != PlatformOpenAI {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(account.GetCredential("plan_type")), "free") {
		return false
	}
	groupIDs := account.GroupIDs
	if len(groupIDs) == 0 && len(account.Groups) > 0 {
		for _, group := range account.Groups {
			if group != nil {
				groupIDs = append(groupIDs, group.ID)
			}
		}
	}
	hasDefault := false
	hasManagedPool := false
	for _, groupID := range groupIDs {
		if groupID == cfg.PlusGroupID {
			return false
		}
		if groupID == cfg.DefaultGroupID {
			hasDefault = true
		}
		if _, ok := freeGroupIDs[groupID]; ok {
			hasManagedPool = true
		}
	}
	return hasDefault || hasManagedPool
}

func isForecastOpenAIFreeAccount(account *Account, cfg *OpenAIFreePoolConfig) bool {
	if account == nil || account.Platform != PlatformOpenAI {
		return false
	}
	if !strings.EqualFold(strings.TrimSpace(account.GetCredential("plan_type")), "free") {
		return false
	}
	groupIDs := account.GroupIDs
	if len(groupIDs) == 0 && len(account.Groups) > 0 {
		for _, group := range account.Groups {
			if group != nil {
				groupIDs = append(groupIDs, group.ID)
			}
		}
	}
	for _, groupID := range groupIDs {
		if groupID == cfg.DefaultGroupID {
			return true
		}
	}
	return false
}

func normalizeOpenAIFreePoolConfig(cfg *OpenAIFreePoolConfig) {
	if cfg == nil {
		return
	}
	if cfg.LookaheadDays <= 0 {
		cfg.LookaheadDays = openAIFreePoolDefaultLookaheadDays
	}
	sort.Slice(cfg.Pools, func(i, j int) bool {
		if cfg.Pools[i].SortOrder == cfg.Pools[j].SortOrder {
			return cfg.Pools[i].GroupID < cfg.Pools[j].GroupID
		}
		return cfg.Pools[i].SortOrder < cfg.Pools[j].SortOrder
	})
}

func validateOpenAIFreePoolConfig(cfg *OpenAIFreePoolConfig) error {
	if cfg == nil {
		return fmt.Errorf("openai free pool config is required")
	}
	if cfg.DefaultGroupID <= 0 {
		return fmt.Errorf("default_group_id is required")
	}
	if cfg.PlusGroupID <= 0 {
		return fmt.Errorf("plus_group_id is required")
	}
	if cfg.DefaultGroupID == cfg.PlusGroupID {
		return fmt.Errorf("default_group_id and plus_group_id must be different")
	}
	if len(cfg.Pools) != 5 {
		return fmt.Errorf("pools must contain exactly 5 bindings")
	}
	groupSeen := make(map[int64]struct{}, len(cfg.Pools))
	proxySeen := make(map[int64]struct{}, len(cfg.Pools))
	for _, pool := range cfg.Pools {
		if pool.GroupID <= 0 || pool.ProxyID <= 0 {
			return fmt.Errorf("pool group_id and proxy_id are required")
		}
		if pool.GroupID == cfg.DefaultGroupID || pool.GroupID == cfg.PlusGroupID {
			return fmt.Errorf("pool group_id %d cannot reuse default or plus group", pool.GroupID)
		}
		if _, ok := groupSeen[pool.GroupID]; ok {
			return fmt.Errorf("duplicate pool group_id: %d", pool.GroupID)
		}
		if _, ok := proxySeen[pool.ProxyID]; ok {
			return fmt.Errorf("duplicate pool proxy_id: %d", pool.ProxyID)
		}
		groupSeen[pool.GroupID] = struct{}{}
		proxySeen[pool.ProxyID] = struct{}{}
	}
	return nil
}

func validateOpenAIFreePoolRuntimeConfig(cfg *OpenAIFreePoolConfig, groups []Group, proxies []Proxy) error {
	groupByID := make(map[int64]struct{}, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = struct{}{}
	}
	if _, ok := groupByID[cfg.DefaultGroupID]; !ok {
		return fmt.Errorf("default group %d does not exist", cfg.DefaultGroupID)
	}
	if _, ok := groupByID[cfg.PlusGroupID]; !ok {
		return fmt.Errorf("plus group %d does not exist", cfg.PlusGroupID)
	}
	for _, pool := range cfg.Pools {
		if _, ok := groupByID[pool.GroupID]; !ok {
			return fmt.Errorf("pool group %d does not exist", pool.GroupID)
		}
	}
	proxyByID := make(map[int64]struct{}, len(proxies))
	for _, proxy := range proxies {
		proxyByID[proxy.ID] = struct{}{}
	}
	for _, pool := range cfg.Pools {
		if _, ok := proxyByID[pool.ProxyID]; !ok {
			return fmt.Errorf("pool proxy %d does not exist", pool.ProxyID)
		}
	}
	return nil
}

func validateOpenAIFreeResetForecastRuntimeConfig(cfg *OpenAIFreePoolConfig, groups []Group) error {
	if cfg == nil {
		return fmt.Errorf("openai free pool config is required")
	}
	if cfg.DefaultGroupID <= 0 {
		return fmt.Errorf("default_group_id is required")
	}
	groupByID := make(map[int64]struct{}, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = struct{}{}
	}
	if _, ok := groupByID[cfg.DefaultGroupID]; !ok {
		return fmt.Errorf("default group %d does not exist", cfg.DefaultGroupID)
	}
	if cfg.LookaheadDays <= 0 {
		cfg.LookaheadDays = openAIFreePoolDefaultLookaheadDays
	}
	return nil
}

type openAIFreePoolPlanCandidate struct {
	account            *Account
	currentGroupID     *int64
	currentGroupName   string
	currentProxyID     *int64
	currentProxyName   string
	resetAt            *string
	resetDate          string
	lockedGroupID      *int64
	lockedProxyID      *int64
	locked             bool
	lockMode           string
	needsMove          bool
	reason             string
	targetGroupID      *int64
	targetProxyID      *int64
}

type openAIFreePoolDateAnchor struct {
	poolIdx int
	locked  bool
}

func planOpenAIFreePoolPreview(
	cfg *OpenAIFreePoolConfig,
	groups []Group,
	proxies []Proxy,
	accounts []*Account,
	forceRebalance bool,
) *OpenAIFreePoolPreview {
	groupByID := make(map[int64]Group, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = group
	}
	proxyByID := make(map[int64]Proxy, len(proxies))
	for _, proxy := range proxies {
		proxyByID[proxy.ID] = proxy
	}

	poolStates := make([]OpenAIFreePoolState, 0, len(cfg.Pools))
	poolIndexByGroupID := make(map[int64]int, len(cfg.Pools))
	for i, pool := range cfg.Pools {
		poolIndexByGroupID[pool.GroupID] = i
		groupName := pool.Label
		if groupName == "" {
			if group, ok := groupByID[pool.GroupID]; ok {
				groupName = group.Name
			}
		}
		proxyName := ""
		if proxy, ok := proxyByID[pool.ProxyID]; ok {
			proxyName = proxy.Name
		}
		poolStates = append(poolStates, OpenAIFreePoolState{
			GroupID:   pool.GroupID,
			GroupName: groupName,
			ProxyID:   pool.ProxyID,
			ProxyName: proxyName,
		})
	}

	candidates := make([]*openAIFreePoolPlanCandidate, 0, len(accounts))
	unknownResetIDs := make([]int64, 0)
	lockedCount := 0
	defaultCount := 0
	managedCount := 0
	plannedAccountsByPool := make([]int, len(poolStates))

	for _, account := range accounts {
		candidate := buildOpenAIFreePoolCandidate(cfg, groupByID, proxyByID, poolIndexByGroupID, account, forceRebalance)
		if candidate == nil {
			continue
		}
		managedCount++
		if candidate.locked {
			lockedCount++
		}
		if candidate.resetDate == openAIFreePoolUnknownResetDate {
			unknownResetIDs = append(unknownResetIDs, account.ID)
		}
		if candidate.currentGroupID != nil && *candidate.currentGroupID == cfg.DefaultGroupID {
			defaultCount++
		}
		if candidate.currentGroupID != nil {
			if idx, ok := poolIndexByGroupID[*candidate.currentGroupID]; ok {
				poolStates[idx].Accounts++
				if candidate.locked && !candidate.needsMove {
					poolStates[idx].LockedAccounts++
				}
				if candidate.resetDate == openAIFreePoolUnknownResetDate {
					poolStates[idx].UnknownResetAccounts++
				}
				if !candidate.needsMove {
					plannedAccountsByPool[idx]++
				}
			}
		}
		candidates = append(candidates, candidate)
	}

	dateCountByPool := make([]map[string]int, len(poolStates))
	for i := range dateCountByPool {
		dateCountByPool[i] = map[string]int{}
	}
	for _, candidate := range candidates {
		if candidate.currentGroupID == nil || candidate.needsMove {
			continue
		}
		idx, ok := poolIndexByGroupID[*candidate.currentGroupID]
		if !ok {
			continue
		}
		dateCountByPool[idx][candidate.resetDate]++
	}
	dateAnchors := buildOpenAIFreePoolDateAnchors(cfg.Pools, dateCountByPool)

	moves := make([]OpenAIFreePoolMove, 0)
	sort.SliceStable(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		if left.needsMove != right.needsMove {
			return left.needsMove
		}
		if left.resetDate != right.resetDate {
			return left.resetDate < right.resetDate
		}
		return left.account.ID < right.account.ID
	})

	for _, candidate := range candidates {
		if !candidate.needsMove {
			continue
		}
		targetIdx := -1
		if candidate.targetGroupID != nil {
			if idx, ok := poolIndexByGroupID[*candidate.targetGroupID]; ok {
				targetIdx = idx
			}
		}
		if targetIdx < 0 {
			targetIdx = chooseOpenAIFreePoolTarget(
				cfg.Pools,
				plannedAccountsByPool,
				dateCountByPool,
				poolStates,
				dateAnchors,
				candidate.resetDate,
			)
		}
		targetPool := poolStates[targetIdx]
		plannedAccountsByPool[targetIdx]++
		dateCountByPool[targetIdx][candidate.resetDate]++
		if candidate.resetDate != openAIFreePoolUnknownResetDate {
			if _, exists := dateAnchors[candidate.resetDate]; !exists {
				dateAnchors[candidate.resetDate] = openAIFreePoolDateAnchor{
					poolIdx: targetIdx,
				}
			}
		} else {
			poolStates[targetIdx].UnknownResetAccounts++
		}
		moves = append(moves, OpenAIFreePoolMove{
			AccountID:        candidate.account.ID,
			AccountName:      candidate.account.Name,
			CurrentGroupID:   candidate.currentGroupID,
			CurrentGroupName: candidate.currentGroupName,
			TargetGroupID:    targetPool.GroupID,
			TargetGroupName:  targetPool.GroupName,
			CurrentProxyID:   candidate.currentProxyID,
			CurrentProxyName: candidate.currentProxyName,
			TargetProxyID:    targetPool.ProxyID,
			TargetProxyName:  targetPool.ProxyName,
			ResetAt:          candidate.resetAt,
			ResetDate:        candidate.resetDate,
			Locked:           candidate.locked,
			Reason:           candidate.reason,
		})
	}

	return &OpenAIFreePoolPreview{
		Config: cfgClone(cfg),
		Summary: OpenAIFreePoolSummary{
			ManagedAccounts:      managedCount,
			DefaultAccounts:      defaultCount,
			LockedAccounts:       lockedCount,
			UnknownResetAccounts: len(unknownResetIDs),
			PendingMoves:         len(moves),
		},
		Pools:           poolStates,
		Moves:           moves,
		UnknownResetIDs: unknownResetIDs,
		GeneratedAt:     time.Now().UTC().Format(time.RFC3339),
		ForceRebalance:  forceRebalance,
	}
}

func buildOpenAIFreePoolCandidate(
	cfg *OpenAIFreePoolConfig,
	groupByID map[int64]Group,
	proxyByID map[int64]Proxy,
	poolIndexByGroupID map[int64]int,
	account *Account,
	forceRebalance bool,
) *openAIFreePoolPlanCandidate {
	if account == nil {
		return nil
	}
	currentGroupID, currentGroupName := detectManagedOpenAIFreeGroup(account, cfg, groupByID, poolIndexByGroupID)
	currentProxyID, currentProxyName := detectProxy(account, proxyByID)
	resetAt, resetDate := extractOpenAIFreeReset(account)
	lockState := extractOpenAIFreeLockState(account)

	candidate := &openAIFreePoolPlanCandidate{
		account:          account,
		currentGroupID:   currentGroupID,
		currentGroupName: currentGroupName,
		currentProxyID:   currentProxyID,
		currentProxyName: currentProxyName,
		resetAt:          resetAt,
		resetDate:        resetDate,
		lockedGroupID:    lockState.groupID,
		lockedProxyID:    lockState.proxyID,
		locked:           lockState.locked,
		lockMode:         lockState.mode,
	}

	if currentGroupID != nil && *currentGroupID == cfg.DefaultGroupID {
		if lockState.mode == OpenAIFreePoolLockModeManual && lockState.groupID != nil && lockState.proxyID != nil {
			candidate.targetGroupID = int64PtrPool(*lockState.groupID)
			candidate.targetProxyID = int64PtrPool(*lockState.proxyID)
			candidate.reason = openAIFreePoolReasonInvalidLock
		}
		candidate.needsMove = true
		if candidate.reason == "" {
			candidate.reason = openAIFreePoolReasonNewFromDefault
		}
		return candidate
	}

	if forceRebalance {
		candidate.needsMove = true
		candidate.reason = openAIFreePoolReasonForcedRebalance
		return candidate
	}

	if lockState.mode == OpenAIFreePoolLockModeManual {
		if lockState.groupID == nil || lockState.proxyID == nil {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidLock
			return candidate
		}
		targetPoolIdx, ok := poolIndexByGroupID[*lockState.groupID]
		if !ok {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidMapping
			return candidate
		}
		targetPool := cfg.Pools[targetPoolIdx]
		if *lockState.proxyID != targetPool.ProxyID {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidMapping
			return candidate
		}
		if currentGroupID == nil || *currentGroupID != targetPool.GroupID || currentProxyID == nil || *currentProxyID != targetPool.ProxyID {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidLock
			candidate.targetGroupID = int64PtrPool(targetPool.GroupID)
			candidate.targetProxyID = int64PtrPool(targetPool.ProxyID)
		}
		return candidate
	}

	if currentGroupID == nil {
		candidate.needsMove = true
		candidate.reason = openAIFreePoolReasonInvalidMapping
		return candidate
	}

	currentPoolIdx, ok := poolIndexByGroupID[*currentGroupID]
	if !ok {
		candidate.needsMove = true
		candidate.reason = openAIFreePoolReasonInvalidMapping
		return candidate
	}
	currentPool := cfg.Pools[currentPoolIdx]

	if lockState.locked {
		if lockState.groupID == nil || lockState.proxyID == nil {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidLock
			return candidate
		}
		if *lockState.groupID != currentPool.GroupID || *lockState.proxyID != currentPool.ProxyID {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidLock
			return candidate
		}
		if currentProxyID == nil || *currentProxyID != currentPool.ProxyID {
			candidate.needsMove = true
			candidate.reason = openAIFreePoolReasonInvalidLock
			return candidate
		}
		return candidate
	}

	if currentProxyID == nil || *currentProxyID != currentPool.ProxyID {
		candidate.needsMove = true
		candidate.reason = openAIFreePoolReasonInvalidMapping
		return candidate
	}

	return candidate
}

func detectManagedOpenAIFreeGroup(account *Account, cfg *OpenAIFreePoolConfig, groupByID map[int64]Group, poolIndexByGroupID map[int64]int) (*int64, string) {
	groupIDs := account.GroupIDs
	if len(groupIDs) == 0 && len(account.Groups) > 0 {
		for _, group := range account.Groups {
			if group != nil {
				groupIDs = append(groupIDs, group.ID)
			}
		}
	}
	for _, groupID := range groupIDs {
		if groupID == cfg.DefaultGroupID {
			return int64PtrPool(groupID), groupNameByID(groupByID, groupID)
		}
		if _, ok := poolIndexByGroupID[groupID]; ok {
			return int64PtrPool(groupID), groupNameByID(groupByID, groupID)
		}
	}
	return nil, ""
}

func detectProxy(account *Account, proxyByID map[int64]Proxy) (*int64, string) {
	if account == nil || account.ProxyID == nil {
		return nil, ""
	}
	name := ""
	if proxy, ok := proxyByID[*account.ProxyID]; ok {
		name = proxy.Name
	}
		return int64PtrPool(*account.ProxyID), name
}

func extractOpenAIFreeReset(account *Account) (*string, string) {
	if account == nil {
		return nil, openAIFreePoolUnknownResetDate
	}
	resetRaw := strings.TrimSpace(account.GetExtraString("codex_7d_reset_at"))
	if resetRaw == "" {
		return nil, openAIFreePoolUnknownResetDate
	}
	parsed, err := time.Parse(time.RFC3339, resetRaw)
	if err != nil {
		return nil, openAIFreePoolUnknownResetDate
	}
	formatted := parsed.UTC().Format(time.RFC3339)
	date := parsed.UTC().Format("2006-01-02")
	return &formatted, date
}

type openAIFreePoolLockState struct {
	groupID *int64
	proxyID *int64
	locked  bool
	mode    string
}

func extractOpenAIFreeLockState(account *Account) openAIFreePoolLockState {
	if account == nil || account.Extra == nil {
		return openAIFreePoolLockState{mode: OpenAIFreePoolLockModeUnlocked}
	}
	if !strings.EqualFold(strings.TrimSpace(account.GetExtraString("auto_pool_source")), OpenAIFreePoolSource) {
		return openAIFreePoolLockState{mode: OpenAIFreePoolLockModeUnlocked}
	}
	mode := strings.ToLower(strings.TrimSpace(account.GetExtraString("auto_pool_lock_mode")))
	switch mode {
	case OpenAIFreePoolLockModeManual, OpenAIFreePoolLockModeAuto:
	default:
		mode = OpenAIFreePoolLockModeAuto
	}
	groupID, groupOK := extraInt64(account.Extra["auto_pool_group_id"])
	proxyID, proxyOK := extraInt64(account.Extra["auto_pool_proxy_id"])
	state := openAIFreePoolLockState{
		locked: true,
		mode:   mode,
	}
	if groupOK {
		state.groupID = int64PtrPool(groupID)
	}
	if proxyOK {
		state.proxyID = int64PtrPool(proxyID)
	}
	return state
}

func chooseOpenAIFreePoolTarget(
	pools []OpenAIFreePool,
	counts []int,
	dateCountByPool []map[string]int,
	poolStates []OpenAIFreePoolState,
	dateAnchors map[string]openAIFreePoolDateAnchor,
	resetDate string,
) int {
	if resetDate == openAIFreePoolUnknownResetDate {
		return chooseUnknownOpenAIFreePoolTarget(pools, poolStates)
	}
	if anchor, ok := dateAnchors[resetDate]; ok {
		return anchor.poolIdx
	}
	return chooseDateAnchorOpenAIFreePoolTarget(pools, counts, dateCountByPool)
}

func buildOpenAIFreePoolDateAnchors(
	pools []OpenAIFreePool,
	dateCountByPool []map[string]int,
) map[string]openAIFreePoolDateAnchor {
	anchors := make(map[string]openAIFreePoolDateAnchor)
	for poolIdx, byDate := range dateCountByPool {
		for resetDate, count := range byDate {
			if resetDate == openAIFreePoolUnknownResetDate || count <= 0 {
				continue
			}
			current, exists := anchors[resetDate]
			if !exists || count > dateCountByPool[current.poolIdx][resetDate] {
				anchors[resetDate] = openAIFreePoolDateAnchor{
					poolIdx: poolIdx,
					locked:  true,
				}
				continue
			}
			if count < dateCountByPool[current.poolIdx][resetDate] {
				continue
			}
			if pools[poolIdx].SortOrder < pools[current.poolIdx].SortOrder {
				anchors[resetDate] = openAIFreePoolDateAnchor{
					poolIdx: poolIdx,
					locked:  true,
				}
				continue
			}
			if pools[poolIdx].SortOrder == pools[current.poolIdx].SortOrder && pools[poolIdx].GroupID < pools[current.poolIdx].GroupID {
				anchors[resetDate] = openAIFreePoolDateAnchor{
					poolIdx: poolIdx,
					locked:  true,
				}
			}
		}
	}
	return anchors
}

func chooseUnknownOpenAIFreePoolTarget(pools []OpenAIFreePool, poolStates []OpenAIFreePoolState) int {
	bestIdx := 0
	for i := 1; i < len(pools); i++ {
		if poolStates[i].UnknownResetAccounts < poolStates[bestIdx].UnknownResetAccounts {
			bestIdx = i
			continue
		}
		if poolStates[i].UnknownResetAccounts > poolStates[bestIdx].UnknownResetAccounts {
			continue
		}
		if pools[i].SortOrder < pools[bestIdx].SortOrder {
			bestIdx = i
			continue
		}
		if pools[i].SortOrder == pools[bestIdx].SortOrder && pools[i].GroupID < pools[bestIdx].GroupID {
			bestIdx = i
		}
	}
	return bestIdx
}

func chooseDateAnchorOpenAIFreePoolTarget(pools []OpenAIFreePool, counts []int, dateCountByPool []map[string]int) int {
	bestIdx := 0
	for i := 1; i < len(pools); i++ {
		leftDates := len(dateCountByPool[i])
		rightDates := len(dateCountByPool[bestIdx])
		if leftDates < rightDates {
			bestIdx = i
			continue
		}
		if leftDates > rightDates {
			continue
		}
		if counts[i] < counts[bestIdx] {
			bestIdx = i
			continue
		}
		if counts[i] > counts[bestIdx] {
			continue
		}
		if pools[i].SortOrder < pools[bestIdx].SortOrder {
			bestIdx = i
			continue
		}
		if pools[i].SortOrder == pools[bestIdx].SortOrder && pools[i].GroupID < pools[bestIdx].GroupID {
			bestIdx = i
		}
	}
	return bestIdx
}

func buildOpenAIFreeResetForecast(cfg *OpenAIFreePoolConfig, groups []Group, proxies []Proxy, accounts []*Account) *OpenAIFreeResetForecast {
	groupByID := make(map[int64]Group, len(groups))
	for _, group := range groups {
		groupByID[group.ID] = group
	}
	proxyByID := make(map[int64]Proxy, len(proxies))
	for _, proxy := range proxies {
		proxyByID[proxy.ID] = proxy
	}

	byDate := map[string][]OpenAIFreeResetForecastAccount{}
	unknownCount := 0
	now := time.Now().UTC()
	lookaheadEnd := now.AddDate(0, 0, cfg.LookaheadDays)
	for _, account := range accounts {
		groupID, groupName := detectManagedOpenAIFreeGroup(account, cfg, groupByID, managedPoolIndex(cfg))
		proxyID, proxyName := detectProxy(account, proxyByID)
		resetAt, resetDate := extractOpenAIFreeReset(account)
		if resetDate == openAIFreePoolUnknownResetDate {
			unknownCount++
			continue
		}
		resetTime, err := time.Parse(time.RFC3339, *resetAt)
		if err != nil {
			unknownCount++
			continue
		}
		if resetTime.Before(now) || !resetTime.Before(lookaheadEnd) {
			continue
		}
		inDefault := groupID != nil && *groupID == cfg.DefaultGroupID
		byDate[resetDate] = append(byDate[resetDate], OpenAIFreeResetForecastAccount{
			AccountID:      account.ID,
			AccountName:    account.Name,
			GroupID:        groupID,
			GroupName:      groupName,
			ProxyID:        proxyID,
			ProxyName:      proxyName,
			InDefaultGroup: inDefault,
			UsagePercent:   extractOpenAIFreeUsagePercent(account),
			ResetAt:        resetAt,
		})
	}

	dates := make([]string, 0, len(byDate))
	for date := range byDate {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	days := make([]OpenAIFreeResetForecastDay, 0, len(dates))
	for _, date := range dates {
		accountsForDay := byDate[date]
		sort.Slice(accountsForDay, func(i, j int) bool {
			return accountsForDay[i].AccountID < accountsForDay[j].AccountID
		})
		days = append(days, OpenAIFreeResetForecastDay{
			Date:     date,
			Count:    len(accountsForDay),
			Accounts: accountsForDay,
		})
	}

	return &OpenAIFreeResetForecast{
		Days:         days,
		UnknownCount: unknownCount,
		GeneratedAt:  time.Now().UTC().Format(time.RFC3339),
	}
}

func managedPoolIndex(cfg *OpenAIFreePoolConfig) map[int64]int {
	result := make(map[int64]int, len(cfg.Pools))
	for i, pool := range cfg.Pools {
		result[pool.GroupID] = i
	}
	return result
}

func groupNameByID(groups map[int64]Group, groupID int64) string {
	if group, ok := groups[groupID]; ok {
		return group.Name
	}
	return ""
}

func proxyNameByID(proxies map[int64]Proxy, proxyID int64) string {
	if proxy, ok := proxies[proxyID]; ok {
		return proxy.Name
	}
	return ""
}

func findManagedOpenAIFreeAccount(accounts []*Account, accountID int64) *Account {
	for _, account := range accounts {
		if account != nil && account.ID == accountID {
			return account
		}
	}
	return nil
}

func extraInt64(raw any) (int64, bool) {
	switch value := raw.(type) {
	case int64:
		return value, true
	case int:
		return int64(value), true
	case float64:
		return int64(value), true
	case json.Number:
		v, err := value.Int64()
		return v, err == nil
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0, false
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func extraFloat64(raw any) (float64, bool) {
	switch value := raw.(type) {
	case float64:
		return value, true
	case float32:
		return float64(value), true
	case int64:
		return float64(value), true
	case int:
		return float64(value), true
	case json.Number:
		parsed, err := value.Float64()
		return parsed, err == nil
	case string:
		value = strings.TrimSpace(value)
		if value == "" {
			return 0, false
		}
		parsed, err := strconv.ParseFloat(value, 64)
		return parsed, err == nil
	default:
		return 0, false
	}
}

func cfgClone(cfg *OpenAIFreePoolConfig) OpenAIFreePoolConfig {
	if cfg == nil {
		return OpenAIFreePoolConfig{}
	}
	cloned := *cfg
	if len(cfg.Pools) > 0 {
		cloned.Pools = append([]OpenAIFreePool(nil), cfg.Pools...)
	}
	return cloned
}

func int64PtrPool(v int64) *int64 {
	return &v
}

func float64PtrPool(v float64) *float64 {
	return &v
}

func cloneExtraMap(extra map[string]any) map[string]any {
	if len(extra) == 0 {
		return map[string]any{}
	}
	cloned := make(map[string]any, len(extra))
	for key, value := range extra {
		cloned[key] = value
	}
	return cloned
}

func extractOpenAIFreeUsagePercent(account *Account) *float64 {
	if account == nil || account.Extra == nil {
		return nil
	}
	value, ok := extraFloat64(account.Extra["codex_7d_used_percent"])
	if !ok {
		return nil
	}
	return float64PtrPool(value)
}
