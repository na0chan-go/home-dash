<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref, toRaw } from 'vue'
import NoticeBoard from '../components/NoticeBoard.vue'
import ShoppingList from '../components/ShoppingList.vue'
import GarbagePanel from '../components/GarbagePanel.vue'
import { createAsyncQueue } from '../utils/asyncQueue'
import {
  APIError,
  createNote,
  deleteNote,
  fetchDashboard,
  type DashboardResponse,
  type Note,
  type NoteKind,
  updateNoteAcknowledged,
  updateNoteDone,
  updateNotePin
} from '../api/client'

const TOKYO_TIMEZONE = 'Asia/Tokyo'
const DEFAULT_DIM_START = '22:00'
const DEFAULT_DIM_END = '06:00'
const DEFAULT_SCREENSAVER_IDLE_MS = 5 * 60 * 1000

const nightStartMinutes = parseHourMinute(import.meta.env.VITE_DIM_START ?? DEFAULT_DIM_START, 22 * 60)
const nightEndMinutes = parseHourMinute(import.meta.env.VITE_DIM_END ?? DEFAULT_DIM_END, 6 * 60)
const screenSaverIdleMs = parsePositiveInt(
  import.meta.env.VITE_SCREENSAVER_IDLE_MS ?? String(DEFAULT_SCREENSAVER_IDLE_MS),
  DEFAULT_SCREENSAVER_IDLE_MS
)

const tokyoHourMinuteFormatter = new Intl.DateTimeFormat('ja-JP', {
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
  timeZone: TOKYO_TIMEZONE
})

const tokyoTimeFormatter = new Intl.DateTimeFormat('ja-JP', {
  hour: '2-digit',
  minute: '2-digit',
  hour12: false,
  timeZone: TOKYO_TIMEZONE
})

const dashboard = ref<DashboardResponse | null>(null)
const loading = ref(true)
const fatalError = ref('')
const refreshFailed = ref(false)
const operationError = ref('')
const lastUpdatedAt = ref<Date | null>(null)
const pendingIDs = ref<number[]>([])
const isOffline = ref(!navigator.onLine)
const isLoadingDashboard = ref(false)
const isNightDim = ref(false)
const isScreenSaver = ref(false)
const screenSaverNow = ref(new Date())

let pollTimer: number | undefined
let nightTimer: number | undefined
let inactivityTimer: number | undefined
let screenSaverClockTimer: number | undefined
let tempIDSeed = -1
// Run note mutations one-by-one to avoid optimistic rollback collisions.
const enqueueNoteOperation = createAsyncQueue()

function parseHourMinute(raw: string, fallbackMinutes: number): number {
  const normalized = raw.trim()
  const match = normalized.match(/^([01]?\d|2[0-3]):([0-5]\d)$/)
  if (!match) {
    return fallbackMinutes
  }
  return Number(match[1]) * 60 + Number(match[2])
}

function parsePositiveInt(raw: string, fallback: number): number {
  const parsed = Number.parseInt(raw.trim(), 10)
  if (Number.isNaN(parsed) || parsed < 1) {
    return fallback
  }
  return parsed
}

function getTokyoMinutes(now: Date): number {
  const parts = tokyoHourMinuteFormatter.formatToParts(now)
  const hour = Number(parts.find((part) => part.type === 'hour')?.value ?? '0')
  const minute = Number(parts.find((part) => part.type === 'minute')?.value ?? '0')
  return hour * 60 + minute
}

function isWithinNightWindow(current: number, start: number, end: number): boolean {
  if (start === end) {
    return false
  }
  if (start < end) {
    return current >= start && current < end
  }
  return current >= start || current < end
}

function updateNightDim(): void {
  isNightDim.value = isWithinNightWindow(getTokyoMinutes(new Date()), nightStartMinutes, nightEndMinutes)
}

function startNightTimer(): void {
  if (nightTimer !== undefined) {
    return
  }
  nightTimer = window.setInterval(() => {
    updateNightDim()
  }, 60_000)
}

function stopNightTimer(): void {
  if (nightTimer === undefined) {
    return
  }
  window.clearInterval(nightTimer)
  nightTimer = undefined
}

function clearInactivityTimer(): void {
  if (inactivityTimer === undefined) {
    return
  }
  window.clearTimeout(inactivityTimer)
  inactivityTimer = undefined
}

function scheduleInactivityTimer(): void {
  clearInactivityTimer()
  if (isScreenSaver.value) {
    return
  }
  inactivityTimer = window.setTimeout(() => {
    enterScreenSaver()
  }, screenSaverIdleMs)
}

function startScreenSaverClock(): void {
  if (screenSaverClockTimer !== undefined) {
    return
  }
  screenSaverNow.value = new Date()
  screenSaverClockTimer = window.setInterval(() => {
    screenSaverNow.value = new Date()
  }, 1_000)
}

function stopScreenSaverClock(): void {
  if (screenSaverClockTimer === undefined) {
    return
  }
  window.clearInterval(screenSaverClockTimer)
  screenSaverClockTimer = undefined
}

