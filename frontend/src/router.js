import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'
import LoginView from './views/LoginView.vue'
import RegisterView from './views/RegisterView.vue'
import DashboardView from './views/DashboardView.vue'
import ProductView from './views/ProductView.vue'
import SettingsView from './views/SettingsView.vue'

const routes = [
  { path: '/login', component: LoginView, meta: { public: true } },
  { path: '/register', component: RegisterView, meta: { public: true } },
  { path: '/', component: DashboardView },
  { path: '/product/:id', component: ProductView },
  { path: '/settings', component: SettingsView },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) return '/login'
  if (to.meta.public && auth.isAuthenticated) return '/'
  return true
})

export default router
