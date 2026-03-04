<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import NoticeBoard from '../components/NoticeBoard.vue'
import ShoppingList from '../components/ShoppingList.vue'
import GarbagePanel from '../components/GarbagePanel.vue'
import { fetchDashboard, type DashboardResponse } from '../api/client'

const dashboard = ref<DashboardResponse | null>(null)
const loading = ref(true)
const error = ref('')

let pollTimer: number | undefined

async function loadDashboard(): Promise<void> {
  try {
    const data = await fetchDashboard()
    dashboard.value = data
    error.value = ''
  } catch (e) {
    console.error(e)
    error.value = 'ダッシュボードの取得に失敗しました'
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await loadDashboard()
  pollTimer = window.setInterval(() => {
    void loadDashboard()
  }, 30_000)
})

onUnmounted(() => {
  if (pollTimer !== undefined) {
    window.clearInterval(pollTimer)
  }
})
</script>

<template>
  <main class="page">
    <header class="header">
      <h1>HomeDash</h1>
      <p v-if="dashboard">最終更新: {{ dashboard.generatedAt }}</p>
    </header>

    <p v-if="loading">読み込み中...</p>
    <p v-else-if="error">{{ error }}</p>

    <div v-else-if="dashboard" class="grid">
      <NoticeBoard :items="dashboard.notes.notice" />
      <ShoppingList :items="dashboard.notes.shopping" />
      <GarbagePanel :today="dashboard.garbage.today" :tomorrow="dashboard.garbage.tomorrow" />
    </div>
  </main>
</template>

<style scoped>
.page {
  max-width: 1200px;
  margin: 0 auto;
  padding: 24px;
}

.header {
  margin-bottom: 16px;
}

h1 {
  margin: 0;
  font-size: 1.8rem;
}

p {
  margin: 6px 0 0;
}

.grid {
  display: grid;
  gap: 16px;
  grid-template-columns: repeat(3, minmax(0, 1fr));
}

@media (max-width: 900px) {
  .grid {
    grid-template-columns: 1fr;
  }
}
</style>
