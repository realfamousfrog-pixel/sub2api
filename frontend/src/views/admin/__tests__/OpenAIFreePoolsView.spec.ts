import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import OpenAIFreePoolsView from '../OpenAIFreePoolsView.vue'

const {
  getOpenAIFreePoolConfig,
  previewOpenAIFreePool,
  getOpenAIFreePoolAccounts,
  getAllGroups,
  getAllProxies
} = vi.hoisted(() => ({
  getOpenAIFreePoolConfig: vi.fn(),
  previewOpenAIFreePool: vi.fn(),
  getOpenAIFreePoolAccounts: vi.fn(),
  getAllGroups: vi.fn(),
  getAllProxies: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      getOpenAIFreePoolConfig,
      previewOpenAIFreePool,
      getOpenAIFreePoolAccounts,
      applyOpenAIFreePool: vi.fn(),
      lockOpenAIFreePoolAccount: vi.fn(),
      unlockOpenAIFreePoolAccount: vi.fn()
    },
    groups: {
      getAll: getAllGroups
    },
    proxies: {
      getAll: getAllProxies
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showSuccess: vi.fn(),
    showError: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

describe('admin OpenAIFreePoolsView', () => {
  beforeEach(() => {
    getOpenAIFreePoolConfig.mockReset()
    previewOpenAIFreePool.mockReset()
    getOpenAIFreePoolAccounts.mockReset()
    getAllGroups.mockReset()
    getAllProxies.mockReset()

    getOpenAIFreePoolConfig.mockResolvedValue({
      enabled: true,
      default_group_id: 1,
      plus_group_id: 7,
      lookahead_days: 14,
      pools: [
        { group_id: 2, proxy_id: 101, label: 'free 秘鲁', sort_order: 10 },
        { group_id: 3, proxy_id: 102, label: 'free 巴西', sort_order: 20 }
      ]
    })
    previewOpenAIFreePool.mockResolvedValue({
      summary: {
        default_accounts: 1,
        unknown_reset_accounts: 0,
        managed_accounts: 3
      },
      pools: [],
      moves: [],
      generated_at: '2026-05-22T08:00:00Z'
    })
    getOpenAIFreePoolAccounts.mockResolvedValue([
      {
        account_id: 201,
        account_name: 'free 秘鲁',
        current_group_id: 2,
        current_group_name: 'free 秘鲁',
        current_proxy_name: '17899 秘鲁',
        reset_at: '2026-05-23T08:30:00Z',
        lock_mode: 'auto',
        lock_group_name: 'free 秘鲁',
        lock_proxy_name: '17899 秘鲁'
      },
      {
        account_id: 202,
        account_name: 'free 巴西',
        current_group_id: 3,
        current_group_name: 'free 巴西',
        current_proxy_name: '17900 巴西',
        reset_at: '2026-05-23T18:10:00Z',
        lock_mode: 'manual',
        lock_group_name: 'free 巴西',
        lock_proxy_name: '17900 巴西'
      },
      {
        account_id: 203,
        account_name: 'free intake 01',
        current_group_id: 1,
        current_group_name: 'default',
        current_proxy_name: '',
        reset_at: '2026-05-25T11:00:00Z',
        lock_mode: 'unlocked'
      }
    ])
    getAllGroups.mockResolvedValue([
      { id: 1, name: 'default' },
      { id: 2, name: 'free 秘鲁' },
      { id: 3, name: 'free 巴西' }
    ])
    getAllProxies.mockResolvedValue([
      { id: 101, name: '17899 秘鲁' },
      { id: 102, name: '17900 巴西' }
    ])
  })

  it('filters managed accounts by current group without extra requests', async () => {
    const wrapper = mount(OpenAIFreePoolsView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          OpenAIFreePoolLockModal: true,
          Select: {
            props: ['modelValue', 'options'],
            emits: ['update:modelValue'],
            template: `
              <select :value="modelValue" @change="$emit('update:modelValue', $event.target.value)">
                <option v-for="option in options" :key="option.value" :value="option.value">{{ option.label }}</option>
              </select>
            `
          }
        }
      }
    })

    await flushPromises()

    expect(getOpenAIFreePoolAccounts).toHaveBeenCalledTimes(1)
    expect(wrapper.text()).toContain('free 秘鲁')
    expect(wrapper.text()).toContain('free 巴西')
    expect(wrapper.text()).toContain('free intake 01')

    const filterSelect = wrapper.find('select')
    await filterSelect.setValue('default')
    expect(wrapper.text()).toContain('free intake 01')
    expect(wrapper.text()).not.toContain('17900 巴西')

    await filterSelect.setValue('group:2')
    expect(wrapper.text()).toContain('free 秘鲁')
    expect(wrapper.text()).not.toContain('free intake 01')

    expect(getOpenAIFreePoolAccounts).toHaveBeenCalledTimes(1)
  })
})
