<script setup lang="ts">
import { nextTick, ref } from 'vue'
import type { Note } from '../api/client'

const noticeAuthors = ['夫', '妻'] as const
const noticeAuthorStorageKey = 'home-dash.notice-author'

const props = defineProps<{
  items: Note[]
  pendingIds: number[]
  onAdd: (body: string, author: string) => Promise<void>
  onTogglePin: (note: Note) => Promise<void>
  onToggleAcknowledged: (note: Note) => Promise<void>
  onDeleteNote: (note: Note) => Promise<void>
}>()

const body = ref('')
const selectedAuthor = ref(loadInitialAuthor())
const validationError = ref('')
const actionError = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const deleteTarget = ref<Note | null>(null)

function loadInitialAuthor(): string {
  if (typeof window === 'undefined') {
    return noticeAuthors[0]
  }

  const stored = window.localStorage.getItem(noticeAuthorStorageKey)
  if (stored && noticeAuthors.includes(stored as (typeof noticeAuthors)[number])) {
    return stored
  }
  return noticeAuthors[0]
}

function saveSelectedAuthor(): void {
  if (typeof window === 'undefined') {
    return
  }
  window.localStorage.setItem(noticeAuthorStorageKey, selectedAuthor.value)
}

function validate(raw: string): string {
  const normalized = raw.trim()
  if (normalized === '') {
    return '本文を入力してください'
  }
  if ([...normalized].length > 200) {
    return '200文字以内で入力してください'
  }
  return ''
}

function isPending(id: number): boolean {
  return props.pendingIds.includes(id)
}

function authorLabel(note: Note): string {
  return note.author.trim() === '' ? '投稿者未設定' : note.author
}

function selectAuthor(author: string): void {
  selectedAuthor.value = author
  saveSelectedAuthor()
}

async function submit(): Promise<void> {
  validationError.value = validate(body.value)
  actionError.value = ''
  if (validationError.value !== '') {
    return
  }

  try {
    await props.onAdd(body.value.trim(), selectedAuthor.value)
    body.value = ''
    await nextTick()
    inputRef.value?.focus()
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '追加に失敗しました'
  }
}

async function onTogglePin(note: Note): Promise<void> {
  actionError.value = ''
  try {
    await props.onTogglePin(note)
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '更新に失敗しました'
  }
}

async function onToggleAcknowledged(note: Note): Promise<void> {
  actionError.value = ''
  try {
    await props.onToggleAcknowledged(note)
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '更新に失敗しました'
  }
}

function requestDelete(note: Note): void {
  if (isPending(note.id)) {
    return
  }
  deleteTarget.value = note
}

function cancelDelete(): void {
  deleteTarget.value = null
}

async function confirmDelete(): Promise<void> {
  if (!deleteTarget.value) {
    return
  }
  const target = deleteTarget.value
  deleteTarget.value = null
  actionError.value = ''
  try {
    await props.onDeleteNote(target)
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '削除に失敗しました'
  }
}
</script>

