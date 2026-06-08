import type {
  Account,
  AdminGroup,
  DashboardStats,
  Group,
  OpenAIFreePoolManagedAccount,
  ModelStat,
  OpenAIFreePoolApplyResult,
  OpenAIFreePoolConfig,
  OpenAIFreePoolPreview,
  OpenAIFreeResetForecast,
  Proxy,
  PublicSettings,
  TrendDataPoint,
  User,
  UserAnnouncement,
  UserSpendingRankingResponse,
  UserSubscription,
  UserUsageTrendPoint
} from '@/types'

const DEV_PREVIEW_MODE_KEY = 'sub2api_dev_preview_mode'
const DEV_PREVIEW_TOKEN = 'sub2api-dev-preview-token'

type RequestParams = Record<string, unknown> | undefined

interface DevPreviewMockResponse {
  data: unknown
  headers?: Record<string, string>
  status?: number
}

const now = '2026-05-22T15:45:00+08:00'

function isoAt(day: string, time: string): string {
  return `${day}T${time}+08:00`
}

function buildAdminGroup(
  id: number,
  name: string,
  sortOrder: number,
  accountCount: number,
  activeAccountCount: number = accountCount
): AdminGroup {
  return {
    id,
    name,
    description: null,
    platform: 'openai',
    rate_multiplier: 1,
    is_exclusive: false,
    status: 'active',
    subscription_type: 'standard',
    daily_limit_usd: null,
    weekly_limit_usd: null,
    monthly_limit_usd: null,
    allow_image_generation: true,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    claude_code_only: false,
    fallback_group_id: null,
    fallback_group_id_on_invalid_request: null,
    require_oauth_only: false,
    require_privacy_set: false,
    created_at: now,
    updated_at: now,
    model_routing: null,
    model_routing_enabled: false,
    mcp_xml_inject: false,
    supported_model_scopes: [],
    account_count: accountCount,
    active_account_count: activeAccountCount,
    rate_limited_account_count: 0,
    sort_order: sortOrder
  }
}

const groups: AdminGroup[] = [
  buildAdminGroup(1, 'default', 10, 1),
  buildAdminGroup(2, 'free 秘鲁', 20, 1),
  buildAdminGroup(3, 'free 巴西', 30, 1),
  buildAdminGroup(4, 'free 智利', 40, 1),
  buildAdminGroup(5, 'free 墨西哥', 50, 1),
  buildAdminGroup(6, 'free 哥伦比亚', 60, 1),
  buildAdminGroup(7, 'plus', 70, 1)
]

