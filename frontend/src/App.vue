<script setup>
import { onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import AppHeader from './components/AppHeader.vue'

const route = useRoute()
const auth = useAuthStore()

onMounted(() => {
  if (auth.isAuthenticated) auth.fetchMe()
})
</script>

<template>
  <AppHeader v-if="auth.isAuthenticated && !route.meta.public" />
  <RouterView />
</template>
