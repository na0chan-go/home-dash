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

.garbage-block {
  padding: 12px;
  border-radius: 12px;
  border: 1px solid #dce5f4;
  background: linear-gradient(180deg, #ffffff 0%, #f5f8fc 100%);
}

.garbage-block + .garbage-block {
  margin-top: 12px;
}

.garbage-subtitle {
  margin: 0 0 6px;
  color: #2e3a57;
  font-size: 1.04rem;
}

.garbage-date {
  margin: 0;
  color: #65748f;
  font-size: 0.9rem;
}

.garbage-label {
  margin: 8px 0 0;
  color: #1e2a45;
  font-size: 1.32rem;
  font-weight: 800;
  letter-spacing: 0.02em;
}
</style>
