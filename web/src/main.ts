import { createApp } from 'vue'
import Dashboard from './pages/Dashboard.vue'
import './style.css'

createApp(Dashboard).mount('#app')

if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch((err) => {
      console.error('service worker registration failed', err)
    })
  })
}
