<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, toRaw } from 'vue'
import NoticeBoard from '../components/NoticeBoard.vue'
import ShoppingList from '../components/ShoppingList.vue'
import GarbagePanel from '../components/GarbagePanel.vue'
import {
  APIError,
  createNote,
  deleteNote,
  fetchDashboard,
  type DashboardResponse,
  type Note,
  type NoteKind,
  updateNoteDone,
  updateNotePin
} from '../api/client'

const dashboard = ref<DashboardResponse | null>(null)
const loading = ref(true)
const fatalError = ref('')
const refreshFailed = ref(false)
const operationError = ref('')
const lastUpdatedAt = ref<Date | null>(null)
const pendingIDs = ref<number[]>([])
const isOffline = ref(!navigator.onLine)
const isLoadingDashboard = ref(false)

let pollTimer: number | undefined
let tempIDSeed = -1

function dateDesc(a: string, b: string): number {
  return Date.parse(b) - Date.parse(a)
}

function sortNotice(items: Note[]): Note[] {
  return items.sort((a, b) => {
    if (a.pinned !== b.pinned) {
      return Number(b.pinned) - Number(a.pinned)
    }
    return dateDesc(a.created_at, b.created_at)
  })
}

function sortShopping(items: Note[]): Note[] {
  return items.sort((a, b) => {
    if (a.done !== b.done) {
      return Number(a.done) - Number(b.done)
    }
    return dateDesc(a.created_at, b.created_at)
  })
}

function normalizeNotesOrder(state: DashboardResponse): void {
  sortNotice(state.notes.notice)
  sortShopping(state.notes.shopping)
}

function replacePendingState(id: number, active: boolean): void {
  if (active) {
    if (!pendingIDs.value.includes(id)) {
      pendingIDs.value = [...pendingIDs.value, id]
    }
    return
  }
  pendingIDs.value = pendingIDs.value.filter((v) => v !== id)
}

function isPending(id: number): boolean {
  return pendingIDs.value.includes(id)
}

function getOperationErrorMessage(err: unknown): string {
  if (err instanceof APIError) {
    if (err.code === 'NOT_FOUND') {
      return '対象が見つかりません。画面を再取得して再度お試しください。'
    }
    if (err.code === 'VALIDATION_ERROR') {
      return '入力内容を確認してください。'
    }
  }
  return '操作に失敗しました。しばらくしてから再度お試しください。'
}

async function loadDashboard(): Promise<void> {
  if (isOffline.value || isLoadingDashboard.value) {
    return
  }

  isLoadingDashboard.value = true
  try {
    const data = await fetchDashboard()
    normalizeNotesOrder(data)
    dashboard.value = data
    refreshFailed.value = false
    fatalError.value = ''
    operationError.value = ''
    lastUpdatedAt.value = new Date()
  } catch (err) {
    console.error(err)
    refreshFailed.value = true
    if (!dashboard.value) {
      fatalError.value = 'ダッシュボードの取得に失敗しました'
    }
  } finally {
    loading.value = false
    isLoadingDashboard.value = false
  }
}

async function withOptimisticUpdate(apply: (state: DashboardResponse) => void, commit: () => Promise<void>): Promise<void> {
  const state = dashboard.value
  if (!state) {
    throw new Error('dashboard state not loaded')
  }

  operationError.value = ''
  const snapshot = structuredClone(toRaw(state)) as DashboardResponse
  apply(state)

  try {
    await commit()
    refreshFailed.value = false
    lastUpdatedAt.value = new Date()
  } catch (err) {
    console.error(err)
    dashboard.value = snapshot
    refreshFailed.value = true
    const message = getOperationErrorMessage(err)
    operationError.value = message
    throw new Error(message)
  }
}

async function handleAddNote(kind: NoteKind, body: string): Promise<void> {
  if (!dashboard.value) {
    throw new Error('dashboard state not loaded')
  }

  const nowISO = new Date().toISOString()
  const tempID = tempIDSeed
  tempIDSeed -= 1
  const tempNote: Note = {
    id: tempID,
    kind,
    body,
    pinned: false,
    done: false,
    created_at: nowISO,
    updated_at: nowISO
  }

  await withOptimisticUpdate(
    (state) => {
      if (kind === 'notice') {
        state.notes.notice = sortNotice([tempNote, ...state.notes.notice])
      } else {
        state.notes.shopping = sortShopping([tempNote, ...state.notes.shopping])
      }
    },
    async () => {
      const created = await createNote(kind, body)
      const latest = dashboard.value
      if (!latest) {
        return
      }
      const target = kind === 'notice' ? latest.notes.notice : latest.notes.shopping
      const idx = target.findIndex((note) => note.id === tempID)
      if (idx >= 0) {
        target[idx] = created
      }
      if (kind === 'notice') {
        sortNotice(target)
      } else {
        sortShopping(target)
      }
    }
  )
}

