<script setup lang="ts">
import { nextTick, ref } from 'vue'
import type { Note } from '../api/client'

const props = defineProps<{
  items: Note[]
  pendingIds: number[]
  onAdd: (body: string) => Promise<void>
  onTogglePin: (note: Note) => Promise<void>
  onDeleteNote: (note: Note) => Promise<void>
}>()

const body = ref('')
const validationError = ref('')
const actionError = ref('')
const inputRef = ref<HTMLInputElement | null>(null)
const deleteTarget = ref<Note | null>(null)

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

async function onTogglePin(note: Note): Promise<void> {
  actionError.value = ''
  try {
    await props.onTogglePin(note)
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
      <li v-for="note in items" :key="note.id" :class="['hd-list-item', { 'is-pinned': note.pinned }]">
        <span class="hd-list-body">{{ note.body }}</span>
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
.is-pinned {
  border-color: #ead59a;
  background: #fff4d2;
  box-shadow: inset 4px 0 0 #e6bb40;
}
</style>
