<script setup lang="ts">
import { computed, nextTick, ref } from 'vue'
import type { Note } from '../api/client'

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
  validationError.value = validate(body.value)
  actionError.value = ''
  if (validationError.value !== '') {
    return
  }

  try {
    await props.onAdd(body.value.trim())
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
  <section class="panel">
    <h2>買い物</h2>

    <form class="composer" @submit.prevent="submit">
      <input
        ref="inputRef"
        v-model="body"
        type="text"
        placeholder="買い物メモを入力"
        maxlength="200"
        autocomplete="off"
      />
      <button type="submit">追加</button>
    </form>
    <p v-if="validationError" class="error">{{ validationError }}</p>
    <p v-if="actionError" class="error">{{ actionError }}</p>

    <label class="toggle">
      <input v-model="showDone" type="checkbox" />
      完了も表示
    </label>

    <ul v-if="displayedItems.length > 0" class="list">
      <li v-for="note in displayedItems" :key="note.id" :class="{ done: note.done }">
        <span class="body">{{ note.body }}</span>
        <button class="small" type="button" :disabled="isPending(note.id)" @click="onToggleDone(note)">
          {{ note.done ? '未完了へ' : '完了' }}
        </button>
        <button class="small danger" type="button" :disabled="isPending(note.id)" @click="requestDelete(note)">
          削除
        </button>
      </li>
    </ul>
    <p v-else class="empty">表示する買い物メモはありません</p>

    <div v-if="deleteTarget" class="confirm-overlay" role="dialog" aria-modal="true" aria-label="削除確認">
      <div class="confirm-modal">
        <p>削除しますか？</p>
        <div class="confirm-actions">
          <button type="button" class="small" @click="cancelDelete">キャンセル</button>
          <button type="button" class="small danger" @click="confirmDelete">削除</button>
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.panel {
  border: 1px solid #dcdcdc;
  border-radius: 10px;
  padding: 14px;
  background: #fff;
}

h2 {
  margin: 0 0 10px;
  font-size: 1.1rem;
}

.composer {
  display: flex;
  gap: 8px;
}

input[type='text'] {
  flex: 1;
  height: 44px;
  padding: 0 12px;
  border: 1px solid #ccc;
  border-radius: 8px;
  font-size: 16px;
}

button {
  min-height: 44px;
  padding: 0 14px;
  border: 1px solid #bbb;
  border-radius: 8px;
  background: #f7f7f7;
  font-size: 14px;
}

button:disabled {
  opacity: 0.5;
}

.toggle {
  display: inline-flex;
  gap: 6px;
  align-items: center;
  margin-top: 10px;
  font-size: 0.9rem;
  color: #444;
}

.list {
  margin: 10px 0 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 8px;
}

li {
  display: flex;
  gap: 8px;
  align-items: center;
  padding: 8px;
  border-radius: 8px;
  background: #f7f7f7;
}

li.done {
  color: #888;
  text-decoration: line-through;
}

.body {
  flex: 1;
}

.small {
  min-height: 38px;
  padding: 0 10px;
  font-size: 13px;
}

.danger {
  color: #b00020;
}

.error {
  margin: 8px 0 0;
  color: #b00020;
  font-size: 0.86rem;
}

.empty {
  margin: 10px 0 0;
  color: #666;
}

.confirm-overlay {
  position: fixed;
  inset: 0;
  z-index: 70;
  background: rgba(0, 0, 0, 0.4);
  display: grid;
  place-items: center;
  padding: 16px;
}

.confirm-modal {
  width: min(360px, 100%);
  border-radius: 12px;
  border: 1px solid #ddd;
  background: #fff;
  padding: 14px;
}

.confirm-modal p {
  margin: 0 0 12px;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}
</style>
