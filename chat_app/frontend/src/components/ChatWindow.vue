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
            <p v-if="chatStore.isUserTyping(chatStore.selectedFriendId)" class="text-xs text-sky-500 font-medium animate-pulse">กำลังพิมพ์...</p>
            <p v-else-if="friendStatus.is_online" class="text-xs text-green-600 font-medium">ออนไลน์</p>
            <p v-else-if="friendStatus.last_seen" class="text-xs text-gray-400 font-medium">ใช้งานล่าสุด: {{ formatLastSeen(friendStatus.last_seen) }}</p>
            <p v-else class="text-xs text-gray-400 font-medium">ออฟไลน์</p>
          </div>
        </div>
      </div>

      <!-- Messages Thread -->
      <div 
        ref="messagesContainer" 
        class="flex-1 overflow-y-auto p-4 space-y-3 bg-white"
        @scroll="handleScroll"
      >
        <div v-if="loadingMore" class="flex justify-center my-2 text-sky-500">
          <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-sky-500"></div>
        </div>
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
          <!-- Message Bubble Container -->
          <div class="relative group flex items-start max-w-[85%] md:max-w-[75%]" :class="isSelf(msg) ? 'flex-row-reverse space-x-reverse space-x-2' : 'flex-row space-x-2'">
            <!-- Message Content -->
            <div
              :class="[
                'px-4 py-2 text-sm rounded-2xl break-words shadow-sm',
                msg.is_deleted ? 'bg-gray-100 text-gray-400 italic border border-gray-200' :
                isSelf(msg) ? 'bg-sky-500 text-white rounded-br-sm' : 'bg-gray-100 text-gray-900 rounded-bl-sm'
              ]"
            >
              <template v-if="msg.is_deleted">
                🚫 ข้อความนี้ถูกลบแล้ว
              </template>
              <template v-else>
                <!-- Image -->
                <div v-if="msg.type === 'image' && msg.file_url">
                  <img :src="'http://localhost:3000' + msg.file_url" alt="image" class="rounded-lg max-w-full h-auto cursor-pointer object-cover mb-1" />
                </div>
                <!-- File -->
                <div v-else-if="msg.type === 'file' && msg.file_url" class="flex items-center space-x-2 bg-black/10 p-2 rounded-lg mb-1">
                  <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"></path></svg>
                  <a :href="'http://localhost:3000' + msg.file_url" target="_blank" class="underline text-sm font-semibold truncate">{{ msg.content || 'ไฟล์แนบ' }}</a>
                </div>
                <!-- Text / Fallback -->
                <div v-if="msg.content && msg.type !== 'file'">
                  {{ msg.content }}
                </div>
              </template>
            </div>
            
            <!-- Actions (Edit/Delete) -->
            <div v-if="isSelf(msg) && !msg.is_deleted && msg.id && !msg.pending" class="opacity-0 group-hover:opacity-100 transition flex flex-col space-y-1 mt-1">
              <button @click="startEdit(msg)" class="text-gray-400 hover:text-sky-500 p-1 bg-gray-50 rounded-full" title="แก้ไข">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>
              </button>
              <button @click="confirmDelete(msg)" class="text-gray-400 hover:text-red-500 p-1 bg-gray-50 rounded-full" title="ลบ">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path></svg>
              </button>
            </div>
          </div>

          <!-- Timestamp & Read Status -->
          <div class="flex items-center mt-1 px-1 space-x-1" :class="isSelf(msg) ? 'justify-end' : 'justify-start'">
            <span class="text-[10px] text-gray-400">
              {{ formatTime(msg.created_at) }}
            </span>
            <span v-if="msg.is_edited && !msg.is_deleted" class="text-[10px] text-gray-400 italic">(แก้ไขแล้ว)</span>
            <span v-if="isSelf(msg) && !msg.is_deleted" class="text-[10px]" :class="msg.is_read ? 'text-sky-500' : 'text-gray-400'">
              <span v-if="msg.is_read">✔✔</span>
              <span v-else>✔</span>
            </span>
          </div>
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
        <div v-if="editingMsg" class="flex items-center justify-between bg-sky-50 p-2 rounded-t-lg border border-sky-100 border-b-0">
          <div class="text-xs text-sky-700 flex flex-col">
            <span class="font-bold">แก้ไขข้อความ</span>
            <span class="truncate max-w-[200px] md:max-w-md">{{ editingMsg.content }}</span>
          </div>
          <button @click="cancelEdit" class="text-gray-500 hover:text-gray-700 p-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>
        <form @submit.prevent="handleSend" class="flex items-end space-x-2 w-full">
          <input type="file" ref="fileInput" @change="handleFileUpload" class="hidden" accept="image/*,.pdf,.doc,.docx,.zip,.rar" />
          <button
            type="button"
            @click="triggerFileInput"
            class="p-2.5 text-gray-400 hover:text-sky-500 hover:bg-sky-50 rounded-full transition flex-shrink-0"
            title="แนบไฟล์"
            :disabled="isUploading"
          >
            <svg v-if="!isUploading" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"></path></svg>
            <div v-else class="animate-spin w-5 h-5 border-2 border-sky-500 border-t-transparent rounded-full"></div>
          </button>

          <textarea
            v-model="inputMessage"
            @input="handleTyping"
            :placeholder="editingMsg ? 'แก้ไขข้อความ...' : 'พิมพ์ข้อความ...'"
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

