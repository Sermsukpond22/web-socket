<template>
  <div class="w-full h-full flex flex-col bg-white dark:bg-gray-900">
    <!-- EMPTY STATE WHEN NO FRIEND OR ROOM SELECTED -->
    <div v-if="!selectedFriend && !selectedRoom" class="h-full flex flex-col items-center justify-center p-8 text-center bg-gray-50 dark:bg-gray-800">
      <div class="w-24 h-24 rounded-full border-2 border-gray-900 flex items-center justify-center text-gray-900 dark:text-gray-100 mb-4 shadow-sm">
        <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
        </svg>
      </div>
      <h2 class="text-xl font-light text-gray-900 dark:text-gray-100 mb-1">ข้อความของคุณ</h2>
      <p class="text-sm text-gray-500 dark:text-gray-400 max-w-xs mb-6">ส่งข้อความและแชทกับเพื่อนหรือกลุ่มของคุณแบบเรียลไทม์</p>
    </div>

    <!-- ACTIVE CHAT WINDOW -->
    <div v-else class="h-full flex flex-col">
      <!-- Chat Header -->
      <div class="px-4 py-3 border-b border-gray-200 flex items-center justify-between bg-white dark:bg-gray-900 z-10">
        <div class="flex items-center space-x-3">
          <!-- Mobile Back Button -->
          <button 
            @click="$emit('back')"
            class="md:hidden text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white dark:text-gray-100 p-1.5 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800 transition mr-1"
            title="กลับไปหน้าแชท"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7" />
            </svg>
          </button>

          <!-- Friend/Room Avatar -->
          <div class="relative cursor-pointer" @click="openFriendProfile">
            <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-sm shadow-sm overflow-hidden">
              <img v-if="currentHeader.avatar_url" :src="getFullUrl(currentHeader.avatar_url)" class="w-full h-full object-cover" />
              <span v-else>{{ getInitial(currentHeader.name) }}</span>
            </div>
            <span v-if="selectedFriend" class="absolute bottom-0 right-0 w-3 h-3 rounded-full border-2 border-white" :class="friendStatus.is_online ? 'bg-green-500' : 'bg-gray-400'"></span>
          </div>

          <!-- Info -->
          <div>
            <h3 class="font-semibold text-sm text-gray-900 dark:text-gray-100 leading-tight">
              {{ currentHeader.name }}
            </h3>
            <div v-if="selectedFriend">
              <p v-if="chatStore.isUserTyping(chatStore.selectedFriendId)" class="text-xs text-sky-500 font-medium animate-pulse">กำลังพิมพ์...</p>
              <p v-else-if="friendStatus.is_online" class="text-xs text-green-600 font-medium">ออนไลน์</p>
              <p v-else-if="friendStatus.last_seen" class="text-xs text-gray-400 font-medium">ใช้งานล่าสุด: {{ formatLastSeen(friendStatus.last_seen) }}</p>
              <p v-else class="text-xs text-gray-400 font-medium">ออฟไลน์</p>
            </div>
            <div v-else-if="selectedRoom">
              <p class="text-xs text-gray-400 font-medium">กลุ่มสนทนา</p>
            </div>
          </div>
        </div>
        
        <!-- Group Actions -->
        <div v-if="selectedRoom" class="flex items-center">
          <button @click="handleInviteMember" class="text-xs bg-sky-500 hover:bg-sky-600 text-white py-1.5 px-3 rounded-md transition font-semibold flex items-center space-x-1 shadow-sm">
            <span>+ เชิญเพื่อน</span>
          </button>
        </div>
      </div>

      <!-- Messages Thread -->
      <div 
        ref="messagesContainer" 
        class="flex-1 overflow-y-auto p-4 space-y-3 bg-white dark:bg-gray-900"
        @scroll="handleScroll"
      >
        <div v-if="loadingMore" class="flex justify-center my-2 text-sky-500">
          <div class="animate-spin rounded-full h-5 w-5 border-b-2 border-sky-500"></div>
        </div>
        <div 
          v-if="chatStore.activeMessages.length === 0" 
          class="h-full flex items-center justify-center text-center text-gray-400 text-sm"
        >
          ยังไม่มีข้อความ ทักทายเลยสิ!
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
            <!-- Sender Avatar (for groups) -->
            <div v-if="!isSelf(msg) && selectedRoom" class="w-8 h-8 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex flex-shrink-0 items-center justify-center font-bold text-xs shadow-sm overflow-hidden mr-1 mt-1">
              <img v-if="msg.sender?.avatar_url" :src="getFullUrl(msg.sender.avatar_url)" class="w-full h-full object-cover" />
              <span v-else>{{ getInitial(msg.sender?.username || msg.sender?.email) }}</span>
            </div>

            <!-- Message Content -->
            <div class="flex flex-col">
              <span v-if="!isSelf(msg) && selectedRoom" class="text-[10px] text-gray-500 mb-0.5 ml-1">{{ msg.sender?.username || msg.sender?.email || 'User' }}</span>
              <div
                :class="[
                  'px-4 py-2 text-sm rounded-2xl break-words shadow-sm flex flex-col',
                  msg.is_deleted ? 'bg-gray-100 dark:bg-gray-800 text-gray-400 italic border border-gray-200' :
                  isSelf(msg) ? 'bg-sky-500 text-white rounded-br-sm' : 'bg-gray-100 dark:bg-gray-800 text-gray-900 dark:text-gray-100 rounded-bl-sm'
                ]"
              >
                <!-- Quoted Message Block -->
                <div v-if="msg.reply_to" class="mb-2 p-2 rounded-lg text-xs border-l-4 opacity-80" :class="isSelf(msg) ? 'bg-white/20 border-white' : 'bg-black/5 dark:bg-white/10 border-sky-500'">
                  <div v-if="msg.reply_to.is_deleted" class="italic">ข้อความนี้ถูกลบแล้ว</div>
                  <div v-else>
                    <span class="font-bold mr-1">{{ msg.reply_to.sender?.username || 'User' }}:</span>
                    <span class="truncate max-w-[150px] inline-block align-bottom">{{ msg.reply_to.content || 'ไฟล์แนบ' }}</span>
                  </div>
                </div>

                <template v-if="msg.is_deleted">
                  🚫 ข้อความนี้ถูกลบแล้ว
                </template>
                <template v-else>
                  <!-- Image -->
                  <div v-if="msg.type === 'image' && msg.file_url">
                    <img :src="apiUrl + msg.file_url" alt="image" class="rounded-lg max-w-full h-auto cursor-pointer object-cover mb-1" />
                  </div>
                  <!-- File -->
                  <div v-else-if="msg.type === 'file' && msg.file_url" class="flex items-center space-x-2 bg-black/10 p-2 rounded-lg mb-1">
                    <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15.172 7l-6.586 6.586a2 2 0 102.828 2.828l6.414-6.586a4 4 0 00-5.656-5.656l-6.415 6.585a6 6 0 108.486 8.486L20.5 13"></path></svg>
                    <a :href="apiUrl + msg.file_url" target="_blank" class="underline text-sm font-semibold truncate">{{ msg.content || 'ไฟล์แนบ' }}</a>
                  </div>
                  <!-- Text / Fallback -->
                  <div v-if="msg.content && msg.type !== 'file'">
                    {{ msg.content }}
                  </div>
                </template>
              </div>
            </div>
            
            <!-- Actions (Edit/Delete/Reply) -->
            <div v-if="!msg.is_deleted && msg.id && !msg.pending" class="opacity-0 group-hover:opacity-100 transition flex flex-row items-center space-x-1 self-center">
              <button @click="startReply(msg)" class="text-gray-400 hover:text-sky-500 p-1.5 bg-gray-50 dark:bg-gray-800 rounded-full" title="ตอบกลับ">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 10h10a8 8 0 018 8v2M3 10l6 6m-6-6l6-6"></path></svg>
              </button>
              <button v-if="isSelf(msg)" @click="startEdit(msg)" class="text-gray-400 hover:text-sky-500 p-1.5 bg-gray-50 dark:bg-gray-800 rounded-full" title="แก้ไข">
                <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path></svg>
              </button>
              <button v-if="isSelf(msg)" @click="confirmDelete(msg)" class="text-gray-400 hover:text-red-500 p-1.5 bg-gray-50 dark:bg-gray-800 rounded-full" title="ลบ">
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
          <div class="px-4 py-3 bg-gray-100 dark:bg-gray-800 rounded-2xl rounded-bl-sm w-[72px] shadow-sm">
            <div class="flex space-x-1.5 items-center justify-center h-full">
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 0ms"></div>
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 150ms"></div>
              <div class="w-1.5 h-1.5 bg-gray-500 rounded-full animate-bounce" style="animation-delay: 300ms"></div>
            </div>
          </div>
        </div>
      </div>

      <!-- Input Area -->
      <div class="p-3 border-t border-gray-200 bg-white dark:bg-gray-900 w-full">
        <div v-if="editingMsg" class="flex items-center justify-between bg-sky-50 dark:bg-sky-900/30 p-2 rounded-t-lg border border-sky-100 border-b-0">
          <div class="text-xs text-sky-700 flex flex-col">
            <span class="font-bold">แก้ไขข้อความ</span>
            <span class="truncate max-w-[200px] md:max-w-md">{{ editingMsg.content }}</span>
          </div>
          <button @click="cancelEdit" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 p-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>
        <div v-if="replyingToMessage" class="flex items-center justify-between bg-gray-50 dark:bg-gray-800 p-2 rounded-t-lg border border-gray-200 border-b-0">
          <div class="text-xs text-gray-700 dark:text-gray-300 flex flex-col">
            <span class="font-bold text-sky-600">ตอบกลับถึง {{ replyingToMessage.sender?.username || 'User' }}</span>
            <span class="truncate max-w-[200px] md:max-w-md italic">{{ replyingToMessage.content || 'ไฟล์แนบ' }}</span>
          </div>
          <button @click="cancelReply" class="text-gray-500 dark:text-gray-400 hover:text-gray-700 p-1">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>
          </button>
        </div>
        <form @submit.prevent="handleSend" class="flex items-end space-x-2 w-full">
          <input type="file" ref="fileInput" @change="handleFileUpload" class="hidden" accept="image/*,.pdf,.doc,.docx,.zip,.rar" />
          <div class="relative flex-shrink-0">
            <button
              type="button"
              @click="showEmojiPicker = !showEmojiPicker"
              class="p-2.5 text-gray-400 hover:text-sky-500 hover:bg-sky-50 dark:bg-sky-900/30 rounded-full transition"
              title="อีโมจิ"
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.828 14.828a4 4 0 01-5.656 0M9 10h.01M15 10h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>
            </button>
            <div v-if="showEmojiPicker" class="absolute bottom-full mb-2 left-0 bg-white dark:bg-gray-900 border border-gray-200 rounded-lg shadow-lg p-2 flex flex-wrap w-48 z-50">
              <span v-for="emoji in emojis" :key="emoji" @click="insertEmoji(emoji)" class="cursor-pointer p-1.5 hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800 rounded text-xl">
                {{ emoji }}
              </span>
            </div>
          </div>

          <textarea
            v-model="inputMessage"
            @input="handleTyping"
            :placeholder="editingMsg ? 'แก้ไขข้อความ...' : 'พิมพ์ข้อความ...'"
            rows="1"
            class="flex-1 px-4 py-2.5 bg-gray-100 dark:bg-gray-800 rounded-2xl text-sm text-gray-900 dark:text-gray-100 placeholder-gray-500 focus:outline-none focus:bg-gray-50 dark:bg-gray-800 focus:ring-1 focus:ring-sky-500 border border-transparent focus:border-sky-500 transition resize-none overflow-hidden min-h-[44px] max-h-[120px]"
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
  
  <ProfileModal 
    :is-open="isProfileOpen" 
    :user="selectedProfileUser" 
    :is-own-profile="false"
    @close="isProfileOpen = false" 
  />
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useFriendsStore } from '../stores/friends'
import { useChatStore } from '../stores/chat'
import { useRoomsStore } from '../stores/rooms'
import ProfileModal from './ProfileModal.vue'

