import { describe, it, expect, beforeEach } from 'vitest'
import { createRouter, createWebHistory } from 'vue-router'

function createTestRouter() {
  const router = createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/login', name: 'login', component: { template: '<div />' } },
      { path: '/', name: 'dashboard', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/keys', name: 'keys', component: { template: '<div />' }, meta: { requiresAuth: true } },
      { path: '/settings', name: 'settings', component: { template: '<div />' }, meta: { requiresAuth: true } },
    ],
  })

  router.beforeEach((to) => {
    const token = localStorage.getItem('gatecha_token')
    if (to.meta.requiresAuth && !token) {
      return { name: 'login' }
    }
    if (to.name === 'login' && token) {
      return { name: 'dashboard' }
    }
  })

  return router
}

describe('router guard', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('redirects to login when accessing protected route without token', async () => {
    const router = createTestRouter()
    router.push('/')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('login')
  })

  it('redirects to dashboard when accessing login with token', async () => {
    localStorage.setItem('gatecha_token', 'test-token')
    const router = createTestRouter()
    router.push('/login')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('allows access to protected route with token', async () => {
    localStorage.setItem('gatecha_token', 'test-token')
    const router = createTestRouter()
    router.push('/settings')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('settings')
  })

  it('allows access to login without token', async () => {
    const router = createTestRouter()
    router.push('/login')
    await router.isReady()
    expect(router.currentRoute.value.name).toBe('login')
  })
})
