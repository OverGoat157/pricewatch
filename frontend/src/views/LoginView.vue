<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { apiError } from '../api'

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = apiError(e, 'Не удалось войти')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-screen">
    <form class="card auth-card" @submit.prevent="submit">
      <h1>📉 PriceWatch</h1>
      <p class="sub">Мониторинг цен на товары Wildberries</p>
      <div v-if="error" class="alert alert--error">{{ error }}</div>
      <div class="field">
        <label>Email</label>
        <input v-model="email" type="email" class="input" required autocomplete="email" />
      </div>
      <div class="field">
        <label>Пароль</label>
        <input v-model="password" type="password" class="input" required autocomplete="current-password" />
      </div>
      <button class="btn" style="width: 100%" :disabled="loading">
        {{ loading ? 'Вход…' : 'Войти' }}
      </button>
      <p class="auth-switch">
        Нет аккаунта? <RouterLink to="/register">Зарегистрироваться</RouterLink>
      </p>
    </form>
  </div>
</template>
