// Gux Service Worker - Cache-first strategy for assets
const CACHE_NAME = 'gux-v1';

// Assets to precache on install
const PRECACHE_URLS = [
  '/',
  '/index.html',
  '/wasm_exec.js',
  '/main.wasm',
  '/manifest.json',
  '/icons/icon-192.png',
  '/icons/icon-512.png'
];

// CDN resources (network-first)
const CDN_HOSTS = [
  'cdnjs.cloudflare.com'
];

// Install event - precache assets
self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('[SW] Precaching app assets');
        return cache.addAll(PRECACHE_URLS);
      })
      .then(() => {
        console.log('[SW] Precache complete');
        return self.skipWaiting();
      })
  );
});

// Activate event - clean old caches
self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames
            .filter((name) => name !== CACHE_NAME)
            .map((name) => {
              console.log('[SW] Deleting old cache:', name);
              return caches.delete(name);
            })
        );
      })
      .then(() => {
        console.log('[SW] Activated and ready');
        return self.clients.claim();
      })
  );
});

// Fetch event - cache-first for same-origin, network-first for CDN
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);

  // Skip non-GET requests
  if (event.request.method !== 'GET') {
    return;
  }

  // CDN resources - network first, cache fallback
  if (CDN_HOSTS.some(host => url.hostname.includes(host))) {
    event.respondWith(networkFirstStrategy(event.request));
    return;
  }

  // Same-origin resources - cache first, network fallback
  if (url.origin === location.origin) {
    event.respondWith(cacheFirstStrategy(event.request));
    return;
  }
});

// Cache-first strategy
async function cacheFirstStrategy(request) {
  const cachedResponse = await caches.match(request);
  if (cachedResponse) {
    return cachedResponse;
  }

  try {
    const networkResponse = await fetch(request);

    // Cache successful responses
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    console.error('[SW] Fetch failed:', error);
    // Return offline fallback if available
    return caches.match('/index.html');
  }
}

// Network-first strategy (for CDN)
async function networkFirstStrategy(request) {
  try {
    const networkResponse = await fetch(request);

    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }

    return networkResponse;
  } catch (error) {
    console.log('[SW] Network failed, trying cache:', request.url);
    const cachedResponse = await caches.match(request);
    if (cachedResponse) {
      return cachedResponse;
    }
    throw error;
  }
}
