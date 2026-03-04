<script setup lang="ts">
import type { DailyGarbage } from '../api/client'

defineProps<{
  today: DailyGarbage
  tomorrow: DailyGarbage
}>()
</script>

<template>
  <section class="hd-panel garbage-panel">
    <h2 class="hd-panel-title">ゴミ</h2>
    <div class="garbage-block">
      <h3 class="garbage-subtitle">今日</h3>
      <p class="garbage-date">{{ today.date }} ({{ today.weekday }})</p>
      <p class="garbage-label">{{ today.label }}</p>
    </div>
    <div class="garbage-block">
      <h3 class="garbage-subtitle">明日</h3>
      <p class="garbage-date">{{ tomorrow.date }} ({{ tomorrow.weekday }})</p>
      <p class="garbage-label">{{ tomorrow.label }}</p>
    </div>
  </section>
</template>

<style scoped>
.garbage-panel {
  padding: 16px;
}

.garbage-block + .garbage-block {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #ddd;
}

.garbage-subtitle {
  margin: 0 0 6px;
  font-size: 1rem;
}

.garbage-date {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.garbage-label {
  margin: 4px 0 0;
  font-weight: 600;
}
</style>
