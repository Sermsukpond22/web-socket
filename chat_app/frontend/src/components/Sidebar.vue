<template>
  <div class="h-full flex flex-col bg-white dark:bg-gray-900 border-r border-gray-200 select-none">
    <!-- Header / User Profile -->
    <div class="p-4 border-b border-gray-200 flex items-center justify-between">
      <div class="flex items-center space-x-3 overflow-hidden">
        <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-purple-500 to-indigo-600 text-white flex items-center justify-center font-bold text-lg flex-shrink-0 shadow-sm cursor-pointer overflow-hidden" @click="openOwnProfile">
          <img v-if="authStore.user?.avatar_url" :src="getFullUrl(authStore.user.avatar_url)" class="w-full h-full object-cover" />
          <span v-else>{{ userInitial }}</span>
        </div>
        <div class="truncate">
          <h2 class="font-semibold text-gray-900 dark:text-gray-100 text-sm truncate">{{ authStore.user?.username || 'User' }}</h2>
          <p class="text-xs text-gray-500 dark:text-gray-400 truncate">{{ authStore.user?.email || '' }}</p>
        </div>
      </div>
      <!-- Actions / Logout -->
      <div class="flex items-center space-x-1">
        <button 
          @click="toggleDarkMode"
          title="สลับโหมดมืด"
          class="text-gray-500 dark:text-gray-400 hover:text-sky-600 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800 transition duration-150"
        >
          <svg v-if="isDark" class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 3v1m0 16v1m9-9h-1M4 12H3m15.364 6.364l-.707-.707M6.343 6.343l-.707-.707m12.728 0l-.707.707M6.343 17.657l-.707.707M16 12a4 4 0 11-8 0 4 4 0 018 0z"></path></svg>
          <svg v-else class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20.354 15.354A9 9 0 018.646 3.646 9.003 9.003 0 0012 21a9.003 9.003 0 008.354-5.646z"></path></svg>
        </button>
        <button 
          @click="handleLogout"
          title="ออกจากระบบ"
          class="text-gray-500 dark:text-gray-400 hover:text-red-600 p-2 rounded-full hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800 transition duration-150"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
          </svg>
        </button>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="p-3 border-b border-gray-100">
      <div class="relative">
        <svg class="w-4 h-4 absolute left-3 top-3 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
        </svg>
        <input
          v-model="searchQuery"
          type="text"
          placeholder="ค้นหาแชท..."
          class="w-full pl-9 pr-3 py-2 bg-gray-100 dark:bg-gray-800 rounded-lg text-sm text-gray-900 dark:text-gray-100 placeholder-gray-500 focus:outline-none focus:bg-gray-50 dark:bg-gray-800 focus:ring-1 focus:ring-sky-500"
        />
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex border-b border-gray-200 bg-gray-50 dark:bg-gray-800 text-xs font-semibold text-gray-600 dark:text-gray-300">
      <button
        @click="friendsStore.activeTab = 'messages'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2',
          friendsStore.activeTab === 'messages' 
            ? 'border-sky-500 text-sky-600 bg-white dark:bg-gray-900' 
            : 'border-transparent hover:text-gray-900 dark:hover:text-white dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800'
        ]"
      >
        ข้อความ
      </button>
      <button
        @click="friendsStore.activeTab = 'groups'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2',
          friendsStore.activeTab === 'groups' 
            ? 'border-sky-500 text-sky-600 bg-white dark:bg-gray-900' 
            : 'border-transparent hover:text-gray-900 dark:hover:text-white dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800'
        ]"
      >
        กลุ่ม
      </button>
      <button
        @click="friendsStore.activeTab = 'pending'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2 relative',
          friendsStore.activeTab === 'pending' 
            ? 'border-sky-500 text-sky-600 bg-white dark:bg-gray-900' 
            : 'border-transparent hover:text-gray-900 dark:hover:text-white dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800'
        ]"
      >
        คำขอ
        <span 
          v-if="friendsStore.pendingCount > 0"
          class="ml-1 px-1.5 py-0.5 text-[10px] bg-red-500 text-white rounded-full font-bold"
        >
          {{ friendsStore.pendingCount }}
        </span>
      </button>
      <button
        @click="friendsStore.activeTab = 'add'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2',
          friendsStore.activeTab === 'add' 
            ? 'border-sky-500 text-sky-600 bg-white dark:bg-gray-900' 
            : 'border-transparent hover:text-gray-900 dark:hover:text-white dark:text-gray-100 hover:bg-gray-100 dark:hover:bg-gray-700 dark:bg-gray-800'
        ]"
      >
        + เพิ่มเพื่อน
      </button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 overflow-y-auto">
      <!-- MESSAGES TAB -->
      <div v-if="friendsStore.activeTab === 'messages'">
        <div v-if="filteredFriends.length === 0" class="p-8 text-center text-gray-500 dark:text-gray-400 text-sm">
          <p v-if="searchQuery">ไม่พบเพื่อนที่ค้นหา "{{ searchQuery }}"</p>
          <p v-else>ยังไม่มีเพื่อน<br/>เพิ่มเพื่อนได้ที่แท็บ "+ เพิ่มเพื่อน"</p>
        </div>

        <div v-else class="divide-y divide-gray-50 dark:divide-gray-800">
          <div
            v-for="friend in filteredFriends"
            :key="friend.id"
            @click="selectFriend(friend.id)"
            :class="[
              'p-3.5 flex items-center space-x-3 cursor-pointer transition duration-150 hover:bg-gray-50 dark:hover:bg-gray-800 dark:bg-gray-800',
              chatStore.selectedFriendId === friend.id ? 'bg-sky-50 dark:bg-sky-900/30 hover:bg-sky-50 dark:bg-sky-900/30' : ''
            ]"
          >
            <!-- Avatar with Status Dot -->
            <div class="relative flex-shrink-0">
              <div class="w-12 h-12 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-base shadow-sm overflow-hidden" @click.stop="openFriendProfile(friend)">
                <img v-if="friend.avatar_url" :src="getFullUrl(friend.avatar_url)" class="w-full h-full object-cover" />
                <span v-else>{{ getInitial(friend.username || friend.email) }}</span>
              </div>
              <span 
                class="absolute bottom-0 right-0 w-3.5 h-3.5 rounded-full border-2 border-white"
                :class="chatStore.userStatus[friend.id]?.is_online ? 'bg-green-500' : 'bg-gray-400'"
                :title="chatStore.userStatus[friend.id]?.is_online ? 'ออนไลน์' : 'ออฟไลน์'"
              ></span>
            </div>

            <!-- Details -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <h3 :class="['text-sm truncate', chatStore.unreadCounts[friend.id] ? 'font-bold text-sky-600' : 'font-semibold text-gray-900 dark:text-gray-100']">
                  {{ friend.username || friend.email }}
                </h3>
                <span v-if="chatStore.unreadCounts[friend.id]" class="ml-2 inline-flex items-center justify-center px-2 py-0.5 text-[10px] font-bold leading-none text-white bg-red-500 rounded-full">
                  {{ chatStore.unreadCounts[friend.id] > 99 ? '99+' : chatStore.unreadCounts[friend.id] }}
                </span>
              </div>
              <p :class="['text-xs truncate', chatStore.unreadCounts[friend.id] ? 'font-bold text-gray-800 dark:text-gray-200' : 'text-gray-500 dark:text-gray-400']">
                {{ getLastMessageText(friend.id) || friend.email || 'คลิกเพื่อเริ่มแชท' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- GROUPS TAB -->
      <div v-else-if="friendsStore.activeTab === 'groups'">
        <div class="p-4 border-b border-gray-100 dark:border-gray-800">
          <button @click="handleCreateGroup" class="w-full bg-sky-500 hover:bg-sky-600 text-white text-sm font-semibold py-2 rounded-md transition flex items-center justify-center space-x-2">
            <span>+ สร้างกลุ่มใหม่</span>
          </button>
        </div>
        
        <!-- Pending Group Invites -->
        <div v-if="roomsStore.pendingInvites.length > 0" class="px-4 py-2 bg-yellow-50 dark:bg-yellow-900/20 border-b border-yellow-100 dark:border-yellow-900/30">
          <h4 class="text-xs font-semibold text-yellow-800 dark:text-yellow-500 mb-2">คำเชิญเข้ากลุ่ม ({{ roomsStore.pendingInvites.length }})</h4>
          <div class="space-y-2">
            <div v-for="inv in roomsStore.pendingInvites" :key="inv.room_id" class="flex items-center justify-between bg-white dark:bg-gray-800 p-2 rounded shadow-sm">
              <div class="flex items-center space-x-2">
                <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-green-400 to-teal-500 text-white flex items-center justify-center font-bold text-xs flex-shrink-0">
                  <img v-if="inv.room?.avatar_url" :src="getFullUrl(inv.room.avatar_url)" class="w-full h-full object-cover rounded-full" />
                  <span v-else>{{ getInitial(inv.room?.name) }}</span>
                </div>
                <div class="truncate max-w-[100px]">
                  <p class="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">{{ inv.room?.name || 'Group' }}</p>
                </div>
              </div>
              <button @click="roomsStore.acceptInvite(inv.room_id)" class="px-2 py-1 bg-sky-500 hover:bg-sky-600 text-white text-[10px] font-semibold rounded">
                ยอมรับ
              </button>
            </div>
          </div>
        </div>

        <!-- Groups List -->
        <div v-if="roomsStore.rooms.length === 0" class="p-8 text-center text-gray-500 dark:text-gray-400 text-sm">
          <p>คุณยังไม่มีกลุ่ม<br/>สร้างกลุ่มหรือรอคำเชิญจากเพื่อน</p>
        </div>
        <div v-else class="divide-y divide-gray-50 dark:divide-gray-800">
          <div
            v-for="room in roomsStore.rooms"
            :key="room.id"
            @click="chatStore.selectRoom(room.id)"
            :class="[
              'p-3.5 flex items-center space-x-3 cursor-pointer transition duration-150 hover:bg-gray-50 dark:hover:bg-gray-800 dark:bg-gray-800',
              chatStore.selectedRoomId === room.id ? 'bg-sky-50 dark:bg-sky-900/30 hover:bg-sky-50 dark:bg-sky-900/30' : ''
            ]"
          >
            <div class="relative flex-shrink-0">
              <div class="w-12 h-12 rounded-full bg-gradient-to-tr from-emerald-400 to-green-600 text-white flex items-center justify-center font-bold text-base shadow-sm overflow-hidden">
                <img v-if="room.avatar_url" :src="getFullUrl(room.avatar_url)" class="w-full h-full object-cover" />
                <span v-else>{{ getInitial(room.name) }}</span>
              </div>
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex justify-between items-baseline mb-0.5">
                <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 truncate">
                  {{ room.name }}
                </h3>
                <span class="text-[10px] text-gray-400 font-medium">
                  {{ formatLastRoomMessageTime(room.id) }}
                </span>
              </div>
              <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">
                {{ getLastRoomMessagePreview(room.id) }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- PENDING REQUESTS TAB -->
      <div v-else-if="friendsStore.activeTab === 'pending'" class="p-4">
        <div v-if="friendsStore.pendingRequests.length === 0" class="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
          ไม่มีคำขอเป็นเพื่อน
        </div>

        <div v-else class="space-y-3">
          <div 
            v-for="req in friendsStore.pendingRequests"
            :key="req.request_id || req.id"
            class="p-3 bg-gray-50 dark:bg-gray-800 border border-gray-200 rounded-lg flex items-center justify-between"
          >
            <div class="flex items-center space-x-3 overflow-hidden">
              <div class="w-9 h-9 rounded-full bg-gradient-to-tr from-amber-400 to-orange-500 text-white flex items-center justify-center font-bold text-sm overflow-hidden cursor-pointer" @click.stop="openFriendProfile(req.from_user)">
                <img v-if="req.from_user?.avatar_url" :src="getFullUrl(req.from_user.avatar_url)" class="w-full h-full object-cover" />
                <span v-else>{{ getInitial(req.from_user?.username || req.from_user?.email || 'U') }}</span>
              </div>
              <div class="truncate">
                <p class="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">
                  {{ req.from_user?.username || req.from_user?.email || 'User #' + req.from_user_id }}
                </p>
                <p class="text-[11px] text-gray-500 dark:text-gray-400">ส่งคำขอเป็นเพื่อนมาให้คุณ</p>
              </div>
            </div>

            <div class="flex items-center space-x-2">
              <button
                @click="acceptRequest(req.request_id || req.id)"
                class="px-3 py-1.5 bg-sky-500 hover:bg-sky-600 text-white text-xs font-semibold rounded-md transition shadow-sm"
              >
                ยอมรับ
              </button>
              <button
                @click="rejectRequest(req.request_id || req.id)"
                class="px-3 py-1.5 bg-gray-200 hover:bg-red-500 hover:text-white text-gray-700 text-xs font-semibold rounded-md transition shadow-sm"
              >
                ปฏิเสธ
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- ADD FRIEND TAB -->
      <div v-else-if="friendsStore.activeTab === 'add'" class="p-4 space-y-4">
        <div>
          <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100 mb-1">เพิ่มเพื่อน</h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">ระบุชื่อผู้ใช้เพื่อส่งคำขอเป็นเพื่อน</p>

          <form @submit.prevent="handleSendRequest" class="space-y-3">
            <input
              v-model="addFriendInput"
              @input="handleSearchInput"
              type="text"
              required
              placeholder="ชื่อผู้ใช้ หรืออีเมล"
              class="w-full px-3 py-2 bg-gray-50 dark:bg-gray-800 border border-gray-300 rounded-md text-sm text-gray-900 dark:text-gray-100 focus:outline-none focus:ring-1 focus:ring-sky-500"
            />
            <button
              type="submit"
              :disabled="friendsStore.loading || !addFriendInput.trim()"
              class="w-full bg-sky-500 hover:bg-sky-600 disabled:opacity-50 text-white text-xs font-semibold py-2 rounded-md transition"
            >
              ส่งคำขอ
            </button>
          </form>

          <!-- Autocomplete Suggestions List -->
          <div v-if="searchResults.length > 0" class="mt-3 border border-gray-200 rounded-md divide-y divide-gray-100 dark:divide-gray-800 max-h-48 overflow-y-auto bg-white dark:bg-gray-900 shadow-sm">
            <div
              v-for="u in searchResults"
              :key="u.id"
              @click="sendRequestToUser(u)"
              class="p-2.5 flex items-center justify-between hover:bg-sky-50 dark:bg-sky-900/30 cursor-pointer transition"
            >
              <div class="flex items-center space-x-2.5 overflow-hidden">
                <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-xs flex-shrink-0 overflow-hidden cursor-pointer" @click.stop="openFriendProfile(u)">
                  <img v-if="u.avatar_url" :src="getFullUrl(u.avatar_url)" class="w-full h-full object-cover" />
                  <span v-else>{{ getInitial(u.username || u.email) }}</span>
                </div>
                <div class="truncate">
                  <p class="text-xs font-semibold text-gray-900 dark:text-gray-100 truncate">{{ u.username || 'User' }}</p>
                  <p class="text-[11px] text-gray-500 dark:text-gray-400 truncate">{{ u.email || '' }}</p>
                </div>
              </div>
              <button
                type="button"
                @click.stop="sendRequestToUser(u)"
                class="px-2.5 py-1 bg-sky-500 hover:bg-sky-600 text-white text-xs font-semibold rounded transition flex-shrink-0"
              >
                เพิ่มเพื่อน
              </button>
            </div>
          </div>
          <div v-else-if="isSearching" class="mt-2 text-xs text-gray-400 text-center">
            กำลังค้นหา...
          </div>

          <!-- Feedback messages -->
          <div v-if="addFriendSuccess" class="mt-3 p-2.5 bg-green-50 text-green-700 text-xs rounded-md border border-green-200">
            {{ addFriendSuccess }}
          </div>
          <div v-if="addFriendError" class="mt-3 p-2.5 bg-red-50 text-red-600 text-xs rounded-md border border-red-200">
            {{ addFriendError }}
          </div>
        </div>
      </div>
    </div>
  </div>
  
  <ProfileModal 
    :is-open="isProfileOpen" 
    :user="selectedProfileUser" 
    :is-own-profile="isOwnProfileMode"
    @close="isProfileOpen = false" 
  />
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFriendsStore } from '../stores/friends'
import { useChatStore } from '../stores/chat'
import { useRoomsStore } from '../stores/rooms'
import Swal from 'sweetalert2'
import ProfileModal from './ProfileModal.vue'

const router = useRouter()
const authStore = useAuthStore()
const friendsStore = useFriendsStore()
const chatStore = useChatStore()
const roomsStore = useRoomsStore()

const searchQuery = ref('')
const addFriendInput = ref('')
const addFriendSuccess = ref('')
const addFriendError = ref('')
const searchResults = ref([])
const isSearching = ref(false)
let searchDebounceTimer = null
const isDark = ref(document.documentElement.classList.contains('dark'))

// Profile State
const isProfileOpen = ref(false)
const selectedProfileUser = ref({})
const isOwnProfileMode = ref(false)

function getFullUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${import.meta.env.VITE_API_URL}${url}`
}

function openOwnProfile() {
  selectedProfileUser.value = authStore.user || {}
  isOwnProfileMode.value = true
  isProfileOpen.value = true
}

async function openFriendProfile(user) {
  if (!user || !user.id) return
  try {
    const res = await fetch(`${import.meta.env.VITE_API_URL}/api/users/${user.id}/profile`, {
      headers: { 'Authorization': `Bearer ${authStore.token}` }
    })
    if (res.ok) {
      selectedProfileUser.value = await res.json()
      isOwnProfileMode.value = false
      isProfileOpen.value = true
    }
  } catch (err) {
    console.error('Failed to fetch profile', err)
  }
}

function toggleDarkMode() {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('theme', 'light')
  }
}

const userInitial = computed(() => {
  const name = authStore.user?.username || authStore.user?.email || 'U'
  return name.charAt(0).toUpperCase()
})

const filteredFriends = computed(() => {
  let list = friendsStore.friends
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.toLowerCase()
    list = list.filter(f => 
      (f.username && f.username.toLowerCase().includes(q)) ||
      (f.email && f.email.toLowerCase().includes(q))
    )
  }

  // Create a copy before sorting to avoid mutating store state directly
  return [...list].sort((a, b) => {
    const msgsA = chatStore.messages[a.id]
    const msgsB = chatStore.messages[b.id]
    
    const timeA = msgsA && msgsA.length > 0 ? new Date(msgsA[msgsA.length - 1].created_at).getTime() : 0
    const timeB = msgsB && msgsB.length > 0 ? new Date(msgsB[msgsB.length - 1].created_at).getTime() : 0
    
    // Sort by latest message descending
    if (timeA !== timeB) {
      return timeB - timeA
    }
    
    // Fallback to unread counts
    const unreadA = chatStore.unreadCounts[a.id] || 0
    const unreadB = chatStore.unreadCounts[b.id] || 0
    return unreadB - unreadA
  })
})

function getInitial(str) {
  if (!str) return 'U'
  return str.charAt(0).toUpperCase()
}

function getLastMessageText(friendId) {
  const msgs = chatStore.messages[friendId]
  if (msgs && msgs.length > 0) {
    return msgs[msgs.length - 1].content
  }
  return null
}

function formatLastRoomMessageTime(roomId) {
  const msgs = chatStore.roomMessages[roomId]
  if (msgs && msgs.length > 0) {
    const timestamp = msgs[msgs.length - 1].created_at
    if (!timestamp) return ''
    try {
      const date = new Date(timestamp)
      const now = new Date()
      if (date.toDateString() === now.toDateString()) {
        return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      }
      return date.toLocaleDateString()
    } catch(e) { return '' }
  }
  return ''
}

function getLastRoomMessagePreview(roomId) {
  const msgs = chatStore.roomMessages[roomId]
  if (msgs && msgs.length > 0) {
    return msgs[msgs.length - 1].content
  }
  return 'กลุ่มสนทนา'
}

function selectFriend(friendId) {
  chatStore.selectFriend(friendId)
}

async function acceptRequest(requestId) {
  try {
    await friendsStore.acceptFriendRequest(requestId)
  } catch (err) {
    console.error('Failed to accept request:', err)
  }
}

async function rejectRequest(requestId) {
  const result = await Swal.fire({
    title: 'ปฏิเสธคำขอ?',
    icon: 'question',
    showCancelButton: true,
    confirmButtonColor: '#ef4444',
    cancelButtonColor: '#d1d5db',
    confirmButtonText: 'ปฏิเสธ',
    cancelButtonText: 'ยกเลิก'
  })
  
  if (result.isConfirmed) {
    try {
      await friendsStore.rejectFriendRequest(requestId)
    } catch (err) {
      console.error('Failed to reject request:', err)
    }
  }
}

function handleSearchInput() {
  addFriendSuccess.value = ''
  addFriendError.value = ''
  if (searchDebounceTimer) clearTimeout(searchDebounceTimer)

  const q = addFriendInput.value.trim()
  if (!q) {
    searchResults.value = []
    isSearching.value = false
    return
  }

  isSearching.value = true
  searchDebounceTimer = setTimeout(async () => {
    try {
      searchResults.value = await friendsStore.searchUsers(q)
    } catch (err) {
      searchResults.value = []
    } finally {
      isSearching.value = false
    }
  }, 300)
}

async function sendRequestToUser(user) {
  addFriendSuccess.value = ''
  addFriendError.value = ''
  try {
    await friendsStore.sendFriendRequest(user.id)
    addFriendSuccess.value = `Friend request sent to "${user.username || user.email}"`
    addFriendInput.value = ''
    searchResults.value = []
  } catch (err) {
    addFriendError.value = err.message || 'Failed to send friend request.'
  }
}

async function handleSendRequest() {
  addFriendSuccess.value = ''
  addFriendError.value = ''
  try {
    await friendsStore.sendFriendRequest(addFriendInput.value.trim())
    addFriendSuccess.value = `Friend request sent to "${addFriendInput.value}"`
    addFriendInput.value = ''
    searchResults.value = []
  } catch (err) {
    addFriendError.value = err.message || 'Failed to send friend request.'
  }
}

async function handleLogout() {
  const result = await Swal.fire({
    title: 'คุณต้องการออกจากระบบใช่หรือไม่?',
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#0ea5e9',
    cancelButtonColor: '#ef4444',
    confirmButtonText: 'ออกจากระบบ',
    cancelButtonText: 'ยกเลิก'
  })

  if (result.isConfirmed) {
    chatStore.disconnectWebSocket()
    authStore.logout()
    router.push('/login')
  }
}

async function handleCreateGroup() {
  const { value: groupName } = await Swal.fire({
    title: 'สร้างกลุ่มใหม่',
    input: 'text',
    inputLabel: 'ชื่อกลุ่ม',
    inputPlaceholder: 'พิมพ์ชื่อกลุ่ม...',
    showCancelButton: true,
    confirmButtonText: 'สร้าง',
    cancelButtonText: 'ยกเลิก',
    inputValidator: (value) => {
      if (!value) {
        return 'กรุณาระบุชื่อกลุ่ม!'
      }
    }
  })

  if (groupName) {
    try {
      await roomsStore.createRoom(groupName)
      Swal.fire({ icon: 'success', title: 'สร้างกลุ่มสำเร็จ', toast: true, position: 'top-end', showConfirmButton: false, timer: 3000 })
    } catch (err) {
      Swal.fire('เกิดข้อผิดพลาด', err.message || 'ไม่สามารถสร้างกลุ่มได้', 'error')
    }
  }
}

onMounted(() => {
  friendsStore.fetchFriends()
  friendsStore.fetchPendingRequests()
  roomsStore.fetchRooms()
  roomsStore.fetchPendingInvites()
})
</script>

