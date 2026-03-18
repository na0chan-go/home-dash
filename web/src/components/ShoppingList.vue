<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import type { Note } from '../api/client'

const fixedSuggestions = ['牛乳', '卵', 'パン', 'ティッシュ', 'トイレットペーパー'] as const

const props = defineProps<{
  items: Note[]
  pendingIds: number[]
  onAdd: (body: string) => Promise<void>
  onToggleDone: (note: Note) => Promise<void>
  onDeleteNote: (note: Note) => Promise<void>
}>()

const body = ref('')
const showDone = ref(false)
const validationError = ref('')
const actionError = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const deleteTarget = ref<Note | null>(null)

const displayedItems = computed(() => {
  if (showDone.value) {
    return props.items
  }
  return props.items.filter((note) => !note.done)
})

const recentSuggestions = computed(() => {
  const seen = new Set<string>()
  const recent = [...props.items]
    .sort((a, b) => Date.parse(b.created_at) - Date.parse(a.created_at))
    .map((note) => note.body.trim())
    .filter((body) => body !== '')
    .filter((body) => {
      if (seen.has(body)) {
        return false
      }
      seen.add(body)
      return true
    })

  return recent.slice(0, 5)
})

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

async function submit(): Promise<void> {
  await submitBody(body.value)
}

async function submitSuggestion(suggestion: string): Promise<void> {
  await submitBody(suggestion)
}

async function submitBody(raw: string): Promise<void> {
  actionError.value = ''
  validationError.value = validate(raw)
  if (validationError.value !== '') {
    return
  }

  try {
    await props.onAdd(raw.trim())
    body.value = ''
    await nextTick()
    inputRef.value?.focus()
  } catch (err) {
    actionError.value = err instanceof Error ? err.message : '追加に失敗しました'
  }
}

async function onToggleDone(note: Note): Promise<void> {
  actionError.value = ''
  try {
    await props.onToggleDone(note)
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
  <section class="hd-panel shopping-list">
    <h2 class="hd-panel-title">買い物</h2>

    <form class="hd-composer" @submit.prevent="submit">
      <input
        ref="inputRef"
        v-model="body"
        class="hd-input-text"
        type="text"
        placeholder="買い物メモを入力"
        maxlength="200"
        autocomplete="off"
      />
      <button type="submit" class="hd-btn hd-btn-primary">追加</button>
    </form>
    <p v-if="validationError" class="hd-error">{{ validationError }}</p>
    <p v-if="actionError" class="hd-error">{{ actionError }}</p>

    <div class="shopping-suggestions">
      <div class="shopping-suggestion-section">
        <p class="shopping-suggestion-title">よく使う項目</p>
        <div class="shopping-suggestion-list">
          <button
            v-for="suggestion in fixedSuggestions"
            :key="suggestion"
            type="button"
            class="shopping-suggestion-chip"
            @click="submitSuggestion(suggestion)"
          >
            {{ suggestion }}
          </button>
        </div>
      </div>
      <div v-if="recentSuggestions.length > 0" class="shopping-suggestion-section">
        <p class="shopping-suggestion-title">最近追加した項目</p>
        <div class="shopping-suggestion-list">
          <button
            v-for="suggestion in recentSuggestions"
            :key="suggestion"
            type="button"
            class="shopping-suggestion-chip shopping-suggestion-chip-secondary"
            @click="submitSuggestion(suggestion)"
          >
            {{ suggestion }}
          </button>
        </div>
      </div>
    </div>

    <label class="hd-toggle">
      <input v-model="showDone" type="checkbox" />
      完了も表示
    </label>

    <ul v-if="displayedItems.length > 0" class="hd-list">
      <li v-for="note in displayedItems" :key="note.id" :class="['hd-list-item', { 'is-done': note.done }]">
        <span class="hd-list-body">{{ note.body }}</span>
        <button class="hd-btn hd-btn-small" type="button" :disabled="isPending(note.id)" @click="onToggleDone(note)">
          {{ note.done ? '未完了へ' : '完了' }}
        </button>
        <button
          class="hd-btn hd-btn-small hd-btn-danger"
          type="button"
          :disabled="isPending(note.id)"
          @click="requestDelete(note)"
        >
          削除
        </button>
      </li>
    </ul>
    <p v-else class="hd-empty">表示する買い物メモはありません</p>

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
.is-done {
  color: #7b8598;
  border-color: #d6deeb;
  background: #f3f6fb;
  text-decoration: line-through;
}

.shopping-suggestions {
  display: grid;
  gap: 10px;
  margin-top: 12px;
}

.shopping-suggestion-section {
  display: grid;
  gap: 8px;
}

.shopping-suggestion-title {
  margin: 0;
  color: #5b6784;
  font-size: 0.82rem;
  font-weight: 700;
}

.shopping-suggestion-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.shopping-suggestion-chip {
  min-height: 34px;
  padding: 0 12px;
  border: 1px solid #c9d6ed;
  border-radius: 999px;
  background: #eef4ff;
  color: #2c5db9;
  font: inherit;
  font-size: 0.85rem;
  font-weight: 700;
}

.shopping-suggestion-chip-secondary {
  background: #f6f8fc;
  color: #52617d;
}
</style>
