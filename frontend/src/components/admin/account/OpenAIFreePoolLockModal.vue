<template>
  <Teleport to="body">
    <div v-if="show" class="fixed inset-0 z-[1000] flex items-center justify-center p-4">
      <div class="absolute inset-0 bg-black/50" @click="emit('close')"></div>
      <div class="relative z-[1001] w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl dark:bg-slate-900">
        <div class="flex items-start justify-between gap-4">
          <div>
            <h3 class="text-lg font-semibold text-slate-900 dark:text-white">
              {{ t('admin.openaiFreePools.lockToGroup') }}
            </h3>
            <p class="mt-1 text-sm text-slate-500 dark:text-slate-400">
              {{ account?.accountName || '-' }}
            </p>
          </div>
          <button
            type="button"
            class="rounded-full p-2 text-slate-500 transition-colors hover:bg-slate-100 hover:text-slate-900 dark:hover:bg-slate-800 dark:hover:text-white"
            @click="emit('close')"
          >
            <span class="sr-only">{{ t('common.close') }}</span>
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.5">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <div class="mt-5 space-y-4">
          <div class="grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm dark:border-slate-700 dark:bg-slate-800/60 md:grid-cols-2">
            <div>
              <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {{ t('admin.openaiFreePools.currentGroup') }}
              </div>
              <div class="mt-1 text-slate-900 dark:text-white">{{ account?.currentGroupName || '-' }}</div>
            </div>
            <div>
              <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {{ t('admin.openaiFreePools.currentProxy') }}
              </div>
              <div class="mt-1 text-slate-900 dark:text-white">{{ account?.currentProxyName || '-' }}</div>
            </div>
          </div>

          <label class="block">
            <span class="mb-2 block text-sm font-medium text-slate-700 dark:text-slate-200">
              {{ t('admin.openaiFreePools.targetGroup') }}
            </span>
            <select
              v-model="selectedGroupId"
              class="w-full rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 focus:border-primary-500 focus:outline-none focus:ring-2 focus:ring-primary-200 dark:border-slate-700 dark:bg-slate-950 dark:text-white dark:focus:ring-primary-900"
            >
              <option value="">{{ t('common.select') }}</option>
              <option v-for="option in poolOptions" :key="option.groupId" :value="String(option.groupId)">
                {{ option.groupName }} / {{ option.proxyName }}
              </option>
            </select>
          </label>

          <div class="rounded-xl border border-dashed border-slate-300 px-4 py-3 text-sm dark:border-slate-700">
            <div class="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">
              {{ t('admin.openaiFreePools.lockedProxy') }}
            </div>
            <div class="mt-1 text-slate-900 dark:text-white">
              {{ selectedPool?.proxyName || '-' }}
            </div>
          </div>
        </div>

        <div class="mt-6 flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" :disabled="submitting" @click="emit('close')">
            {{ t('common.cancel') }}
          </button>
          <button
            type="button"
            class="btn btn-primary"
            :disabled="submitting || !selectedPool || !account"
            @click="handleSubmit"
          >
            {{ submitting ? t('common.processing') : t('common.confirm') }}
          </button>
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

interface PoolOption {
  groupId: number
  groupName: string
  proxyId: number
  proxyName: string
}

interface LockModalAccount {
  accountId: number
  accountName: string
  currentGroupName?: string
  currentProxyName?: string
  lockGroupId?: number
}

const props = defineProps<{
  show: boolean
  account: LockModalAccount | null
  pools: PoolOption[]
  submitting?: boolean
}>()

const emit = defineEmits<{
  close: []
  submit: [{ account_id: number; target_group_id: number }]
}>()

const { t } = useI18n()
const selectedGroupId = ref('')

const poolOptions = computed(() => props.pools ?? [])
const selectedPool = computed(() => {
  const groupId = Number(selectedGroupId.value)
  if (!groupId) return null
  return poolOptions.value.find(pool => pool.groupId === groupId) ?? null
})

watch(
  () => [props.show, props.account?.lockGroupId, props.pools] as const,
  () => {
    if (!props.show) {
      selectedGroupId.value = ''
      return
    }
    const fallbackGroupId = props.account?.lockGroupId ?? poolOptions.value[0]?.groupId ?? 0
    selectedGroupId.value = fallbackGroupId ? String(fallbackGroupId) : ''
  },
  { immediate: true }
)

const handleSubmit = () => {
  if (!props.account || !selectedPool.value) return
  emit('submit', {
    account_id: props.account.accountId,
    target_group_id: selectedPool.value.groupId
  })
}
</script>
