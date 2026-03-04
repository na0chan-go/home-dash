<script setup lang="ts">
import type { DailyGarbage } from '../api/client'

defineProps<{
  today: DailyGarbage
  tomorrow: DailyGarbage
}>()
</script>

<template>
  <section class="panel">
    <h2>ゴミ</h2>
    <div class="block">
      <h3>今日</h3>
      <p class="date">{{ today.date }} ({{ today.weekday }})</p>
      <p class="label">{{ today.label }}</p>
    </div>
    <div class="block">
      <h3>明日</h3>
      <p class="date">{{ tomorrow.date }} ({{ tomorrow.weekday }})</p>
      <p class="label">{{ tomorrow.label }}</p>
    </div>
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

.block + .block {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed #ddd;
}

h3 {
  margin: 0 0 6px;
  font-size: 1rem;
}

.date {
  margin: 0;
  color: #666;
  font-size: 0.9rem;
}

.label {
  margin: 4px 0 0;
  font-weight: 600;
}
</style>