const proxies: Proxy[] = [
  {
    id: 101,
    name: '17899 秘鲁',
    protocol: 'http',
    host: '127.0.0.1',
    port: 17899,
    username: null,
    status: 'active',
    account_count: 1,
    ip_address: '181.177.12.34',
    country: 'Peru',
    country_code: 'PE',
    region: 'Lima',
    city: 'Lima',
    quality_status: 'healthy',
    quality_score: 94,
    quality_grade: 'A',
    quality_summary: 'Stable residential exit IP',
    quality_checked: Date.now(),
    expires_at: null,
    fallback_mode: 'none',
    expiry_warn_days: 7,
    created_at: now,
    updated_at: now
  },
  {
    id: 102,
    name: '17900 巴西',
    protocol: 'http',
    host: '127.0.0.1',
    port: 17900,
    username: null,
    status: 'active',
    account_count: 1,
    ip_address: '177.55.12.98',
    country: 'Brazil',
    country_code: 'BR',
    region: 'Sao Paulo',
    city: 'Sao Paulo',
    quality_status: 'healthy',
    quality_score: 92,
    quality_grade: 'A',
    quality_summary: 'Healthy exit IP',
    quality_checked: Date.now(),
    expires_at: null,
    fallback_mode: 'none',
    expiry_warn_days: 7,
    created_at: now,
    updated_at: now
  },
  {
    id: 103,
    name: '17901 智利',
    protocol: 'http',
    host: '127.0.0.1',
    port: 17901,
    username: null,
    status: 'active',
    account_count: 1,
    ip_address: '186.67.45.21',
    country: 'Chile',
    country_code: 'CL',
    region: 'Santiago',
    city: 'Santiago',
    quality_status: 'healthy',
    quality_score: 91,
    quality_grade: 'A',
    quality_summary: 'Healthy exit IP',
    quality_checked: Date.now(),
    expires_at: null,
    fallback_mode: 'none',
    expiry_warn_days: 7,
    created_at: now,
    updated_at: now
  },
  {
    id: 104,
    name: '17902 墨西哥',
    protocol: 'http',
    host: '127.0.0.1',
    port: 17902,
    username: null,
    status: 'active',
    account_count: 1,
    ip_address: '189.201.77.10',
    country: 'Mexico',
    country_code: 'MX',
    region: 'Mexico City',
    city: 'Mexico City',
    quality_status: 'healthy',
    quality_score: 90,
    quality_grade: 'A',
    quality_summary: 'Healthy exit IP',
    quality_checked: Date.now(),
    expires_at: null,
    fallback_mode: 'none',
    expiry_warn_days: 7,
    created_at: now,
    updated_at: now
  },
  {
    id: 105,
    name: '17903 哥伦比亚',
    protocol: 'http',
    host: '127.0.0.1',
    port: 17903,
    username: null,
    status: 'active',
    account_count: 1,
    ip_address: '181.57.22.100',
    country: 'Colombia',
    country_code: 'CO',
    region: 'Bogota',
    city: 'Bogota',
    quality_status: 'warn',
    quality_score: 82,
    quality_grade: 'B',
    quality_summary: 'Available with occasional challenges',
    quality_checked: Date.now(),
    expires_at: null,
    fallback_mode: 'none',
    expiry_warn_days: 7,
    created_at: now,
    updated_at: now
  }
]

const adminUser: User & { run_mode?: 'standard' | 'simple' } = {
  id: 1,
  username: 'preview-admin',
  email: 'preview@local.test',
  role: 'admin',
  balance: 0,
  concurrency: 99,
  status: 'active',
  allowed_groups: null,
  balance_notify_enabled: false,
  balance_notify_threshold: null,
  balance_notify_extra_emails: [],
  created_at: now,
  updated_at: now,
  run_mode: 'standard'
}

function buildAccount(
  id: number,
  name: string,
  groupId: number,
  proxyId: number | null,
  extra: Record<string, unknown>,
  lastUsedAt: string | null = now
): Account {
  const group = groups.find((item) => item.id === groupId) as Group
  const proxy = proxies.find((item) => item.id === proxyId) ?? undefined
  return {
    id,
    name,
    notes: null,
    platform: 'openai',
    type: 'oauth',
    credentials: {
      plan_type: groupId === 7 ? 'plus' : 'free',
      email: `${name.replace(/\s+/g, '.').toLowerCase()}@example.com`
    },
    credentials_status: {},
    extra,
    proxy_id: proxyId,
    concurrency: 1,
    load_factor: 1,
    current_concurrency: 0,
    priority: 10,
    rate_multiplier: 0,
    status: 'active',
    error_message: null,
    last_used_at: lastUsedAt,
    expires_at: null,
    auto_pause_on_expired: false,
    created_at: now,
    updated_at: now,
    proxy,
    group_ids: [groupId],
    groups: [group],
    schedulable: true,
    rate_limited_at: null,
    rate_limit_reset_at: null,
    overload_until: null,
    temp_unschedulable_until: null,
    temp_unschedulable_reason: null,
    session_window_start: null,
    session_window_end: null,
    session_window_status: null
  }
}

