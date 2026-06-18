<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api, { apiError } from '../api'
import { formatRub, formatDateTime } from '../format'
import PriceChart from '../components/PriceChart.vue'

const route = useRoute()
const router = useRouter()

const sub = ref(null)
const history = ref([])
const loading = ref(true)
const error = ref('')
const editing = ref(false)
const newTarget = ref('')
const checking = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    const { data } = await api.get(`/api/subscriptions/${route.params.id}`)
    sub.value = data.subscription
    history.value = data.history
    newTarget.value = data.subscription.target_price
  } catch (e) {
    error.value = apiError(e)
  } finally {
    loading.value = false
  }
}

const stats = computed(() => {
  if (!history.value.length) return null
  const prices = history.value.map((p) => p.price)
  return {
    min: Math.min(...prices),
    max: Math.max(...prices),
    current: prices[prices.length - 1],
  }
})

const reversedHistory = computed(() => [...history.value].reverse())

async function saveTarget() {
  try {
    const { data } = await api.patch(`/api/subscriptions/${route.params.id}`, {
      target_price: Number(newTarget.value),
    })
    sub.value = data
    editing.value = false
  } catch (e) {
    error.value = apiError(e)
  }
}

async function checkNow() {
  checking.value = true
  error.value = ''
  try {
    await api.post(`/api/subscriptions/${route.params.id}/check`)
    await load()
  } catch (e) {
    error.value = apiError(e, 'Не удалось получить актуальную цену')
  } finally {
    checking.value = false
  }
}

async function remove() {
  if (!confirm('Удалить товар из отслеживания?')) return
  await api.delete(`/api/subscriptions/${route.params.id}`)
  router.push('/')
}

onMounted(load)
</script>

<template>
  <div class="container">
    <RouterLink to="/" class="back">← К списку товаров</RouterLink>

    <div v-if="loading" class="muted">Загрузка…</div>
    <div v-else-if="error" class="alert alert--error">{{ error }}</div>

    <template v-else-if="sub">
      <div class="card head">
        <img
          v-if="sub.product.image_url"
          :src="sub.product.image_url"
          class="head__img"
          alt=""
          @error="(e) => (e.target.style.display = 'none')"
        />
        <div class="head__info">
          <h1 class="head__title">{{ sub.product.title || 'Товар ' + sub.product.external_id }}</h1>
          <div class="head__price">
            {{ formatRub(sub.product.last_price) }}
            <span class="badge" :class="sub.below_target ? 'badge--success' : 'badge--muted'">
              {{ sub.below_target ? 'цена достигнута' : 'выше цели' }}
            </span>
            <span v-if="!sub.product.is_available" class="badge badge--danger">нет в наличии</span>
          </div>
          <div class="muted" style="font-size: 13px">
            Артикул {{ sub.product.external_id }} ·
            <a :href="sub.product.url" target="_blank" rel="noopener">открыть на Wildberries ↗</a>
          </div>

          <div class="target-row">
            <template v-if="!editing">
              <span>Целевая цена: <b>{{ formatRub(sub.target_price) }}</b></span>
              <button class="btn btn--ghost btn--sm" @click="editing = true">Изменить</button>
            </template>
            <template v-else>
              <input v-model="newTarget" type="number" min="1" class="input" style="width: 140px" />
              <button class="btn btn--sm" @click="saveTarget">Сохранить</button>
              <button class="btn btn--ghost btn--sm" @click="editing = false">Отмена</button>
            </template>
          </div>

          <div class="head__actions">
            <button class="btn btn--sm" :disabled="checking" @click="checkNow">
              {{ checking ? 'Проверяю…' : 'Проверить цену сейчас' }}
            </button>
            <button class="btn btn--danger btn--sm" @click="remove">Удалить</button>
          </div>
        </div>
      </div>

      <div v-if="stats" class="stats">
        <div class="card stat">
          <div class="stat__label">Текущая</div>
          <div class="stat__value">{{ formatRub(stats.current) }}</div>
        </div>
        <div class="card stat">
          <div class="stat__label">Минимум</div>
          <div class="stat__value" style="color: var(--success)">{{ formatRub(stats.min) }}</div>
        </div>
        <div class="card stat">
          <div class="stat__label">Максимум</div>
          <div class="stat__value">{{ formatRub(stats.max) }}</div>
        </div>
        <div class="card stat">
          <div class="stat__label">Точек истории</div>
          <div class="stat__value">{{ history.length }}</div>
        </div>
      </div>

      <div class="card section">
        <h2 class="section__title">Динамика цены</h2>
        <PriceChart v-if="history.length" :history="history" :target="sub.target_price" />
        <p v-else class="muted">История ещё пустая — нажмите «Проверить цену сейчас».</p>
      </div>

      <div class="card section" v-if="history.length">
        <h2 class="section__title">История проверок</h2>
        <table class="history-table">
          <thead>
            <tr>
              <th>Дата и время</th>
              <th>Цена</th>
              <th>Наличие</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(p, i) in reversedHistory" :key="i">
              <td>{{ formatDateTime(p.checked_at) }}</td>
              <td>{{ formatRub(p.price) }}</td>
              <td>
                <span class="badge" :class="p.is_available ? 'badge--success' : 'badge--danger'">
                  {{ p.is_available ? 'в наличии' : 'нет' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </template>
  </div>
</template>

<style scoped>
.back {
  display: inline-block;
  margin-bottom: 16px;
  font-size: 14px;
  color: var(--muted);
}
.head {
  display: flex;
  gap: 20px;
  padding: 20px;
}
.head__img {
  width: 120px;
  height: 120px;
  object-fit: cover;
  border-radius: 12px;
  background: var(--bg);
}
.head__title {
  font-size: 20px;
  margin: 0 0 10px;
}
.head__price {
  font-size: 26px;
  font-weight: 700;
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 6px;
}
.target-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin: 14px 0 4px;
  font-size: 14px;
}
.head__actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}
.stats {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 14px;
  margin: 18px 0;
}
.stat {
  padding: 16px;
}
.stat__label {
  font-size: 13px;
  color: var(--muted);
  margin-bottom: 6px;
}
.stat__value {
  font-size: 20px;
  font-weight: 700;
}
.section {
  padding: 20px;
  margin-bottom: 18px;
}
.section__title {
  font-size: 16px;
  margin: 0 0 16px;
}
.history-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 14px;
}
.history-table th {
  text-align: left;
  color: var(--muted);
  font-weight: 600;
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
}
.history-table td {
  padding: 8px 10px;
  border-bottom: 1px solid var(--border);
}
@media (max-width: 720px) {
  .head {
    flex-direction: column;
  }
  .stats {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