async function handleTogglePin(note: Note): Promise<void> {
  if (note.kind !== 'notice' || isPending(note.id)) {
    return
  }

  const nextPinned = !note.pinned
  replacePendingState(note.id, true)
  try {
    await withOptimisticUpdate(
      (state) => {
        const target = state.notes.notice.find((item) => item.id === note.id)
        if (!target) {
          return
        }
        target.pinned = nextPinned
        sortNotice(state.notes.notice)
      },
      async () => {
        const updated = await updateNotePin(note.id, nextPinned)
        const target = dashboard.value?.notes.notice.find((item) => item.id === note.id)
        if (target) {
          target.pinned = updated.pinned
          target.updated_at = updated.updated_at
          sortNotice(dashboard.value!.notes.notice)
        }
      }
    )
  } finally {
    replacePendingState(note.id, false)
  }
}

async function handleToggleDone(note: Note): Promise<void> {
  if (note.kind !== 'shopping' || isPending(note.id)) {
    return
  }

  const nextDone = !note.done
  replacePendingState(note.id, true)
  try {
    await withOptimisticUpdate(
      (state) => {
        const target = state.notes.shopping.find((item) => item.id === note.id)
        if (!target) {
          return
        }
        target.done = nextDone
        sortShopping(state.notes.shopping)
      },
      async () => {
        const updated = await updateNoteDone(note.id, nextDone)
        const target = dashboard.value?.notes.shopping.find((item) => item.id === note.id)
        if (target) {
          target.done = updated.done
          target.updated_at = updated.updated_at
          sortShopping(dashboard.value!.notes.shopping)
        }
      }
    )
  } finally {
    replacePendingState(note.id, false)
  }
}

async function handleDelete(note: Note): Promise<void> {
  if (isPending(note.id)) {
    return
  }

  replacePendingState(note.id, true)
  try {
    await withOptimisticUpdate(
      (state) => {
        if (note.kind === 'notice') {
          state.notes.notice = state.notes.notice.filter((item) => item.id !== note.id)
        } else {
          state.notes.shopping = state.notes.shopping.filter((item) => item.id !== note.id)
        }
      },
      async () => {
        await deleteNote(note.id)
      }
    )
  } finally {
    replacePendingState(note.id, false)
  }
}

const lastUpdatedLabel = computed(() => {
  if (!lastUpdatedAt.value) {
    return '--:--'
  }
  return new Intl.DateTimeFormat('ja-JP', {
    hour: '2-digit',
    minute: '2-digit',
    hour12: false,
    timeZone: 'Asia/Tokyo'
  }).format(lastUpdatedAt.value)
})

function startPolling(): void {
  if (pollTimer !== undefined || isOffline.value) {
    return
  }
  pollTimer = window.setInterval(() => {
    void loadDashboard()
  }, 30_000)
}

function stopPolling(): void {
  if (pollTimer === undefined) {
    return
  }
  window.clearInterval(pollTimer)
  pollTimer = undefined
}

function handleVisibilityChange(): void {
  if (document.visibilityState === 'visible' && !isOffline.value) {
    void loadDashboard()
  }
}

function handleOffline(): void {
  isOffline.value = true
  stopPolling()
}

function handleOnline(): void {
  isOffline.value = false
  startPolling()
  void loadDashboard()
}

onMounted(async () => {
  window.addEventListener('offline', handleOffline)
  window.addEventListener('online', handleOnline)
  document.addEventListener('visibilitychange', handleVisibilityChange)

  if (isOffline.value) {
    loading.value = false
    stopPolling()
    return
  }

  await loadDashboard()
  startPolling()
})

onUnmounted(() => {
  stopPolling()
  window.removeEventListener('offline', handleOffline)
  window.removeEventListener('online', handleOnline)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
})
</script>

<template>
  <main class="page">
    <header class="header">
      <h1>HomeDash</h1>
      <div class="status">
        <span>更新: {{ lastUpdatedLabel }}</span>
        <span v-if="refreshFailed" class="status-error">更新失敗</span>
        <span v-if="isOffline" class="status-offline">オフライン</span>
      </div>
    </header>

    <p v-if="loading">読み込み中...</p>
    <p v-else-if="!dashboard && fatalError">{{ fatalError }}</p>

    <p v-if="operationError" class="operation-error">{{ operationError }}</p>

    <div v-if="dashboard" class="grid">
      <NoticeBoard
        :items="dashboard.notes.notice"
        :pending-ids="pendingIDs"
        :on-add="(body) => handleAddNote('notice', body)"
        :on-toggle-pin="handleTogglePin"
        :on-delete-note="handleDelete"
      />
      <ShoppingList
        :items="dashboard.notes.shopping"
        :pending-ids="pendingIDs"
        :on-add="(body) => handleAddNote('shopping', body)"
        :on-toggle-done="handleToggleDone"
        :on-delete-note="handleDelete"
      />
      <GarbagePanel :today="dashboard.garbage.today" :tomorrow="dashboard.garbage.tomorrow" />
    </div>
  </main>
</template>

<style scoped>
.page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px;
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-end;
  gap: 12px;
  margin-bottom: 12px;
}

h1 {
  margin: 0;
  font-size: 1.8rem;
}

.status {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #666;
  font-size: 0.86rem;
}

.status-error {
  color: #b00020;
  font-weight: 700;
}

.status-offline {
  color: #0b5fa8;
  font-weight: 700;
}

.operation-error {
  margin: 0 0 12px;
  padding: 8px 10px;
  border-radius: 8px;
  background: #ffe8e8;
  color: #7a0000;
  font-size: 0.9rem;
}

.grid {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

@media (max-width: 900px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