import Swal from 'sweetalert2'

const apiUrl = import.meta.env.VITE_API_URL || ''

const emit = defineEmits(['back'])

const authStore = useAuthStore()
const friendsStore = useFriendsStore()
const chatStore = useChatStore()
const roomsStore = useRoomsStore()

const inputMessage = ref('')
const messagesContainer = ref(null)
const loadingMore = ref(false)
const allLoaded = ref(false)
const fileInput = ref(null)
const isUploading = ref(false)
const editingMsg = ref(null)
const replyingToMessage = ref(null)
const showEmojiPicker = ref(false)
const emojis = ['😀','😂','🥰','😎','😭','😡','👍','🙏','❤️','🔥','✨','🎉']
let typingThrottleTimer = null

// Profile State
const isProfileOpen = ref(false)
const selectedProfileUser = ref({})

function getFullUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${import.meta.env.VITE_API_URL}${url}`
}

async function openFriendProfile() {
  if (!chatStore.selectedFriendId) return
  try {
    const res = await fetch(`${import.meta.env.VITE_API_URL}/api/users/${chatStore.selectedFriendId}/profile`, {
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    })
    if (res.ok) {
      selectedProfileUser.value = await res.json()
      isProfileOpen.value = true
    }
  } catch (err) {
    console.error('Failed to fetch profile', err)
  }
}

const selectedFriend = computed(() => {
  if (!chatStore.selectedFriendId) return null
  return friendsStore.friends.find(f => f.id === chatStore.selectedFriendId) || {
    id: chatStore.selectedFriendId,
    username: `User #${chatStore.selectedFriendId}`
  }
})

