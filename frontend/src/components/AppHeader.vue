<script setup>
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import NotificationsBell from './NotificationsBell.vue'

const auth = useAuthStore()
const router = useRouter()

function logout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <header class="app-header">
    <div class="app-header__inner">
      <RouterLink to="/" class="logo">📉 Price<span>Watch</span></RouterLink>
      <nav class="nav">
        <RouterLink to="/">Товары</RouterLink>
        <RouterLink to="/settings">Настройки</RouterLink>
      </nav>
      <div class="app-header__right">
        <NotificationsBell />
        <span class="user-email">{{ auth.user?.email }}</span>
        <button class="btn btn--ghost btn--sm" @click="logout">Выйти</button>
      </div>
    </div>
  </header>
</template>