const accounts: Account[] = [
  buildAccount(201, 'free 秘鲁', 2, 101, {
    codex_7d_reset_at: isoAt('2026-05-23', '08:30:00'),
    auto_pool_source: 'openai_free_pool_v1',
    auto_pool_group_id: 2,
    auto_pool_proxy_id: 101,
    auto_pool_locked_at: now
  }),
  buildAccount(202, 'free 巴西', 3, 102, {
    codex_7d_reset_at: isoAt('2026-05-23', '18:10:00'),
    auto_pool_source: 'openai_free_pool_v1',
    auto_pool_group_id: 3,
    auto_pool_proxy_id: 102,
    auto_pool_locked_at: now
  }),
  buildAccount(203, 'free 智利', 4, 103, {
    codex_7d_reset_at: isoAt('2026-05-24', '09:20:00'),
    auto_pool_source: 'openai_free_pool_v1',
    auto_pool_group_id: 4,
    auto_pool_proxy_id: 103,
    auto_pool_locked_at: now
  }),
  buildAccount(204, 'free 墨西哥', 5, 104, {
    codex_7d_reset_at: isoAt('2026-05-24', '22:15:00'),
    auto_pool_source: 'openai_free_pool_v1',
    auto_pool_group_id: 5,
    auto_pool_proxy_id: 104,
    auto_pool_locked_at: now
  }),
  buildAccount(205, 'free 哥伦比亚', 6, 105, {
    auto_pool_source: 'openai_free_pool_v1',
    auto_pool_group_id: 6,
    auto_pool_proxy_id: 105,
    auto_pool_locked_at: now
  }),
  buildAccount(206, 'free intake 01', 1, null, {
    codex_7d_reset_at: isoAt('2026-05-25', '11:00:00')
  }, null),
  buildAccount(207, 'plus 样例', 7, null, {
    codex_7d_reset_at: isoAt('2026-05-26', '12:00:00')
  })
]

const openAIFreePoolConfig: OpenAIFreePoolConfig = {
  enabled: true,
  default_group_id: 1,
  plus_group_id: 7,
  lookahead_days: 14,
  pools: [
    { group_id: 2, proxy_id: 101, label: 'Peru pool', sort_order: 10 },
    { group_id: 3, proxy_id: 102, label: 'Brazil pool', sort_order: 20 },
    { group_id: 4, proxy_id: 103, label: 'Chile pool', sort_order: 30 },
    { group_id: 5, proxy_id: 104, label: 'Mexico pool', sort_order: 40 },
    { group_id: 6, proxy_id: 105, label: 'Colombia pool', sort_order: 50 }
  ]
}

const openAIFreePoolPreview: OpenAIFreePoolPreview = {
  config: openAIFreePoolConfig,
  summary: {
    managed_accounts: 6,
    default_accounts: 1,
    locked_accounts: 5,
    unknown_reset_accounts: 1,
    pending_moves: 1
  },
  pools: [
    { group_id: 2, group_name: 'free 秘鲁', proxy_id: 101, proxy_name: '17899 秘鲁', accounts: 1, locked_accounts: 1, unknown_reset_accounts: 0 },
    { group_id: 3, group_name: 'free 巴西', proxy_id: 102, proxy_name: '17900 巴西', accounts: 1, locked_accounts: 1, unknown_reset_accounts: 0 },
    { group_id: 4, group_name: 'free 智利', proxy_id: 103, proxy_name: '17901 智利', accounts: 1, locked_accounts: 1, unknown_reset_accounts: 0 },
    { group_id: 5, group_name: 'free 墨西哥', proxy_id: 104, proxy_name: '17902 墨西哥', accounts: 1, locked_accounts: 1, unknown_reset_accounts: 0 },
    { group_id: 6, group_name: 'free 哥伦比亚', proxy_id: 105, proxy_name: '17903 哥伦比亚', accounts: 1, locked_accounts: 1, unknown_reset_accounts: 1 }
  ],
  moves: [
    {
      account_id: 206,
      account_name: 'free intake 01',
      current_group_id: 1,
      current_group_name: 'default',
      target_group_id: 3,
      target_group_name: 'free 巴西',
      current_proxy_id: undefined,
      current_proxy_name: undefined,
      target_proxy_id: 102,
      target_proxy_name: '17900 巴西',
      reset_at: isoAt('2026-05-25', '11:00:00'),
      reset_date: '2026-05-25',
      locked: false,
      reason: 'new_from_default'
    }
  ],
  unknown_reset_ids: [205],
  generated_at: now,
  force_rebalance: false
}

