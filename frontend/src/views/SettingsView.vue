<script setup>
import { ref, onMounted } from 'vue'
import api, { apiError } from '../api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const link = ref('')
const token = ref('')
const botName = ref('')
const linked = ref(false)
const loading = ref(false)
const error = ref('')

async function loadLink() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get('/api/telegram/link')
    token.value = data.token
    link.value = data.link || ''
    botName.value = data.bot_name || ''
    linked.value = data.linked
  } catch (e) {
    error.value = apiError(e)
  } finally {
    loading.value = false
  }
}

async function refreshStatus() {
  await auth.fetchMe()
  await loadLink()
}

onMounted(loadLink)
</script>

<template>
  <div class="container">
    <h1 class="page-title">Настройки</h1>

    <div class="card section">
      <h2 class="section__title">Уведомления в Telegram</h2>

      <div v-if="error" class="alert alert--error">{{ error }}</div>

      <div v-if="linked" class="alert alert--info">
        ✅ Telegram подключён — вы получаете уведомления о снижении цен.
      </div>

      <template v-else>
        <p class="muted">
          Подключите Telegram, чтобы получать сообщение, как только цена товара опустится до
          целевой.
        </p>

        <ol class="steps">
          <li>
            Нажмите кнопку ниже — откроется наш бот.
            <div style="margin-top: 8px" v-if="link">
              <a :href="link" target="_blank" rel="noopener" class="btn btn--sm">Подключить Telegram</a>
            </div>
            <div v-else class="alert alert--info" style="margin-top: 8px">
              Имя бота не настроено на сервере. Откройте бота вручную и отправьте ему команду:
              <code>/start {{ token }}</code>
            </div>
          </li>
          <li>В Telegram нажмите кнопку <b>«Запустить» / Start</b>.</li>
          <li>Вернитесь сюда и нажмите «Обновить статус».</li>
        </ol>

        <button class="btn btn--ghost btn--sm" :disabled="loading" @click="refreshStatus">
          {{ loading ? 'Проверяю…' : 'Обновить статус' }}
        </button>
      </template>
    </div>

    <div class="card section">
      <h2 class="section__title">Аккаунт</h2>
      <p class="muted" style="margin: 0">Вы вошли как <b>{{ auth.user?.email }}</b></p>
    </div>
  </div>
</template>

<style scoped>
.section {
  padding: 20px;
  margin-bottom: 18px;
}
.section__title {
  font-size: 16px;
  margin: 0 0 14px;
}
.steps {
  padding-left: 20px;
  line-height: 1.7;
  font-size: 14px;
}
.steps li {
  margin-bottom: 10px;
}
code {
  background: var(--bg);
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 13px;
}
</style>
