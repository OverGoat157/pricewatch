<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import api from '../api'
import { formatRub, formatDateTime } from '../format'

const open = ref(false)
const items = ref([])
const unread = ref(0)
let timer = null

async function load() {
  try {
    const { data } = await api.get('/api/notifications')
    items.value = data.items
    unread.value = data.unread
  } catch {
    /* тихо игнорируем — фоновый опрос */
  }
}

async function toggle() {
  open.value = !open.value
  if (open.value && unread.value > 0) {
    try {
      await api.post('/api/notifications/read')
    } catch {
      /* не критично */
    }
    unread.value = 0
    items.value = items.value.map((n) => ({ ...n, is_read: true }))
  }
}

onMounted(() => {
  load()
  timer = setInterval(load, 30000)
})
onUnmounted(() => clearInterval(timer))
</script>

<template>
  <div class="bell">
    <button class="bell__btn" @click="toggle" aria-label="Уведомления">
      🔔
      <span v-if="unread > 0" class="bell__badge">{{ unread > 9 ? '9+' : unread }}</span>
    </button>

    <template v-if="open">
      <div class="bell__overlay" @click="open = false"></div>
      <div class="bell__dropdown card">
        <div class="bell__head">Уведомления</div>
        <div v-if="items.length === 0" class="bell__empty muted">Пока нет уведомлений</div>
        <ul v-else class="bell__list">
          <li v-for="n in items" :key="n.id" :class="{ unread: !n.is_read }">
            <div class="bell__title">{{ n.product_title }}</div>
            <div class="bell__price">
              {{ formatRub(n.triggered_price) }}
              <span class="muted">(цель {{ formatRub(n.target_price) }})</span>
            </div>
            <div class="bell__date muted">{{ formatDateTime(n.sent_at) }}</div>
          </li>
        </ul>
      </div>
    </template>
  </div>
</template>

<style scoped>
.bell {
  position: relative;
  display: flex;
}
.bell__btn {
  background: transparent;
  border: none;
  font-size: 20px;
  cursor: pointer;
  position: relative;
  line-height: 1;
  padding: 4px;
}
.bell__badge {
  position: absolute;
  top: -2px;
  right: -4px;
  background: var(--danger);
  color: #fff;
  font-size: 10px;
  font-weight: 700;
  border-radius: 999px;
  padding: 1px 5px;
}
.bell__overlay {
  position: fixed;
  inset: 0;
  z-index: 30;
}
.bell__dropdown {
  position: absolute;
  top: 38px;
  right: 0;
  width: 320px;
  max-height: 420px;
  overflow-y: auto;
  z-index: 31;
  padding: 0;
}
.bell__head {
  padding: 14px 16px;
  font-weight: 700;
  border-bottom: 1px solid var(--border);
}
.bell__empty {
  padding: 24px 16px;
  text-align: center;
  font-size: 14px;
}
.bell__list {
  list-style: none;
  margin: 0;
  padding: 0;
}
.bell__list li {
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}
.bell__list li.unread {
  background: #eef2ff;
}
.bell__title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 2px;
}
.bell__price {
  font-size: 14px;
  color: var(--success);
  font-weight: 600;
}
.bell__date {
  font-size: 12px;
  margin-top: 2px;
}
</style>
