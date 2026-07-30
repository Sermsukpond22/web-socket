<template>
  <div class="h-full flex flex-col bg-white border-r border-gray-200 select-none">
    <!-- Header / User Profile -->
    <div class="p-4 border-b border-gray-200 flex items-center justify-between">
      <div class="flex items-center space-x-3 overflow-hidden">
        <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-purple-500 to-indigo-600 text-white flex items-center justify-center font-bold text-lg flex-shrink-0 shadow-sm">
          {{ userInitial }}
        </div>
        <div class="truncate">
          <h2 class="font-semibold text-gray-900 text-sm truncate">{{ authStore.user?.username || 'User' }}</h2>
          <p class="text-xs text-gray-500 truncate">{{ authStore.user?.email || '' }}</p>
        </div>
      </div>
      <!-- Actions / Logout -->
      <button 
        @click="handleLogout"
        title="ออกจากระบบ"
        class="text-gray-500 hover:text-red-600 p-2 rounded-full hover:bg-gray-100 transition duration-150"
      >
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
        </svg>
      </button>
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
          class="w-full pl-9 pr-3 py-2 bg-gray-100 rounded-lg text-sm text-gray-900 placeholder-gray-500 focus:outline-none focus:bg-gray-50 focus:ring-1 focus:ring-sky-500"
        />
      </div>
    </div>

    <!-- Navigation Tabs -->
    <div class="flex border-b border-gray-200 bg-gray-50 text-xs font-semibold text-gray-600">
      <button
        @click="friendsStore.activeTab = 'messages'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2',
          friendsStore.activeTab === 'messages' 
            ? 'border-sky-500 text-sky-600 bg-white' 
            : 'border-transparent hover:text-gray-900 hover:bg-gray-100'
        ]"
      >
        ข้อความ
      </button>
      <button
        @click="friendsStore.activeTab = 'pending'"
        :class="[
          'flex-1 py-2.5 text-center transition-colors border-b-2 relative',
          friendsStore.activeTab === 'pending' 
            ? 'border-sky-500 text-sky-600 bg-white' 
            : 'border-transparent hover:text-gray-900 hover:bg-gray-100'
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
            ? 'border-sky-500 text-sky-600 bg-white' 
            : 'border-transparent hover:text-gray-900 hover:bg-gray-100'
        ]"
      >
        + เพิ่มเพื่อน
      </button>
    </div>

    <!-- Tab Content -->
    <div class="flex-1 overflow-y-auto">
      <!-- MESSAGES TAB -->
      <div v-if="friendsStore.activeTab === 'messages'">
        <div v-if="filteredFriends.length === 0" class="p-8 text-center text-gray-500 text-sm">
          <p v-if="searchQuery">ไม่พบเพื่อนที่ค้นหา "{{ searchQuery }}"</p>
          <p v-else>ยังไม่มีเพื่อน<br/>เพิ่มเพื่อนได้ที่แท็บ "+ เพิ่มเพื่อน"</p>
        </div>

        <div v-else class="divide-y divide-gray-50">
          <div
            v-for="friend in filteredFriends"
            :key="friend.id"
            @click="selectFriend(friend.id)"
            :class="[
              'p-3.5 flex items-center space-x-3 cursor-pointer transition duration-150 hover:bg-gray-50',
              chatStore.selectedFriendId === friend.id ? 'bg-sky-50 hover:bg-sky-50' : ''
            ]"
          >
            <!-- Avatar with Status Dot -->
            <div class="relative flex-shrink-0">
              <div class="w-12 h-12 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-base shadow-sm">
                {{ getInitial(friend.username || friend.email) }}
              </div>
              <span 
                class="absolute bottom-0 right-0 w-3.5 h-3.5 rounded-full border-2 border-white bg-green-500"
                title="ออนไลน์"
              ></span>
            </div>

            <!-- Details -->
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-0.5">
                <h3 :class="['text-sm truncate', chatStore.unreadCounts[friend.id] ? 'font-bold text-sky-600' : 'font-semibold text-gray-900']">
                  {{ friend.username || friend.email }}
                </h3>
                <span v-if="chatStore.unreadCounts[friend.id]" class="ml-2 inline-flex items-center justify-center px-2 py-0.5 text-[10px] font-bold leading-none text-white bg-red-500 rounded-full">
                  {{ chatStore.unreadCounts[friend.id] > 99 ? '99+' : chatStore.unreadCounts[friend.id] }}
                </span>
              </div>
              <p :class="['text-xs truncate', chatStore.unreadCounts[friend.id] ? 'font-bold text-gray-800' : 'text-gray-500']">
                {{ getLastMessageText(friend.id) || friend.email || 'คลิกเพื่อเริ่มแชท' }}
              </p>
            </div>
          </div>
        </div>
      </div>

      <!-- PENDING REQUESTS TAB -->
      <div v-else-if="friendsStore.activeTab === 'pending'" class="p-4">
        <div v-if="friendsStore.pendingRequests.length === 0" class="text-center py-8 text-gray-500 text-sm">
          ไม่มีคำขอเป็นเพื่อน
        </div>

        <div v-else class="space-y-3">
          <div 
            v-for="req in friendsStore.pendingRequests"
            :key="req.request_id || req.id"
            class="p-3 bg-gray-50 border border-gray-200 rounded-lg flex items-center justify-between"
          >
            <div class="flex items-center space-x-3 overflow-hidden">
              <div class="w-9 h-9 rounded-full bg-gradient-to-tr from-amber-400 to-orange-500 text-white flex items-center justify-center font-bold text-sm">
                {{ getInitial(req.from_user?.username || req.from_user?.email || 'U') }}
              </div>
              <div class="truncate">
                <p class="text-xs font-semibold text-gray-900 truncate">
                  {{ req.from_user?.username || req.from_user?.email || 'User #' + req.from_user_id }}
                </p>
                <p class="text-[11px] text-gray-500">ส่งคำขอเป็นเพื่อนมาให้คุณ</p>
              </div>
            </div>

            <button
              @click="acceptRequest(req.request_id || req.id)"
              class="px-3 py-1.5 bg-sky-500 hover:bg-sky-600 text-white text-xs font-semibold rounded-md transition shadow-sm"
            >
              ยอมรับ
            </button>
          </div>
        </div>
      </div>

      <!-- ADD FRIEND TAB -->
      <div v-else-if="friendsStore.activeTab === 'add'" class="p-4 space-y-4">
        <div>
          <h3 class="text-sm font-semibold text-gray-900 mb-1">เพิ่มเพื่อน</h3>
          <p class="text-xs text-gray-500 mb-3">ระบุชื่อผู้ใช้เพื่อส่งคำขอเป็นเพื่อน</p>

          <form @submit.prevent="handleSendRequest" class="space-y-3">
            <input
              v-model="addFriendInput"
              @input="handleSearchInput"
              type="text"
              required
              placeholder="ชื่อผู้ใช้ หรืออีเมล"
              class="w-full px-3 py-2 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500"
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
          <div v-if="searchResults.length > 0" class="mt-3 border border-gray-200 rounded-md divide-y divide-gray-100 max-h-48 overflow-y-auto bg-white shadow-sm">
            <div
              v-for="u in searchResults"
              :key="u.id"
              @click="sendRequestToUser(u)"
              class="p-2.5 flex items-center justify-between hover:bg-sky-50 cursor-pointer transition"
            >
              <div class="flex items-center space-x-2.5 overflow-hidden">
                <div class="w-8 h-8 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-xs flex-shrink-0">
                  {{ getInitial(u.username || u.email) }}
                </div>
                <div class="truncate">
                  <p class="text-xs font-semibold text-gray-900 truncate">{{ u.username || 'User' }}</p>
                  <p class="text-[11px] text-gray-500 truncate">{{ u.email || '' }}</p>
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
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useFriendsStore } from '../stores/friends'
import { useChatStore } from '../stores/chat'
import Swal from 'sweetalert2'

const router = useRouter()
const authStore = useAuthStore()
const friendsStore = useFriendsStore()
const chatStore = useChatStore()

const searchQuery = ref('')
const addFriendInput = ref('')
const addFriendSuccess = ref('')
const addFriendError = ref('')
const searchResults = ref([])
const isSearching = ref(false)
let searchDebounceTimer = null

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

onMounted(() => {
  friendsStore.fetchFriends()
  friendsStore.fetchPendingRequests()
})
</script>