function enterScreenSaver(): void {
  if (isScreenSaver.value || document.visibilityState !== 'visible') {
    return
  }
  isScreenSaver.value = true
  clearInactivityTimer()
  startScreenSaverClock()
}

async function exitScreenSaver(): Promise<void> {
  if (!isScreenSaver.value) {
    scheduleInactivityTimer()
    return
  }

  isScreenSaver.value = false
  stopScreenSaverClock()
  scheduleInactivityTimer()

  if (!isOffline.value && document.visibilityState === 'visible') {
    await loadDashboard()
  }
}

function handleScreenSaverDismiss(): void {
  void exitScreenSaver()
}

function handleUserActivity(): void {
  if (document.visibilityState !== 'visible') {
    return
  }
  if (isScreenSaver.value) {
    void exitScreenSaver()
    return
  }
  scheduleInactivityTimer()
}

function dateDesc(a: string, b: string): number {
  return Date.parse(b) - Date.parse(a)
}

function sortNotice(items: Note[]): Note[] {
  return items.sort((a, b) => {
    if (a.pinned !== b.pinned) {
      return Number(b.pinned) - Number(a.pinned)
    }
    if (a.acknowledged !== b.acknowledged) {
      return Number(a.acknowledged) - Number(b.acknowledged)
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
    if (err.status === 401 || err.code === 'UNAUTHORIZED') {
      return '認証が必要です。管理者にトークン設定を確認してください。'
    }
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
    if (err instanceof APIError && (err.status === 401 || err.code === 'UNAUTHORIZED')) {
      fatalError.value = '認証が必要です。Authorization: Bearer <token> を設定してください。'
      return
    }
    if (!dashboard.value) {
      fatalError.value = 'ダッシュボードの取得に失敗しました'
    }
  } finally {
    loading.value = false
    isLoadingDashboard.value = false
  }
}

function cloneDashboardState(state: DashboardResponse): DashboardResponse {
  // `structuredClone` can fail on iOS Safari when reactive proxy values are mixed in.
  // Dashboard DTO is JSON-safe, so we snapshot through JSON for stable optimistic rollback.
  return JSON.parse(JSON.stringify(state)) as DashboardResponse
}

async function withOptimisticUpdate(apply: (state: DashboardResponse) => void, commit: () => Promise<void>): Promise<void> {
  return enqueueNoteOperation(async () => {
    const state = dashboard.value
    if (!state) {
      throw new Error('dashboard state not loaded')
    }

    operationError.value = ''
    const snapshot = cloneDashboardState(toRaw(state) as DashboardResponse)
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
  })
}

async function handleAddNote(kind: NoteKind, body: string, author = ''): Promise<void> {
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
    author: kind === 'notice' ? author : '',
    pinned: false,
    acknowledged: false,
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
      const created = await createNote(kind, body, author)
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

async function handleToggleAcknowledged(note: Note): Promise<void> {
  if (note.kind !== 'notice' || isPending(note.id)) {
    return
  }

  const nextAcknowledged = !note.acknowledged
  replacePendingState(note.id, true)
  try {
    await withOptimisticUpdate(
      (state) => {
        const target = state.notes.notice.find((item) => item.id === note.id)
        if (!target) {
          return
        }
        target.acknowledged = nextAcknowledged
        sortNotice(state.notes.notice)
      },
      async () => {
        const updated = await updateNoteAcknowledged(note.id, nextAcknowledged)
        const target = dashboard.value?.notes.notice.find((item) => item.id === note.id)
        if (target) {
          target.acknowledged = updated.acknowledged
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
    timeZone: TOKYO_TIMEZONE
  }).format(lastUpdatedAt.value)
})

const screenSaverTimeLabel = computed(() => tokyoTimeFormatter.format(screenSaverNow.value))

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
  if (document.visibilityState === 'visible') {
    updateNightDim()
    if (!isOffline.value) {
      void loadDashboard()
    }
    if (!isScreenSaver.value) {
      scheduleInactivityTimer()
    }
    return
  }
  clearInactivityTimer()
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
  window.addEventListener('pointerdown', handleUserActivity)
  window.addEventListener('keydown', handleUserActivity)
  window.addEventListener('touchstart', handleUserActivity)

  updateNightDim()
  startNightTimer()

  if (isOffline.value) {
    loading.value = false
    stopPolling()
    scheduleInactivityTimer()
    return
  }

  await loadDashboard()
  startPolling()
  scheduleInactivityTimer()
})

onUnmounted(() => {
  stopPolling()
  stopNightTimer()
  clearInactivityTimer()
  stopScreenSaverClock()

  window.removeEventListener('offline', handleOffline)
  window.removeEventListener('online', handleOnline)
  document.removeEventListener('visibilitychange', handleVisibilityChange)
  window.removeEventListener('pointerdown', handleUserActivity)
  window.removeEventListener('keydown', handleUserActivity)
  window.removeEventListener('touchstart', handleUserActivity)
})
</script>

<template>
  <main class="page">
    <header class="header">
      <div class="brand">
        <h1 class="title">HomeDash</h1>
      </div>
      <div class="status">
        <span class="status-chip">更新: {{ lastUpdatedLabel }}</span>
        <span v-if="refreshFailed" class="status-chip status-error">更新失敗</span>
        <span v-if="isOffline" class="status-chip status-offline">オフライン</span>
      </div>
    </header>

    <p v-if="loading" class="page-message">読み込み中...</p>
    <p v-else-if="!dashboard && fatalError" class="page-message page-message-error">{{ fatalError }}</p>

    <p v-if="operationError" class="operation-error">{{ operationError }}</p>

    <div v-if="dashboard" class="grid">
      <NoticeBoard
        :items="dashboard.notes.notice"
        :pending-ids="pendingIDs"
        :on-add="(body, author) => handleAddNote('notice', body, author)"
        :on-toggle-pin="handleTogglePin"
        :on-toggle-acknowledged="handleToggleAcknowledged"
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

    <div v-if="isNightDim && !isScreenSaver" class="night-dim" aria-hidden="true" />

    <div
      v-if="isScreenSaver"
      class="screensaver"
      role="button"
      tabindex="0"
      aria-label="スクリーンセーバー。タップで解除"
      @pointerdown.prevent="handleScreenSaverDismiss"
      @keydown.enter.prevent="handleScreenSaverDismiss"
      @keydown.space.prevent="handleScreenSaverDismiss"
    >
      <div class="screensaver-time">{{ screenSaverTimeLabel }}</div>
    </div>
  </main>
</template>

<style scoped>
.page {
  position: relative;
  max-width: 1260px;
  margin: 0 auto;
  padding: 20px;
}

.page::before,
.page::after {
  content: '';
  position: absolute;
  border-radius: 999px;
  pointer-events: none;
  z-index: -1;
  filter: blur(42px);
}

.page::before {
  width: 320px;
  height: 180px;
  top: -36px;
  left: -40px;
  background: rgba(106, 146, 255, 0.24);
}

.page::after {
  width: 260px;
  height: 160px;
  top: -18px;
  right: 6px;
  background: rgba(255, 186, 133, 0.2);
}

.header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 14px;
}

.brand {
  display: block;
}

.title {
  margin: 0;
  font-size: clamp(2rem, 3.3vw, 2.6rem);
  letter-spacing: 0.01em;
}

.status {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
}

.status-chip {
  display: inline-flex;
  align-items: center;
  min-height: 32px;
  padding: 0 12px;
  border-radius: 999px;
  border: 1px solid #d7e1f3;
  background: rgba(255, 255, 255, 0.85);
  color: #4f5f7f;
  font-size: 0.84rem;
  font-weight: 700;
  box-shadow: 0 6px 16px rgba(23, 43, 81, 0.08);
}

.status-error {
  border-color: #efc2ca;
  background: #fff1f3;
  color: #9e2632;
}

.status-offline {
  border-color: #c5dbf6;
  background: #edf6ff;
  color: #1f5f95;
}

.page-message {
  margin: 0 0 12px;
  padding: 12px 14px;
  border-radius: 12px;
  border: 1px solid #d6dff0;
  background: rgba(255, 255, 255, 0.8);
  color: #51617f;
  font-size: 0.92rem;
  font-weight: 700;
}

.page-message-error {
  border-color: #efc2ca;
  background: #fff4f6;
  color: #8c1d28;
}

.operation-error {
  margin: 0 0 12px;
  padding: 11px 12px;
  border-radius: 12px;
  border: 1px solid #efc1c8;
  background: linear-gradient(180deg, #fff5f6 0%, #ffecee 100%);
  color: #8b1f2b;
  font-size: 0.9rem;
  font-weight: 700;
  box-shadow: 0 8px 18px rgba(164, 40, 51, 0.08);
}

.grid {
  display: grid;
  align-items: start;
  gap: 14px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

.night-dim {
  position: fixed;
  inset: 0;
  background: rgba(5, 9, 20, 0.38);
  pointer-events: none;
  z-index: 30;
}

.screensaver {
  position: fixed;
  inset: 0;
  z-index: 40;
  background: rgba(7, 10, 17, 0.96);
  display: grid;
  place-items: center;
  cursor: pointer;
  touch-action: manipulation;
}

.screensaver-time {
  color: #d8e7ff;
  font-size: clamp(2rem, 9vw, 4.8rem);
  letter-spacing: 0.08em;
  user-select: none;
  animation: drift 48s linear infinite alternate;
}

@keyframes drift {
  0% {
    transform: translate(-15vw, -10vh);
  }
  50% {
    transform: translate(12vw, 8vh);
  }
  100% {
    transform: translate(-8vw, 14vh);
  }
}

@media (max-width: 900px) {
  .header {
    flex-direction: column;
    align-items: flex-start;
  }

  .status {
    justify-content: flex-start;
  }

  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