<template>
  <section class="hd-panel notice-board">
    <h2 class="hd-panel-title">連絡</h2>

    <form class="hd-composer" @submit.prevent="submit">
      <div class="notice-author-field">
        <span class="notice-author-label">投稿者</span>
        <div class="notice-author-toggle" role="radiogroup" aria-label="投稿者">
          <button
            v-for="author in noticeAuthors"
            :key="author"
            type="button"
            :class="['hd-btn', 'hd-btn-small', 'notice-author-option', { 'is-selected': selectedAuthor === author }]"
            :aria-pressed="selectedAuthor === author"
            @click="selectAuthor(author)"
          >
            {{ author }}
          </button>
        </div>
      </div>
      <input
        ref="inputRef"
        v-model="body"
        class="hd-input-text"
        type="text"
        placeholder="連絡を入力"
        maxlength="200"
        autocomplete="off"
      />
      <button type="submit" class="hd-btn hd-btn-primary">追加</button>
    </form>
    <p v-if="validationError" class="hd-error">{{ validationError }}</p>
    <p v-if="actionError" class="hd-error">{{ actionError }}</p>

    <ul v-if="items.length > 0" class="hd-list">
      <li
        v-for="note in items"
        :key="note.id"
        :class="['hd-list-item', { 'is-pinned': note.pinned, 'is-acknowledged': note.acknowledged }]"
      >
        <div class="notice-content">
          <span class="notice-author-chip">{{ authorLabel(note) }}</span>
          <span class="hd-list-body">{{ note.body }}</span>
        </div>
        <div class="notice-actions">
          <button
            :class="['hd-btn', 'hd-btn-small', 'notice-icon-button', { 'is-active': note.acknowledged }]"
            type="button"
            :disabled="isPending(note.id)"
            :aria-label="note.acknowledged ? '確認解除' : '確認済みにする'"
            @click="onToggleAcknowledged(note)"
          >
            <span class="notice-icon-indicator" aria-hidden="true">
              {{ note.acknowledged ? '✓' : '' }}
            </span>
          </button>
          <button class="hd-btn hd-btn-small" type="button" :disabled="isPending(note.id)" @click="onTogglePin(note)">
            {{ note.pinned ? 'ピン解除' : 'ピン' }}
          </button>
          <button
            class="hd-btn hd-btn-small hd-btn-danger"
            type="button"
            :disabled="isPending(note.id)"
            @click="requestDelete(note)"
          >
            削除
          </button>
        </div>
      </li>
    </ul>
    <p v-else class="hd-empty">連絡はありません</p>

    <div v-if="deleteTarget" class="hd-confirm-overlay" role="dialog" aria-modal="true" aria-label="削除確認">
      <div class="hd-confirm-modal">
        <p>削除しますか？</p>
        <div class="hd-confirm-actions">
          <button type="button" class="hd-btn hd-btn-small" @click="cancelDelete">キャンセル</button>
          <button type="button" class="hd-btn hd-btn-small hd-btn-danger" @click="confirmDelete">削除</button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.notice-board .hd-composer {
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: end;
  gap: 10px;
}

.is-pinned {
  border-color: #ead59a;
  background: #fff4d2;
  box-shadow: inset 4px 0 0 #e6bb40;
}

.is-acknowledged {
  opacity: 0.82;
}

.notice-board .hd-list-item {
  align-items: flex-start;
}

.notice-author-field {
  display: grid;
  grid-column: 1 / -1;
  gap: 6px;
}

.notice-author-label {
  color: #5c6886;
  font-size: 0.78rem;
  font-weight: 700;
}

.notice-author-toggle {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.notice-author-option {
  min-width: 64px;
  border-color: #cad5ea;
  background: rgba(255, 255, 255, 0.92);
}

.notice-author-option.is-selected {
  border-color: #2559c0;
  background: #2f6ce5;
  color: #ffffff;
  box-shadow: 0 8px 16px rgba(47, 108, 229, 0.18);
}

.notice-content {
  flex: 1;
  min-width: 0;
  display: grid;
  gap: 8px;
}

.notice-board .hd-list-body {
  overflow-wrap: break-word;
  word-break: normal;
}

.notice-actions {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  flex-shrink: 0;
}

.notice-icon-button {
  min-width: 40px;
  padding: 0;
  background: rgba(255, 255, 255, 0.92);
}

.notice-icon-button.is-active {
  border-color: #2559c0;
  background: #2f6ce5;
  color: #ffffff;
}

.notice-icon-indicator {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 18px;
  height: 18px;
  border: 2px solid currentColor;
  border-radius: 999px;
  font-size: 0.82rem;
  font-weight: 800;
  line-height: 1;
}

.notice-author-chip {
  display: inline-flex;
  align-items: center;
  justify-self: start;
  min-height: 28px;
  padding: 0 10px;
  border-radius: 999px;
  background: rgba(53, 102, 221, 0.1);
  color: #2b5fc1;
  font-size: 0.78rem;
  font-weight: 700;
}

@media (max-width: 680px) {
  .notice-board .hd-composer {
    grid-template-columns: 1fr;
  }

  .notice-board .hd-composer .hd-btn-primary {
    width: 100%;
  }

  .notice-actions {
    width: 100%;
    justify-content: flex-end;
  }
}
</style>
