<script setup>
import { ref, onMounted } from 'vue'
import api, { apiError } from '../api'
import AddProductForm from '../components/AddProductForm.vue'
import ProductCard from '../components/ProductCard.vue'

const subs = ref([])
const loading = ref(true)
const error = ref('')

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/api/subscriptions')
    subs.value = data
  } catch (e) {
    error.value = apiError(e)
  } finally {
    loading.value = false
  }
}

function onAdded(sub) {
  const idx = subs.value.findIndex((s) => s.id === sub.id)
  if (idx >= 0) subs.value[idx] = sub
  else subs.value = [sub, ...subs.value]
}
function onRemoved(id) {
  subs.value = subs.value.filter((s) => s.id !== id)
}
function onUpdated(sub) {
  subs.value = subs.value.map((s) => (s.id === sub.id ? sub : s))
}

onMounted(load)
</script>

<template>
  <div class="container">
    <h1 class="page-title">Отслеживаемые товары</h1>
    <AddProductForm @added="onAdded" />

    <div v-if="error" class="alert alert--error" style="margin-top: 16px">{{ error }}</div>

    <div v-if="loading" class="muted" style="margin-top: 20px">Загрузка…</div>
    <div v-else-if="subs.length === 0" class="card empty">
      <p class="muted">
        Пока нет отслеживаемых товаров.<br />
        Вставьте ссылку на товар Wildberries и целевую цену в форму выше.
      </p>
    </div>
    <div v-else class="product-grid">
      <ProductCard
        v-for="s in subs"
        :key="s.id"
        :sub="s"
        @removed="onRemoved"
        @updated="onUpdated"
      />
    </div>
  </div>
</template>

<style scoped>
.product-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 16px;
  margin-top: 20px;
}
.empty {
  margin-top: 20px;
  padding: 48px 24px;
  text-align: center;
}
</style>
