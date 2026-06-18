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
  if (password.value.length < 6) {
    error.value = 'Пароль должен быть не короче 6 символов'
    return
  }
  loading.value = true
  try {
    await auth.register(email.value, password.value)
    router.push('/')
  } catch (e) {
    error.value = apiError(e, 'Не удалось зарегистрироваться')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-screen">
    <form class="card auth-card" @submit.prevent="submit">
      <h1>Регистрация</h1>
      <p class="sub">Создайте аккаунт PriceWatch</p>
      <div v-if="error" class="alert alert--error">{{ error }}</div>
      <div class="field">
        <label>Email</label>
        <input v-model="email" type="email" class="input" required autocomplete="email" />
      </div>
      <div class="field">
        <label>Пароль</label>
        <input v-model="password" type="password" class="input" required autocomplete="new-password" />
      </div>
      <button class="btn" style="width: 100%" :disabled="loading">
        {{ loading ? 'Создаём…' : 'Зарегистрироваться' }}
      </button>
      <p class="auth-switch">
        Уже есть аккаунт? <RouterLink to="/login">Войти</RouterLink>
      </p>
    </form>
  </div>
</template>