const openAIFreePoolApplyResult: OpenAIFreePoolApplyResult = {
  applied: 1,
  skipped: 0,
  failed: 0,
  preview: openAIFreePoolPreview,
  generated_at: now
}

const openAIFreePoolManagedAccounts: OpenAIFreePoolManagedAccount[] = [
  {
    account_id: 201,
    account_name: 'free 秘鲁',
    current_group_id: 2,
    current_group_name: 'free 秘鲁',
    current_proxy_id: 101,
    current_proxy_name: '17899 秘鲁',
    reset_at: isoAt('2026-05-23', '08:30:00'),
    in_default_group: false,
    lock_mode: 'auto',
    lock_group_id: 2,
    lock_group_name: 'free 秘鲁',
    lock_proxy_id: 101,
    lock_proxy_name: '17899 秘鲁'
  },
  {
    account_id: 202,
    account_name: 'free 巴西',
    current_group_id: 3,
    current_group_name: 'free 巴西',
    current_proxy_id: 102,
    current_proxy_name: '17900 巴西',
    reset_at: isoAt('2026-05-23', '18:10:00'),
    in_default_group: false,
    lock_mode: 'manual',
    lock_group_id: 3,
    lock_group_name: 'free 巴西',
    lock_proxy_id: 102,
    lock_proxy_name: '17900 巴西'
  },
  {
    account_id: 203,
    account_name: 'free 智利',
    current_group_id: 4,
    current_group_name: 'free 智利',
    current_proxy_id: 103,
    current_proxy_name: '17901 智利',
    reset_at: isoAt('2026-05-24', '09:20:00'),
    in_default_group: false,
    lock_mode: 'auto',
    lock_group_id: 4,
    lock_group_name: 'free 智利',
    lock_proxy_id: 103,
    lock_proxy_name: '17901 智利'
  },
  {
    account_id: 204,
    account_name: 'free 墨西哥',
    current_group_id: 5,
    current_group_name: 'free 墨西哥',
    current_proxy_id: 104,
    current_proxy_name: '17902 墨西哥',
    reset_at: isoAt('2026-05-24', '22:15:00'),
    in_default_group: false,
    lock_mode: 'auto',
    lock_group_id: 5,
    lock_group_name: 'free 墨西哥',
    lock_proxy_id: 104,
    lock_proxy_name: '17902 墨西哥'
  },
  {
    account_id: 205,
    account_name: 'free 哥伦比亚',
    current_group_id: 6,
    current_group_name: 'free 哥伦比亚',
    current_proxy_id: 105,
    current_proxy_name: '17903 哥伦比亚',
    in_default_group: false,
    lock_mode: 'unlocked',
    lock_group_id: 6,
    lock_group_name: 'free 哥伦比亚',
    lock_proxy_id: 105,
    lock_proxy_name: '17903 哥伦比亚'
  },
  {
    account_id: 206,
    account_name: 'free intake 01',
    current_group_id: 1,
    current_group_name: 'default',
    reset_at: isoAt('2026-05-25', '11:00:00'),
    in_default_group: true,
    lock_mode: 'unlocked'
  }
]

