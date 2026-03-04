<script setup lang="ts">
import type { Note } from '../api/client'

defineProps<{
  items: Note[]
}>()
</script>

<template>
  <section class="panel">
    <h2>買い物</h2>
    <ul v-if="items.length > 0" class="list">
      <li v-for="note in items" :key="note.id" :class="{ done: note.done }">
        <span class="check">{{ note.done ? '✓' : '・' }}</span>
        <span class="body">{{ note.body }}</span>
      </li>
    </ul>
    <p v-else class="empty">買い物メモはありません</p>
  </section>
</template>

<style scoped>
.panel {
  border: 1px solid #dcdcdc;
  border-radius: 10px;
  padding: 16px;
  background: #fff;
}

h2 {
  margin: 0 0 12px;
  font-size: 1.1rem;
}

.list {
  margin: 0;
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

.check {
  width: 18px;
  text-align: center;
}

.body {
  flex: 1;
}

.empty {
  margin: 0;
  color: #666;
}
</style>