const selectedRoom = computed(() => {
  if (!chatStore.selectedRoomId) return null
  return roomsStore.rooms.find(r => r.id === chatStore.selectedRoomId) || {
    id: chatStore.selectedRoomId,
    name: `Room #${chatStore.selectedRoomId}`
  }
})

const currentHeader = computed(() => {
  if (selectedFriend.value) {
    return { name: selectedFriend.value.username || selectedFriend.value.email, avatar_url: selectedFriend.value.avatar_url }
  } else if (selectedRoom.value) {
    return { name: selectedRoom.value.name, avatar_url: selectedRoom.value.avatar_url }
  }
  return { name: 'Chat', avatar_url: '' }
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
  if (!messagesContainer.value) return
  if (!chatStore.selectedFriendId && !chatStore.selectedRoomId) return
  
  if (messagesContainer.value.scrollTop === 0 && !loadingMore.value && !allLoaded.value) {
    const messages = chatStore.activeMessages
    if (messages.length > 0) {
      const oldestMessageId = messages[0].id
      if (oldestMessageId) {
        loadingMore.value = true
        const previousScrollHeight = messagesContainer.value.scrollHeight
        const previousCount = messages.length
        
        if (chatStore.selectedFriendId) {
          await chatStore.fetchMessages(chatStore.selectedFriendId, oldestMessageId)
        } else if (chatStore.selectedRoomId) {
          await chatStore.fetchRoomMessages(chatStore.selectedRoomId, oldestMessageId)
        }
        
        if (chatStore.activeMessages.length === previousCount) {
          allLoaded.value = true
        }

        nextTick(() => {
          if (messagesContainer.value && chatStore.activeMessages.length > previousCount) {
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

function insertEmoji(emoji) {
  inputMessage.value += emoji
  showEmojiPicker.value = false
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
      chatStore.sendMessage(
        chatStore.selectedFriendId, 
        file.name, 
        isImage ? 'image' : 'file', 
        data.file_url, 
        chatStore.selectedRoomId,
        replyingToMessage.value ? replyingToMessage.value.id : null
      )
      replyingToMessage.value = null
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

function startReply(msg) {
  replyingToMessage.value = { ...msg }
  inputMessage.value = ''
}

function cancelReply() {
  replyingToMessage.value = null
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
    chatStore.deleteMessage(msg.id, chatStore.selectedFriendId, chatStore.selectedRoomId)
  }
}

function handleSend() {
  if (!inputMessage.value.trim() && !fileInput.value?.files?.length) return
  if (!chatStore.selectedFriendId && !chatStore.selectedRoomId) return
  
  if (editingMsg.value) {
    chatStore.editMessage(editingMsg.value.id, chatStore.selectedFriendId, inputMessage.value.trim())
    cancelEdit()
  } else {
    chatStore.sendMessage(
      chatStore.selectedFriendId, 
      inputMessage.value, 
      'text', 
      '', 
      chatStore.selectedRoomId,
      replyingToMessage.value ? replyingToMessage.value.id : null
    )
    replyingToMessage.value = null
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

watch(
  () => chatStore.selectedRoomId,
  () => {
    allLoaded.value = false
    scrollToBottom()
  }
)

async function handleInviteMember() {
  if (!chatStore.selectedRoomId) return

  // Create a list of friends to select from
  const friendOptions = {}
  friendsStore.friends.forEach(f => {
    friendOptions[f.id] = f.username || f.email
  })

  if (Object.keys(friendOptions).length === 0) {
    Swal.fire('ไม่มีเพื่อน', 'คุณต้องมีเพื่อนก่อนถึงจะเชิญเข้ากลุ่มได้', 'warning')
    return
  }

  const { value: selectedUserId } = await Swal.fire({
    title: 'เชิญเพื่อนเข้ากลุ่ม',
    input: 'select',
    inputOptions: friendOptions,
    inputPlaceholder: 'เลือกเพื่อนที่ต้องการเชิญ...',
    showCancelButton: true,
    confirmButtonText: 'เชิญ',
    cancelButtonText: 'ยกเลิก',
    inputValidator: (value) => {
      if (!value) return 'กรุณาเลือกเพื่อน'
    }
  })

  if (selectedUserId) {
    try {
      await roomsStore.inviteToRoom(chatStore.selectedRoomId, parseInt(selectedUserId))
      Swal.fire({ icon: 'success', title: 'ส่งคำเชิญแล้ว', toast: true, position: 'top-end', showConfirmButton: false, timer: 3000 })
    } catch (err) {
      Swal.fire('เกิดข้อผิดพลาด', err.message || 'ไม่สามารถส่งคำเชิญได้', 'error')
    }
  }
}
</script>