const openAIFreeResetForecast: OpenAIFreeResetForecast = {
  days: [
    {
      date: '2026-05-23',
      count: 2,
      accounts: [
        {
          account_id: 201,
          account_name: 'free 秘鲁',
          group_id: 2,
          group_name: 'free 秘鲁',
          proxy_id: 101,
          proxy_name: '17899 秘鲁',
          in_default_group: false,
          usage_percent: 82,
          reset_at: isoAt('2026-05-23', '08:30:00')
        },
        {
          account_id: 202,
          account_name: 'free 巴西',
          group_id: 3,
          group_name: 'free 巴西',
          proxy_id: 102,
          proxy_name: '17900 巴西',
          in_default_group: false,
          usage_percent: 37,
          reset_at: isoAt('2026-05-23', '18:10:00')
        }
      ]
    },
    {
      date: '2026-05-24',
      count: 2,
      accounts: [
        {
          account_id: 203,
          account_name: 'free 智利',
          group_id: 4,
          group_name: 'free 智利',
          proxy_id: 103,
          proxy_name: '17901 智利',
          in_default_group: false,
          usage_percent: 64,
          reset_at: isoAt('2026-05-24', '09:20:00')
        },
        {
          account_id: 204,
          account_name: 'free 墨西哥',
          group_id: 5,
          group_name: 'free 墨西哥',
          proxy_id: 104,
          proxy_name: '17902 墨西哥',
          in_default_group: false,
          usage_percent: 18,
          reset_at: isoAt('2026-05-24', '22:15:00')
        }
      ]
    },
    {
      date: '2026-05-25',
      count: 1,
      accounts: [
        {
          account_id: 206,
          account_name: 'free intake 01',
          group_id: 1,
          group_name: 'default',
          proxy_name: '',
          in_default_group: true,
          usage_percent: 0,
          reset_at: isoAt('2026-05-25', '11:00:00')
        }
      ]
    }
  ],
  unknown_count: 1,
  generated_at: now
}

const dashboardStats: DashboardStats = {
  total_users: 128,
  today_new_users: 7,
  active_users: 19,
  hourly_active_users: 6,
  stats_updated_at: now,
  stats_stale: false,
  total_api_keys: 214,
  active_api_keys: 187,
  total_accounts: 7,
  normal_accounts: 6,
  error_accounts: 0,
  ratelimit_accounts: 0,
  overload_accounts: 0,
  total_requests: 18420,
  total_input_tokens: 1280000,
  total_output_tokens: 960000,
  total_cache_creation_tokens: 320000,
  total_cache_read_tokens: 210000,
  total_tokens: 2770000,
  total_cost: 391.22,
  total_actual_cost: 286.48,
  total_account_cost: 204.17,
  today_requests: 968,
  today_input_tokens: 68200,
  today_output_tokens: 51400,
  today_cache_creation_tokens: 9100,
  today_cache_read_tokens: 7000,
  today_tokens: 135700,
  today_cost: 18.26,
  today_actual_cost: 12.81,
  today_account_cost: 8.64,
  average_duration_ms: 912,
  uptime: 864000,
  rpm: 14,
  tpm: 2120
}

const dashboardTrend: TrendDataPoint[] = [
  { date: '09:00', requests: 88, input_tokens: 8200, output_tokens: 6100, cache_creation_tokens: 600, cache_read_tokens: 300, total_tokens: 15200, cost: 1.82, actual_cost: 1.21 },
  { date: '10:00', requests: 116, input_tokens: 9300, output_tokens: 6800, cache_creation_tokens: 800, cache_read_tokens: 500, total_tokens: 17400, cost: 2.14, actual_cost: 1.47 },
  { date: '11:00', requests: 140, input_tokens: 10200, output_tokens: 7700, cache_creation_tokens: 900, cache_read_tokens: 600, total_tokens: 19400, cost: 2.46, actual_cost: 1.65 },
  { date: '12:00', requests: 172, input_tokens: 11800, output_tokens: 8800, cache_creation_tokens: 1100, cache_read_tokens: 700, total_tokens: 22400, cost: 2.88, actual_cost: 1.98 },
  { date: '13:00', requests: 160, input_tokens: 10900, output_tokens: 8300, cache_creation_tokens: 1000, cache_read_tokens: 650, total_tokens: 20850, cost: 2.61, actual_cost: 1.79 },
  { date: '14:00', requests: 151, input_tokens: 9800, output_tokens: 7400, cache_creation_tokens: 900, cache_read_tokens: 550, total_tokens: 18650, cost: 2.35, actual_cost: 1.61 }
]

