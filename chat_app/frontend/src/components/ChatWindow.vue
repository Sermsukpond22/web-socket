<template>
  <div class="w-full h-full flex flex-col bg-white">
    <!-- EMPTY STATE WHEN NO FRIEND SELECTED -->
    <div v-if="!selectedFriend" class="h-full flex flex-col items-center justify-center p-8 text-center bg-gray-50">
      <div class="w-24 h-24 rounded-full border-2 border-gray-900 flex items-center justify-center text-gray-900 mb-4 shadow-sm">
        <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
      </div>
      <h2 class="text-xl font-light text-gray-900 mb-1">ข้อความของคุณ</h2>
      <p class="text-sm text-gray-500 max-w-xs mb-6">ส่งข้อความและแชทกับเพื่อนของคุณแบบเรียลไทม์</p>
    </div>

    <!-- ACTIVE CHAT WINDOW -->
    <div v-else class="h-full flex flex-col">
      <!-- Chat Header -->
      <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between bg-white z-10">
        <div class="flex items-center space-x-3">
          <!-- Mobile Back Button -->
          <button 
            @click="$emit('back')"
            class="md:hidden text-gray-600 hover:text-gray-900 p-1.5 rounded-full hover:bg-gray-100 transition mr-1"
            title="กลับไปหน้าแชท"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>

          <!-- Friend Avatar -->
          <div class="relative">
            <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-sm shadow-sm">
              {{ getInitial(selectedFriend.username || selectedFriend.email) }}
            </div>
            <span class="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white bg-green-500"></span>
          </div>

          <!-- Friend Info -->
          <div>
            <h3 class="font-semibold text-sm text-gray-900 leading-tight">
              {{ selectedFriend.username || selectedFriend.email }}
            </h3>
            <p v-if="chatStore.isUserTyping(chatStore.selectedFriendId)" class="text-xs text-green-600 font-medium">กำลังพิมพ์...</p>
            <p v-else class="text-xs text-green-600 font-medium">กำลังใช้งาน</p>
          </div>
        </div>
      </div>

      <!-- Messages Thread -->
      <div 
        ref="messagesContainer" 
        class="flex-1 overflow-y-auto p-4 space-y-3 bg-white"
      >
        <div 
          v-if="chatStore.activeMessages.length === 0" 
          class="h-full flex items-center justify-center text-center text-gray-400 text-sm"
        >
          ยังไม่มีข้อความ ทักทาย {{ selectedFriend.username || 'เพื่อน' }} เลยสิ!
        </div>

        <div
          v-for="msg in chatStore.activeMessages"
          :key="msg.id || msg.created_at"
          :class="[
            'flex flex-col',
            isSelf(msg) ? 'items-end' : 'items-start'
          ]"
        >
          <!-- Message Bubble -->
          <div
            :class="[
              'max-w-[85%] md:max-w-[75%] px-4 py-2 text-sm rounded-2xl break-words shadow-sm',
              isSelf(msg)
                ? 'bg-sky-500 text-white rounded-br-sm'
                : 'bg-gray-100 text-gray-900 rounded-bl-sm'
            ]"
          >
            {{ msg.content }}
          </div>

          <!-- Timestamp -->
          <span class="text-[10px] text-gray-400 mt-1 px-1">
            {{ formatTime(msg.created_at) }}
          </span>
        </div>

        <!-- Typing Indicator -->
        <div v-if="chatStore.isUserTyping(chatStore.selectedFriendId)" class="flex flex-col items-start mt-2 mb-2">
          <div class="px-4 py-3 bg-gray-100 rounded-2xl rounded-bl-sm w-[72px] shadow-sm">
            <div class="flex space-x-1.5 items-center justify-center h-full">
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 0ms"></div>
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 150ms"></div>
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 300ms"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <div class="p-3 border-t border-gray-200 bg-white w-full">
        <form @submit.prevent="handleSend" class="flex items-end space-x-2 w-full">
          <textarea
            v-model="inputMessage"
            @input="handleTyping"
            placeholder="พิมพ์ข้อความ..."
            rows="1"
            class="flex-1 px-4 py-2.5 bg-gray-100 rounded-2xl text-sm text-gray-900 placeholder-gray-500 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-sky-500 border border-transparent focus:border-sky-500 transition resize-none overflow-hidden min-h-[44px] max-h-[120px]"
            @keydown.enter.exact.prevent="handleSend"
          ></textarea>

          <button
            type="submit"
            :disabled="!inputMessage.trim()"
            class="px-4 py-2.5 bg-sky-500 hover:bg-sky-600 active:bg-sky-700 disabled:opacity-40 text-white text-sm font-semibold rounded-full transition duration-150 flex items-center justify-center h-[44px]"
          >
            ส่ง
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFriendsStore } from '../stores/friends'
import { useChatStore } from '../stores/chat'

const emit = defineEmits(['back'])

const authStore = useAuthStore()
const friendsStore = useFriendsStore()
const chatStore = useChatStore()

const inputMessage = ref('')
const messagesContainer = ref(null)
let typingThrottleTimer = null

const selectedFriend = computed(() => {
  if (!chatStore.selectedFriendId) return null
  return friendsStore.friends.find(f => f.id === chatStore.selectedFriendId) || {
    id: chatStore.selectedFriendId,
    username: `User #${chatStore.selectedFriendId}`
  }
})

function isSelf(msg) {
  return msg.sender_id === authStore.user?.id
}

function getInitial(str) {
  if (!str) return 'U'
  return str.charAt(0).toUpperCase()
}

function formatTime(timestamp) {
  if (!timestamp) return ''
  try {
    const date = new Date(timestamp)
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch (e) {
    return ''
  }
}

function scrollToBottom() {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

function handleTyping() {
  if (!chatStore.selectedFriendId) return
  if (!typingThrottleTimer) {
    chatStore.sendTyping(chatStore.selectedFriendId, true)
    typingThrottleTimer = setTimeout(() => {
      typingThrottleTimer = null
    }, 2000)
  }
}

function handleSend() {
  if (!inputMessage.value.trim() || !chatStore.selectedFriendId) return
  chatStore.sendMessage(chatStore.selectedFriendId, inputMessage.value)
  inputMessage.value = ''
  scrollToBottom()
}

watch(
  () => chatStore.activeMessages,
  () => {
    scrollToBottom()
  },
  { deep: true }
)

watch(
  () => chatStore.selectedFriendId,
  () => {
    scrollToBottom()
  }
)
</script>

