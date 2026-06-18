<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import api, { apiError } from '../api'
import { formatRub, formatDateTime } from '../format'

const props = defineProps({ sub: { type: Object, required: true } })
const emit = defineEmits(['removed', 'updated'])
const router = useRouter()
const checking = ref(false)
const error = ref('')

async function checkNow() {
  checking.value = true
  error.value = ''
  try {
    const { data } = await api.post(`/api/subscriptions/${props.sub.id}/check`)
    emit('updated', data)
  } catch (e) {
    error.value = apiError(e, 'Не удалось проверить цену')
  } finally {
    checking.value = false
  }
}

async function remove() {
  if (!confirm('Удалить товар из отслеживания?')) return
  try {
    await api.delete(`/api/subscriptions/${props.sub.id}`)
    emit('removed', props.sub.id)
  } catch (e) {
    error.value = apiError(e)
  }
}

function hideImg(e) {
  e.target.style.display = 'none'
}
</script>

<template>
  <div class="card product-card">
    <div class="product-card__top" @click="router.push(`/product/${sub.id}`)">
      <img
        v-if="sub.product.image_url"
        :src="sub.product.image_url"
        class="thumb"
        alt=""
        @error="hideImg"
      />
      <div v-else class="thumb thumb--ph">🛍️</div>
      <div class="product-card__info">
        <div class="product-card__title">
          {{ sub.product.title || 'Товар ' + sub.product.external_id }}
        </div>
        <div class="prices">
          <span class="price">{{ formatRub(sub.product.last_price) }}</span>
          <span class="badge" :class="sub.below_target ? 'badge--success' : 'badge--muted'">
            цель {{ formatRub(sub.target_price) }}
          </span>
        </div>
        <div class="meta">
          <span v-if="!sub.product.is_available" class="badge badge--danger">нет в наличии</span>
          <span class="muted">обновлено {{ formatDateTime(sub.product.last_checked_at) }}</span>
        </div>
      </div>
    </div>

    <div v-if="error" class="alert alert--error" style="margin: 0 14px">{{ error }}</div>

    <div class="product-card__actions">
      <button class="btn btn--ghost btn--sm" :disabled="checking" @click="checkNow">
        {{ checking ? '…' : 'Проверить' }}
      </button>
      <RouterLink :to="`/product/${sub.id}`" class="btn btn--ghost btn--sm">История</RouterLink>
      <button class="btn btn--danger btn--sm" style="margin-left: auto" @click="remove">Удалить</button>
    </div>
  </div>
</template>

<style scoped>
.product-card {
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.product-card__top {
  display: flex;
  gap: 14px;
  padding: 16px;
  cursor: pointer;
}
.thumb {
  width: 72px;
  height: 72px;
  object-fit: cover;
  border-radius: 10px;
  background: var(--bg);
  flex-shrink: 0;
}
.thumb--ph {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 30px;
}
.product-card__info {
  min-width: 0;
  flex: 1;
}
.product-card__title {
  font-weight: 600;
  font-size: 15px;
  line-height: 1.35;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.prices {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  margin-bottom: 8px;
}
.price {
  font-size: 20px;
  font-weight: 700;
}
.meta {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  flex-wrap: wrap;
}
.product-card__actions {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-top: 1px solid var(--border);
  margin-top: auto;
}
</style>