import Swal from 'sweetalert2'

const emit = defineEmits(['back'])

const authStore = useAuthStore()
const friendsStore = useFriendsStore()
const chatStore = useChatStore()

const inputMessage = ref('')
const messagesContainer = ref(null)
const loadingMore = ref(false)
const allLoaded = ref(false)
const fileInput = ref(null)
const isUploading = ref(false)
const editingMsg = ref(null)
let typingThrottleTimer = null

const selectedFriend = computed(() => {
  if (!chatStore.selectedFriendId) return null
  return friendsStore.friends.find(f => f.id === chatStore.selectedFriendId) || {
    id: chatStore.selectedFriendId,
    username: `User #${chatStore.selectedFriendId}`
  }
})

const friendStatus = computed(() => {
  if (!chatStore.selectedFriendId) return {}
  return chatStore.userStatus[chatStore.selectedFriendId] || {}
})

function formatLastSeen(timestamp) {
  if (!timestamp) return ''
  try {
    const date = new Date(timestamp)
    const now = new Date()
    if (date.toDateString() === now.toDateString()) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    }
    return date.toLocaleDateString() + ' ' + date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  } catch(e) { return '' }
}

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

async function handleScroll() {
  if (!messagesContainer.value || !chatStore.selectedFriendId) return
  
  if (messagesContainer.value.scrollTop === 0 && !loadingMore.value && !allLoaded.value) {
    const messages = chatStore.activeMessages
    if (messages.length > 0) {
      const oldestMessageId = messages[0].id
      if (oldestMessageId) {
        loadingMore.value = true
        const previousScrollHeight = messagesContainer.value.scrollHeight
        const previousCount = messages.length
        
        await chatStore.fetchMessages(chatStore.selectedFriendId, oldestMessageId)
        
        if (chatStore.activeMessages.length === previousCount) {
          allLoaded.value = true
        }

        nextTick(() => {
          if (messagesContainer.value && chatStore.activeMessages.length > previousCount) {
            // maintain scroll position only if new messages were added
            messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight - previousScrollHeight
          }
        })
        loadingMore.value = false
      }
    }
  }
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

function triggerFileInput() {
  if (fileInput.value) {
    fileInput.value.click()
  }
}

async function handleFileUpload(event) {
  const file = event.target.files[0]
  if (!file || !chatStore.selectedFriendId) return
  
  if (file.size > 5 * 1024 * 1024) {
    Swal.fire({ icon: 'error', title: 'ไฟล์ใหญ่เกินไป', text: 'ขนาดไฟล์ต้องไม่เกิน 5MB' })
    return
  }

  isUploading.value = true
  const formData = new FormData()
  formData.append('file', file)

  try {
    const response = await fetch('/api/messages/upload', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${authStore.token}` },
      body: formData
    })
    if (response.ok) {
      const data = await response.json()
      const isImage = file.type.startsWith('image/')
      chatStore.sendMessage(chatStore.selectedFriendId, file.name, isImage ? 'image' : 'file', data.file_url)
      scrollToBottom()
    } else {
      throw new Error('อัปโหลดไม่สำเร็จ')
    }
  } catch (err) {
    Swal.fire({ icon: 'error', title: 'เกิดข้อผิดพลาด', text: err.message })
  } finally {
    isUploading.value = false
    if (fileInput.value) fileInput.value.value = null
  }
}

function startEdit(msg) {
  editingMsg.value = { ...msg }
  inputMessage.value = msg.content
}

function cancelEdit() {
  editingMsg.value = null
  inputMessage.value = ''
}

async function confirmDelete(msg) {
  const result = await Swal.fire({
    title: 'ต้องการลบข้อความนี้?',
    text: "ลบแล้วจะไม่สามารถกู้คืนได้",
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#d33',
    cancelButtonColor: '#3085d6',
    confirmButtonText: 'ลบเลย',
    cancelButtonText: 'ยกเลิก'
  })
  if (result.isConfirmed) {
    chatStore.deleteMessage(msg.id, chatStore.selectedFriendId)
  }
}

function handleSend() {
  if (!inputMessage.value.trim() || !chatStore.selectedFriendId) return
  if (editingMsg.value) {
    chatStore.editMessage(editingMsg.value.id, chatStore.selectedFriendId, inputMessage.value.trim())
    cancelEdit()
  } else {
    chatStore.sendMessage(chatStore.selectedFriendId, inputMessage.value)
    inputMessage.value = ''
  }
  scrollToBottom()
}

watch(
  () => chatStore.activeMessages,
  (newVal, oldVal) => {
    // Only scroll to bottom if we are not paginating up
    if (newVal.length > 0 && oldVal.length > 0) {
      if (newVal[newVal.length - 1].id !== oldVal[oldVal.length - 1]?.id) {
        scrollToBottom()
      }
    } else {
      scrollToBottom()
    }
  },
  { deep: true }
)

watch(
  () => chatStore.selectedFriendId,
  () => {
    allLoaded.value = false
    scrollToBottom()
  }
)
</script>

