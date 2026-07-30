<template>
  <div class="h-screen w-screen flex bg-gray-100 overflow-hidden">
    <div class="w-full max-w-7xl mx-auto my-0 md:my-6 md:px-4 flex h-full md:h-[calc(100vh-3rem)] bg-white md:rounded-xl md:border md:border-gray-300 md:shadow-md overflow-hidden">
      <!-- SIDEBAR -->
      <div 
        :class="[
          'w-full md:w-[350px] flex-shrink-0 h-full',
          chatStore.selectedFriendId ? 'hidden md:block' : 'block'
        ]"
      >
        <Sidebar />
      </div>

      <!-- CHAT WINDOW -->
      <div 
        :class="[
          'flex-1 h-full min-w-0 border-l border-gray-200',
          !chatStore.selectedFriendId ? 'hidden md:flex' : 'flex'
        ]"
      >
        <ChatWindow @back="handleBack" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { onMounted, onUnmounted } from 'vue'
import Sidebar from '../components/Sidebar.vue'
import ChatWindow from '../components/ChatWindow.vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'

const authStore = useAuthStore()
const chatStore = useChatStore()

function handleBack() {
  chatStore.selectedFriendId = null
}

onMounted(() => {
  if (authStore.token) {
    chatStore.fetchUnreadCounts()
    chatStore.connectWebSocket(authStore.token)
  }
})

onUnmounted(() => {
  chatStore.disconnectWebSocket()
})
</script>
