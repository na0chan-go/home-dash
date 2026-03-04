const CACHE_VERSION = 'homedash-shell-v2'
const APP_SHELL_ASSETS = [
  '/',
  '/index.html',
  '/manifest.webmanifest',
  '/apple-touch-icon.png',
  '/icons/icon-192.png',
  '/icons/icon-512.png'
]

async function cacheBuiltAssets(cache) {
  const response = await fetch('/index.html', { cache: 'no-store' })
  if (!response.ok) {
    return
  }
  const html = await response.text()
  const assetPaths = new Set()
  const assetPattern = /(?:src|href)="(\/assets\/[^"]+)"/g
  let match = assetPattern.exec(html)
  while (match !== null) {
    assetPaths.add(match[1])
    match = assetPattern.exec(html)
  }
  if (assetPaths.size > 0) {
    await cache.addAll(Array.from(assetPaths))
  }
}

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_VERSION).then(async (cache) => {
      await cache.addAll(APP_SHELL_ASSETS)
      await cacheBuiltAssets(cache)
    })
  )
  self.skipWaiting()
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) =>
      Promise.all(
        keys
          .filter((key) => key !== CACHE_VERSION)
          .map((key) => caches.delete(key))
      )
    )
  )
  self.clients.claim()
})

self.addEventListener('fetch', (event) => {
  const request = event.request
  if (request.method !== 'GET') {
    return
  }

  const url = new URL(request.url)
  if (url.origin !== self.location.origin) {
    return
  }

  if (url.pathname === '/api' || url.pathname.startsWith('/api/')) {
    return
  }

  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request)
        .then((response) => {
          const cloned = response.clone()
          caches.open(CACHE_VERSION).then((cache) => cache.put('/index.html', cloned))
          return response
        })
        .catch(async () => {
          const cached = await caches.match('/index.html')
          if (cached) {
            return cached
          }
          return new Response('offline', { status: 503, statusText: 'Service Unavailable' })
        })
    )
    return
  }

  const isStaticResource =
    request.destination === 'script' ||
    request.destination === 'style' ||
    request.destination === 'image' ||
    request.destination === 'font' ||
    url.pathname === '/manifest.webmanifest'

  if (!isStaticResource) {
    return
  }

  event.respondWith(
    caches.match(request).then((cached) => {
      if (cached) {
        return cached
      }
      return fetch(request).then((response) => {
        if (response.ok) {
          const cloned = response.clone()
          caches.open(CACHE_VERSION).then((cache) => cache.put(request, cloned))
        }
        return response
      })
    })
  )
})
