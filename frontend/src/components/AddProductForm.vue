<script setup>
import { ref } from 'vue'
import api, { apiError } from '../api'

const emit = defineEmits(['added'])
const url = ref('')
const target = ref('')
const loading = ref(false)
const error = ref('')

async function submit() {
  error.value = ''
  if (!url.value.trim() || !target.value) {
    error.value = 'Заполните ссылку и целевую цену'
    return
  }
  loading.value = true
  try {
    const { data } = await api.post('/api/subscriptions', {
      url: url.value.trim(),
      target_price: Number(target.value),
    })
    emit('added', data)
    url.value = ''
    target.value = ''
  } catch (e) {
    error.value = apiError(e, 'Не удалось добавить товар')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <form class="card add-form" @submit.prevent="submit">
    <div class="field add-form__url">
      <label>Ссылка или артикул Wildberries</label>
      <input
        v-model="url"
        class="input"
        placeholder="https://www.wildberries.ru/catalog/179978204/detail.aspx"
      />
    </div>
    <div class="field add-form__price">
      <label>Целевая цена, ₽</label>
      <input v-model="target" type="number" min="1" step="1" class="input" placeholder="2990" />
    </div>
    <button class="btn" :disabled="loading">{{ loading ? 'Добавляю…' : 'Отслеживать' }}</button>
  </form>
  <div v-if="error" class="alert alert--error" style="margin-top: 10px">{{ error }}</div>
</template>

<style scoped>
.add-form {
  display: flex;
  gap: 14px;
  align-items: flex-end;
  padding: 18px;
}
.add-form .field {
  margin: 0;
}
.add-form__url {
  flex: 1;
}
.add-form__price {
  width: 160px;
}
@media (max-width: 640px) {
  .add-form {
    flex-direction: column;
    align-items: stretch;
  }
  .add-form__price {
    width: 100%;
  }
}
</style>
