import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'login',
      component: () => import('../views/LoginView.vue'),
    },
    {
      path: '/',
      name: 'dashboard',
      component: () => import('../views/DashboardView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/keys',
      name: 'keys',
      component: () => import('../views/ApiKeysView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/keys/new',
      name: 'keys-new',
      component: () => import('../views/ApiKeyFormView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/keys/:id',
      name: 'keys-detail',
      component: () => import('../views/ApiKeyDetailView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/keys/:id/edit',
      name: 'keys-edit',
      component: () => import('../views/ApiKeyFormView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/settings',
      name: 'settings',
      component: () => import('../views/SettingsView.vue'),
      meta: { requiresAuth: true },
    },
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

export default router