const dashboardModels: ModelStat[] = [
  { model: 'gpt-4.1', requests: 322, input_tokens: 280000, output_tokens: 201000, cache_creation_tokens: 41000, cache_read_tokens: 16000, total_tokens: 538000, cost: 72.18, actual_cost: 54.04, account_cost: 36.7 },
  { model: 'gpt-4.1-mini', requests: 418, input_tokens: 190000, output_tokens: 144000, cache_creation_tokens: 21000, cache_read_tokens: 15000, total_tokens: 370000, cost: 31.46, actual_cost: 21.7, account_cost: 15.88 },
  { model: 'o4-mini', requests: 228, input_tokens: 122000, output_tokens: 94000, cache_creation_tokens: 16000, cache_read_tokens: 9000, total_tokens: 241000, cost: 18.2, actual_cost: 12.1, account_cost: 8.4 }
]

const usersTrend: UserUsageTrendPoint[] = [
  { date: '09:00', user_id: 11, email: 'ops@example.com', username: 'ops', requests: 18, tokens: 4200, cost: 0.31, actual_cost: 0.2 },
  { date: '10:00', user_id: 11, email: 'ops@example.com', username: 'ops', requests: 24, tokens: 5100, cost: 0.42, actual_cost: 0.29 },
  { date: '11:00', user_id: 11, email: 'ops@example.com', username: 'ops', requests: 29, tokens: 6400, cost: 0.56, actual_cost: 0.37 },
  { date: '12:00', user_id: 12, email: 'team@example.com', username: 'team', requests: 31, tokens: 7300, cost: 0.61, actual_cost: 0.41 },
  { date: '13:00', user_id: 12, email: 'team@example.com', username: 'team', requests: 26, tokens: 6900, cost: 0.57, actual_cost: 0.38 },
  { date: '14:00', user_id: 13, email: 'admin@example.com', username: 'admin', requests: 35, tokens: 8100, cost: 0.68, actual_cost: 0.45 }
]

const userRanking: UserSpendingRankingResponse = {
  ranking: [
    { user_id: 13, email: 'admin@example.com', actual_cost: 4.82, requests: 181, tokens: 41200 },
    { user_id: 12, email: 'team@example.com', actual_cost: 3.77, requests: 143, tokens: 33500 },
    { user_id: 11, email: 'ops@example.com', actual_cost: 2.64, requests: 101, tokens: 25800 }
  ],
  total_actual_cost: 11.23,
  total_requests: 425,
  total_tokens: 100500,
  start_date: '2026-05-21',
  end_date: '2026-05-22'
}

