<template>
  <div class="min-h-screen bg-gray-50 dark:bg-gray-800 text-gray-900 dark:text-gray-100 font-sans">
    <router-view />
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'

const authStore = useAuthStore()

onMounted(() => {
  authStore.initAuth()
})
</script>
