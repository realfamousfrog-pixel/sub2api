<template>
  <AppLayout>
    <div class="space-y-6">
      <section class="rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900/70">
        <div class="flex flex-wrap items-start justify-between gap-4">
          <div>
            <h1 class="text-xl font-semibold text-slate-900 dark:text-white">
              {{ t('admin.openaiFreePools.title') }}
            </h1>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
              {{ t('admin.openaiFreePools.description') }}
            </p>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <label class="inline-flex items-center gap-2 rounded-full border border-slate-200 px-3 py-1.5 text-xs font-medium text-slate-600 dark:border-slate-700 dark:text-slate-300">
              <input
                v-model="forceRebalance"
                type="checkbox"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <span>{{ t('admin.openaiFreePools.forceRebalance') }}</span>
            </label>
            <button class="btn btn-secondary" :disabled="refreshing" @click="loadPageData(true)">
              {{ refreshing ? t('common.loading') : t('admin.openaiFreePools.preview') }}
            </button>
            <button class="btn btn-primary" :disabled="applying || !previewData" @click="handleApply">
              {{ applying ? t('common.processing') : t('admin.openaiFreePools.apply') }}
            </button>
          </div>
        </div>
      </section>

      <section class="rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900/70">
        <div class="flex items-center justify-between gap-4">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-white">
            {{ t('admin.openaiFreePools.statusSummary') }}
          </h2>
          <span v-if="loading" class="text-xs text-slate-500 dark:text-slate-400">{{ t('common.loading') }}</span>
        </div>
        <div v-if="errorMessage" class="mt-4 rounded-xl border border-red-200 bg-red-50 px-3 py-2 text-sm text-red-700 dark:border-red-800/60 dark:bg-red-900/20 dark:text-red-200">
          {{ errorMessage }}
        </div>
        <div class="mt-4 grid gap-3 md:grid-cols-2 xl:grid-cols-3">
          <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-800/60">
            <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {{ t('admin.openaiFreePools.configSummary') }}
            </div>
            <div class="mt-2 space-y-1 text-sm text-slate-700 dark:text-slate-200">
              <div>{{ t('admin.openaiFreePools.defaultGroup') }}: {{ defaultGroupName }}</div>
              <div>{{ t('admin.openaiFreePools.plusGroup') }}: {{ plusGroupName }}</div>
              <div>{{ t('admin.openaiFreePools.lookaheadDays') }}: {{ configData?.lookahead_days ?? 14 }}</div>
            </div>
          </div>
          <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4 dark:border-slate-700 dark:bg-slate-800/60">
            <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {{ t('admin.openaiFreePools.managedSummary') }}
            </div>
            <div class="mt-2 space-y-1 text-sm text-slate-700 dark:text-slate-200">
              <div>{{ t('admin.openaiFreePools.defaultAccounts') }}: {{ previewData?.summary.default_accounts ?? 0 }}</div>
              <div>{{ t('admin.openaiFreePools.unknownAccounts') }}: {{ previewData?.summary.unknown_reset_accounts ?? 0 }}</div>
              <div>{{ t('admin.openaiFreePools.managedAccounts') }}: {{ previewData?.summary.managed_accounts ?? 0 }}</div>
            </div>
          </div>
          <div
            v-for="pool in previewData?.pools ?? []"
            :key="pool.group_id"
            class="rounded-2xl border border-slate-200 bg-white p-4 dark:border-slate-700 dark:bg-slate-950/60"
          >
            <div class="flex items-center justify-between gap-2">
              <div class="text-sm font-semibold text-slate-900 dark:text-white">{{ pool.group_name }}</div>
              <span class="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-medium text-slate-700 dark:bg-slate-800 dark:text-slate-200">
                {{ pool.accounts }}
              </span>
            </div>
            <div class="mt-1 text-xs text-slate-500 dark:text-slate-400">{{ pool.proxy_name || '-' }}</div>
          </div>
        </div>
      </section>

      <section class="rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900/70">
        <div class="flex items-center justify-between gap-4">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-white">
            {{ t('admin.openaiFreePools.suggestions') }}
          </h2>
          <div class="text-right text-xs text-slate-500 dark:text-slate-400">
            <div>{{ t('admin.openaiFreePools.generatedAt') }}: {{ previewGeneratedAt }}</div>
            <div>{{ t('admin.openaiFreePools.moveCount', { count: previewData?.moves?.length ?? 0 }) }}</div>
          </div>
        </div>
        <div class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
            <thead class="bg-slate-50 dark:bg-slate-800/70">
              <tr>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.account') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.currentGroup') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.targetGroup') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.currentProxy') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.targetProxy') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.nextReset') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.reason') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.lockMode') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-if="!(previewData?.moves?.length)">
                <td colspan="8" class="px-3 py-6 text-center text-sm text-slate-500 dark:text-slate-400">
                  {{ t('admin.openaiFreePools.noSuggestions') }}
                </td>
              </tr>
              <tr v-for="move in previewData?.moves ?? []" :key="move.account_id">
                <td class="px-3 py-2 text-slate-900 dark:text-white">{{ move.account_name }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ move.current_group_name || '-' }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ move.target_group_name }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ move.current_proxy_name || '-' }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ move.target_proxy_name }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ formatResetAt(move.reset_at) }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ reasonLabel(move.reason) }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">
                  {{ move.locked ? t('admin.openaiFreePools.lockModeLocked') : t('admin.openaiFreePools.lockModeUnlocked') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <section class="rounded-3xl border border-slate-200 bg-white/90 p-6 shadow-sm dark:border-slate-700 dark:bg-slate-900/70">
        <div class="flex items-center justify-between gap-4">
          <h2 class="text-sm font-semibold text-slate-900 dark:text-white">
            {{ t('admin.openaiFreePools.lockManagement') }}
          </h2>
          <div class="flex items-center gap-3">
            <div class="w-44">
              <Select
                v-model="managedAccountFilter"
                :options="managedAccountFilterOptions"
              />
            </div>
            <span class="text-xs text-slate-500 dark:text-slate-400">
              {{ t('admin.openaiFreePools.managedAccounts') }}: {{ filteredManagedAccounts.length }}
            </span>
          </div>
        </div>
        <div class="mt-4 overflow-x-auto">
          <table class="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-700">
            <thead class="bg-slate-50 dark:bg-slate-800/70">
              <tr>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.account') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.currentGroup') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.currentProxy') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.nextReset') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.lockMode') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('admin.openaiFreePools.lockTarget') }}</th>
                <th class="px-3 py-2 text-left font-medium text-slate-500 dark:text-slate-300">{{ t('common.actions') }}</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-800">
              <tr v-if="!filteredManagedAccounts.length">
                <td colspan="7" class="px-3 py-6 text-center text-sm text-slate-500 dark:text-slate-400">
                  {{ t('admin.openaiFreePools.noManagedAccounts') }}
                </td>
              </tr>
              <tr v-for="account in filteredManagedAccounts" :key="account.account_id">
                <td class="px-3 py-2 text-slate-900 dark:text-white">{{ account.account_name }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ account.current_group_name || '-' }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ account.current_proxy_name || '-' }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ formatResetAt(account.reset_at) }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ lockModeLabel(account.lock_mode) }}</td>
                <td class="px-3 py-2 text-slate-600 dark:text-slate-300">{{ lockTargetLabel(account) }}</td>
                <td class="px-3 py-2">
                  <div class="flex flex-wrap gap-2">
                    <button class="btn btn-secondary btn-sm" :disabled="submittingLock" @click="openLockModal(account)">
                      {{ t('admin.openaiFreePools.lockToGroup') }}
                    </button>
                    <button
                      class="btn btn-secondary btn-sm"
                      :disabled="submittingLock || account.lock_mode !== 'manual'"
                      @click="handleUnlock(account)"
                    >
                      {{ t('admin.openaiFreePools.unlockManual') }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>
    </div>

    <OpenAIFreePoolLockModal
      :show="showLockModal"
      :account="lockModalAccount"
      :pools="lockPoolOptions"
      :submitting="submittingLock"
      @close="showLockModal = false"
      @submit="handleLockSubmit"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import AppLayout from '@/components/layout/AppLayout.vue'
import OpenAIFreePoolLockModal from '@/components/admin/account/OpenAIFreePoolLockModal.vue'
import Select from '@/components/common/Select.vue'
import { adminAPI } from '@/api/admin'
import { useAppStore } from '@/stores/app'
import type {
  AdminGroup,
  OpenAIFreePoolConfig,
  OpenAIFreePoolManagedAccount,
  OpenAIFreePoolPreview,
  Proxy
} from '@/types'

const { t } = useI18n()
const appStore = useAppStore()

const groups = ref<AdminGroup[]>([])
const proxies = ref<Proxy[]>([])
const configData = ref<OpenAIFreePoolConfig | null>(null)
const previewData = ref<OpenAIFreePoolPreview | null>(null)
const managedAccounts = ref<OpenAIFreePoolManagedAccount[]>([])
const managedAccountFilter = ref('all')
const forceRebalance = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const applying = ref(false)
const submittingLock = ref(false)
const errorMessage = ref('')
const showLockModal = ref(false)
const lockModalAccount = ref<{
  accountId: number
  accountName: string
  currentGroupName?: string
  currentProxyName?: string
  lockGroupId?: number
} | null>(null)

const groupNameById = computed(() => {
  const index = new Map<number, string>()
  groups.value.forEach(group => index.set(group.id, group.name))
  return index
})

const proxyNameById = computed(() => {
  const index = new Map<number, string>()
  proxies.value.forEach(proxy => index.set(proxy.id, proxy.name))
  return index
})

const defaultGroupName = computed(() => {
  const id = configData.value?.default_group_id
  return id ? (groupNameById.value.get(id) ?? String(id)) : '-'
})

const plusGroupName = computed(() => {
  const id = configData.value?.plus_group_id
  return id ? (groupNameById.value.get(id) ?? String(id)) : '-'
})

const previewGeneratedAt = computed(() => {
  if (!previewData.value?.generated_at) return '-'
  return new Date(previewData.value.generated_at).toLocaleString()
})

const lockPoolOptions = computed(() => {
  return (configData.value?.pools ?? []).map(pool => ({
    groupId: pool.group_id,
    groupName: groupNameById.value.get(pool.group_id) ?? pool.label ?? String(pool.group_id),
    proxyId: pool.proxy_id,
    proxyName: proxyNameById.value.get(pool.proxy_id) ?? String(pool.proxy_id)
  }))
})

const managedAccountFilterOptions = computed(() => {
  const options = [
    {
      value: 'all',
      label: t('admin.openaiFreePools.filters.all')
    }
  ]

  const defaultGroupId = configData.value?.default_group_id
  if (defaultGroupId) {
    options.push({
      value: 'default',
      label: defaultGroupName.value
    })
  }

  for (const pool of configData.value?.pools ?? []) {
    options.push({
      value: `group:${pool.group_id}`,
      label: groupNameById.value.get(pool.group_id) ?? pool.label ?? String(pool.group_id)
    })
  }

  return options
})

const filteredManagedAccounts = computed(() => {
  const filter = managedAccountFilter.value
  if (filter === 'all') {
    return managedAccounts.value
  }

  if (filter === 'default') {
    const defaultGroupId = configData.value?.default_group_id
    return managedAccounts.value.filter(account => account.current_group_id === defaultGroupId)
  }

  if (filter.startsWith('group:')) {
    const groupId = Number(filter.slice('group:'.length))
    return managedAccounts.value.filter(account => account.current_group_id === groupId)
  }

  return managedAccounts.value
})

const formatResetAt = (value?: string | null) => {
  if (!value) return t('admin.accounts.openaiFreePool.unknown')
  return new Date(value).toLocaleString()
}

const reasonLabel = (reason: string) => {
  const key = `admin.accounts.openaiFreePool.reasons.${reason}`
  const translated = t(key)
  return translated === key ? reason : translated
}

const lockModeLabel = (mode: string) => {
  if (mode === 'manual') return t('admin.openaiFreePools.lockModeManual')
  if (mode === 'auto') return t('admin.openaiFreePools.lockModeAuto')
  return t('admin.openaiFreePools.lockModeUnlocked')
}

const lockTargetLabel = (account: OpenAIFreePoolManagedAccount) => {
  const parts = [account.lock_group_name, account.lock_proxy_name].filter(Boolean)
  return parts.length ? parts.join(' / ') : '-'
}

const loadPageData = async (manual: boolean = false) => {
  if (manual) refreshing.value = true
  else loading.value = true
  errorMessage.value = ''
  try {
    const [config, preview, accounts, allGroups, allProxies] = await Promise.all([
      adminAPI.accounts.getOpenAIFreePoolConfig(),
      adminAPI.accounts.previewOpenAIFreePool(forceRebalance.value),
      adminAPI.accounts.getOpenAIFreePoolAccounts(),
      adminAPI.groups.getAll(),
      adminAPI.proxies.getAll()
    ])
    configData.value = config
    previewData.value = preview
    managedAccounts.value = accounts
    groups.value = allGroups
    proxies.value = allProxies
  } catch (error: any) {
    console.error('Failed to load OpenAI free pools view:', error)
    errorMessage.value = error?.message || t('admin.openaiFreePools.loadFailed')
  } finally {
    loading.value = false
    refreshing.value = false
  }
}

const refreshManagedSections = async () => {
  await loadPageData(true)
}

const handleApply = async () => {
  if (!previewData.value) return
  const confirmed = window.confirm(
    t('admin.openaiFreePools.applyConfirm', { count: previewData.value.moves.length })
  )
  if (!confirmed) return
  applying.value = true
  try {
    const result = await adminAPI.accounts.applyOpenAIFreePool(forceRebalance.value)
    appStore.showSuccess(
      t('admin.openaiFreePools.applySuccess', {
        applied: result.applied,
        failed: result.failed
      })
    )
    await refreshManagedSections()
  } catch (error: any) {
    console.error('Failed to apply OpenAI free pool preview:', error)
    appStore.showError(error?.message || t('admin.openaiFreePools.applyFailed'))
  } finally {
    applying.value = false
  }
}

const openLockModal = (account: OpenAIFreePoolManagedAccount) => {
  lockModalAccount.value = {
    accountId: account.account_id,
    accountName: account.account_name,
    currentGroupName: account.current_group_name,
    currentProxyName: account.current_proxy_name,
    lockGroupId: account.lock_group_id
  }
  showLockModal.value = true
}

const handleLockSubmit = async (payload: { account_id: number; target_group_id: number }) => {
  submittingLock.value = true
  try {
    await adminAPI.accounts.lockOpenAIFreePoolAccount(payload)
    appStore.showSuccess(t('admin.openaiFreePools.lockSuccess'))
    showLockModal.value = false
    await refreshManagedSections()
  } catch (error: any) {
    console.error('Failed to lock OpenAI free pool account:', error)
    appStore.showError(error?.message || t('admin.openaiFreePools.lockFailed'))
  } finally {
    submittingLock.value = false
  }
}

const handleUnlock = async (account: OpenAIFreePoolManagedAccount) => {
  submittingLock.value = true
  try {
    await adminAPI.accounts.unlockOpenAIFreePoolAccount(account.account_id)
    appStore.showSuccess(t('admin.openaiFreePools.unlockSuccess'))
    await refreshManagedSections()
  } catch (error: any) {
    console.error('Failed to unlock OpenAI free pool account:', error)
    appStore.showError(error?.message || t('admin.openaiFreePools.unlockFailed'))
  } finally {
    submittingLock.value = false
  }
}

onMounted(() => {
  loadPageData()
})
</script>