const publicSettings: PublicSettings = {
  registration_enabled: true,
  email_verify_enabled: false,
  force_email_on_third_party_signup: false,
  registration_email_suffix_whitelist: [],
  promo_code_enabled: false,
  password_reset_enabled: false,
  invitation_code_enabled: false,
  login_agreement_enabled: false,
  login_agreement_mode: 'modal',
  login_agreement_updated_at: '',
  login_agreement_revision: '',
  login_agreement_documents: [],
  turnstile_enabled: false,
  turnstile_site_key: '',
  site_name: 'Sub2API DEV Preview',
  site_logo: '',
  site_subtitle: 'Frontend preview without backend',
  api_base_url: '/api/v1',
  contact_info: '',
  doc_url: '',
  home_content: '',
  hide_ccs_import_button: false,
  payment_enabled: false,
  risk_control_enabled: false,
  table_default_page_size: 20,
  table_page_size_options: [10, 20, 50, 100],
  custom_menu_items: [],
  custom_endpoints: [],
  linuxdo_oauth_enabled: false,
  dingtalk_oauth_enabled: false,
  wechat_oauth_enabled: false,
  wechat_oauth_open_enabled: false,
  wechat_oauth_mp_enabled: false,
  wechat_oauth_mobile_enabled: false,
  oidc_oauth_enabled: false,
  oidc_oauth_provider_name: 'OIDC',
  github_oauth_enabled: false,
  google_oauth_enabled: false,
  backend_mode_enabled: false,
  version: 'dev-preview',
  balance_low_notify_enabled: false,
  account_quota_notify_enabled: false,
  balance_low_notify_threshold: 0,
  channel_monitor_enabled: true,
  channel_monitor_default_interval_seconds: 60,
  available_channels_enabled: false,
  service_quota_enabled: false,
  affiliate_enabled: false
}

const userAnnouncements: UserAnnouncement[] = []
const activeSubscriptions: UserSubscription[] = []

function filterAccounts(items: Account[], params: RequestParams): Account[] {
  let result = [...items]
  const platform = String(params?.platform ?? '').trim()
  const type = String(params?.type ?? '').trim()
  const status = String(params?.status ?? '').trim()
  const search = String(params?.search ?? '').trim().toLowerCase()
  const group = String(params?.group ?? '').trim()

  if (platform) result = result.filter((item) => item.platform === platform)
  if (type) result = result.filter((item) => item.type === type)
  if (status) result = result.filter((item) => item.status === status)
  if (group) result = result.filter((item) => (item.group_ids ?? []).includes(Number(group)))
  if (search) result = result.filter((item) => item.name.toLowerCase().includes(search))

  return result
}

function paginateAccounts(params: RequestParams) {
  const page = Math.max(1, Number(params?.page ?? 1))
  const pageSize = Math.max(1, Number(params?.page_size ?? 20))
  const filtered = filterAccounts(accounts, params)
  const start = (page - 1) * pageSize
  const items = filtered.slice(start, start + pageSize)

  return {
    items,
    total: filtered.length,
    page,
    page_size: pageSize,
    pages: Math.max(1, Math.ceil(filtered.length / pageSize))
  }
}

function buildSnapshot(params: RequestParams) {
  return {
    generated_at: now,
    start_date: String(params?.start_date ?? '2026-05-21'),
    end_date: String(params?.end_date ?? '2026-05-22'),
    granularity: String(params?.granularity ?? 'hour'),
    stats: dashboardStats,
    trend: dashboardTrend,
    models: dashboardModels,
    groups: []
  }
}

function normalizePath(url?: string): string {
  if (!url) return ''
  return url.split('?')[0]
}

export function isDevPreviewAvailable(): boolean {
  return import.meta.env.DEV
}

export function isDevPreviewEnabled(): boolean {
  if (!isDevPreviewAvailable()) return false
  try {
    return localStorage.getItem(DEV_PREVIEW_MODE_KEY) === 'admin'
  } catch {
    return false
  }
}

export function enableDevPreviewAdmin(): void {
  if (!isDevPreviewAvailable()) return

  try {
    localStorage.setItem(DEV_PREVIEW_MODE_KEY, 'admin')
    localStorage.setItem('auth_token', DEV_PREVIEW_TOKEN)
    localStorage.setItem('auth_user', JSON.stringify(adminUser))
    localStorage.setItem('token_expires_at', String(Date.now() + 24 * 60 * 60 * 1000))
  } catch {
    // Ignore localStorage failures in local preview mode.
  }
}

export function getDevPreviewMockResponse(
  method: string | undefined,
  url: string | undefined,
  params?: RequestParams
): DevPreviewMockResponse | null {
  if (!isDevPreviewEnabled()) {
    return null
  }

  const normalizedMethod = String(method ?? 'get').toLowerCase()
  const path = normalizePath(url)

  if (normalizedMethod === 'get' && path === '/settings/public') {
    return { data: publicSettings }
  }

  if (normalizedMethod === 'get' && path === '/auth/me') {
    return { data: adminUser }
  }

  if (normalizedMethod === 'post' && path === '/auth/logout') {
    return { data: { message: 'ok' } }
  }

  if (normalizedMethod === 'get' && path === '/admin/settings') {
    return {
      data: {
        ops_monitoring_enabled: true,
        ops_realtime_monitoring_enabled: true,
        ops_query_mode_default: 'auto',
        custom_menu_items: []
      }
    }
  }

  if (normalizedMethod === 'get' && path === '/admin/payment/config') {
    return {
      data: {
        enabled: false,
        min_amount: 0,
        max_amount: 0,
        daily_limit: 0,
        order_timeout_minutes: 30,
        max_pending_orders: 0,
        enabled_payment_types: [],
        balance_disabled: true,
        balance_recharge_multiplier: 1,
        load_balance_strategy: 'round_robin',
        product_name_prefix: '',
        product_name_suffix: '',
        help_image_url: '',
        help_text: ''
      }
    }
  }

  if (normalizedMethod === 'get' && path === '/admin/proxies/all') {
    return { data: proxies }
  }

  if (normalizedMethod === 'get' && path === '/admin/groups/all') {
    return { data: groups }
  }

  if (normalizedMethod === 'get' && path === '/admin/accounts') {
    return {
      data: paginateAccounts(params),
      headers: {
        etag: 'W/"dev-preview-accounts-v1"'
      }
    }
  }

  if (normalizedMethod === 'get' && path === '/admin/accounts/openai-free-pools/config') {
    return { data: openAIFreePoolConfig }
  }

  if (normalizedMethod === 'get' && path === '/admin/accounts/openai-free-pools/preview') {
    return { data: openAIFreePoolPreview }
  }

  if (normalizedMethod === 'get' && path === '/admin/accounts/openai-free-pools/accounts') {
    return { data: openAIFreePoolManagedAccounts }
  }

  if (normalizedMethod === 'post' && path === '/admin/accounts/openai-free-pools/apply') {
    return { data: openAIFreePoolApplyResult }
  }

  if (normalizedMethod === 'post' && path === '/admin/accounts/openai-free-pools/locks') {
    return {
      data: {
        message: 'ok'
      }
    }
  }

  if (normalizedMethod === 'delete' && path.startsWith('/admin/accounts/openai-free-pools/locks/')) {
    return {
      data: {
        message: 'ok'
      }
    }
  }

  if (normalizedMethod === 'get' && path === '/admin/dashboard/snapshot-v2') {
    return { data: buildSnapshot(params) }
  }

  if (normalizedMethod === 'get' && path === '/admin/dashboard/users-trend') {
    return {
      data: {
        trend: usersTrend,
        start_date: String(params?.start_date ?? '2026-05-21'),
        end_date: String(params?.end_date ?? '2026-05-22'),
        granularity: String(params?.granularity ?? 'hour')
      }
    }
  }

  if (normalizedMethod === 'get' && path === '/admin/dashboard/users-ranking') {
    return { data: userRanking }
  }

  if (normalizedMethod === 'get' && path === '/admin/dashboard/openai-free-reset-forecast') {
    return { data: openAIFreeResetForecast }
  }

  if (normalizedMethod === 'get' && path === '/subscriptions/active') {
    return { data: activeSubscriptions }
  }

  if (normalizedMethod === 'get' && path === '/announcements') {
    return { data: userAnnouncements }
  }

  if (normalizedMethod === 'post' && path.startsWith('/announcements/') && path.endsWith('/read')) {
    return { data: { message: 'ok' } }
  }

  return null
}
