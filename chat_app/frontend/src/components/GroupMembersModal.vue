<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm">
    <div class="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 rounded-xl shadow-xl border border-gray-200 dark:border-gray-800 w-full max-w-md overflow-hidden flex flex-col max-h-[85vh]">
      <!-- Modal Header -->
      <div class="px-5 py-4 border-b border-gray-200 dark:border-gray-800 flex items-center justify-between">
        <div>
          <h3 class="font-bold text-base text-gray-900 dark:text-gray-100">
            สมาชิกในกลุ่ม {{ roomName }}
          </h3>
          <p class="text-xs text-gray-500 dark:text-gray-400 mt-0.5">
            สมาชิกทั้งหมด {{ members.length }} คน
          </p>
        </div>
        <button 
          @click="closeModal" 
          class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 p-1.5 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition"
          title="ปิด"
        >
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <!-- Member List -->
      <div class="p-5 flex-1 overflow-y-auto space-y-3">
        <div v-if="loading" class="flex justify-center items-center py-8">
          <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-sky-500"></div>
        </div>

        <div v-else-if="members.length === 0" class="text-center py-8 text-gray-400 text-sm">
          ไม่พบสมาชิกในกลุ่ม
        </div>

        <div v-else class="divide-y divide-gray-100 dark:divide-gray-800">
          <div 
            v-for="member in members" 
            :key="getMemberUserId(member)"
            class="py-2.5 flex items-center justify-between first:pt-0 last:pb-0"
          >
            <!-- Member Details -->
            <div class="flex items-center space-x-3 min-w-0">
              <!-- Avatar -->
              <div class="w-10 h-10 rounded-full bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-sm shadow-sm overflow-hidden flex-shrink-0">
                <img 
                  v-if="member.user?.avatar_url" 
                  :src="getFullUrl(member.user.avatar_url)" 
                  class="w-full h-full object-cover" 
                />
                <span v-else>{{ getInitial(getMemberUsername(member)) }}</span>
              </div>

              <!-- Name, Email, Badges -->
              <div class="min-w-0 flex-1">
                <div class="flex items-center space-x-1.5 flex-wrap">
                  <span class="text-sm font-semibold truncate text-gray-900 dark:text-gray-100">
                    {{ getMemberUsername(member) }}
                  </span>
                  <span v-if="String(getMemberUserId(member)) === String(currentUserId)" class="text-[10px] bg-sky-100 dark:bg-sky-900/40 text-sky-600 dark:text-sky-400 px-1.5 py-0.5 rounded font-medium">
                    คุณ
                  </span>
                  <span 
                    :class="[
                      'text-[10px] px-2 py-0.5 rounded-full font-medium',
                      member.role === 'admin' 
                        ? 'bg-amber-100 dark:bg-amber-900/40 text-amber-600 dark:text-amber-400' 
                        : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'
                    ]"
                  >
                    {{ member.role === 'admin' ? 'แอดมิน' : 'สมาชิก' }}
                  </span>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-400 truncate mt-0.5">
                  {{ getMemberEmail(member) }}
                </p>
              </div>
            </div>

            <!-- Remove Member Button (Admin only & excluding self) -->
            <div v-if="isAdmin && String(getMemberUserId(member)) !== String(currentUserId)" class="ml-2 flex-shrink-0">
              <button 
                @click="handleRemoveMember(member)"
                class="text-xs text-red-500 hover:text-red-700 dark:hover:text-red-400 hover:bg-red-50 dark:hover:bg-red-950/30 px-2.5 py-1 rounded-md border border-red-200 dark:border-red-900/50 transition font-medium"
              >
                ลบออก
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-5 py-3 border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-900/50 flex justify-between items-center">
        <button 
          @click="handleLeaveGroup"
          class="text-xs bg-red-500 hover:bg-red-600 text-white font-semibold px-4 py-2 rounded-lg transition shadow-sm flex items-center space-x-1"
        >
          <span>ออกจากกลุ่ม</span>
        </button>

        <button 
          @click="closeModal"
          class="text-xs bg-gray-200 dark:bg-gray-800 hover:bg-gray-300 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300 font-medium px-4 py-2 rounded-lg transition"
        >
          ปิด
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRoomsStore } from '../stores/rooms'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import Swal from 'sweetalert2'

const props = defineProps({
  isOpen: {
    type: Boolean,
    default: false
  },
  roomId: {
    type: [Number, String],
    default: null
  },
  roomName: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['close'])

const roomsStore = useRoomsStore()
const authStore = useAuthStore()
const chatStore = useChatStore()

const members = ref([])
const loading = ref(false)

const currentUserId = computed(() => authStore.user?.id)

const currentMember = computed(() => {
  if (!currentUserId.value) return null
  return members.value.find(m => String(getMemberUserId(m)) === String(currentUserId.value))
})

const isAdmin = computed(() => {
  return currentMember.value?.role === 'admin'
})

function getMemberUserId(member) {
  if (!member) return null
  return member.user_id || member.user?.id
}

function getMemberUsername(member) {
  if (!member) return 'User'
  return member.user?.username || member.user?.email || `User #${getMemberUserId(member)}`
}

function getMemberEmail(member) {
  if (!member) return ''
  return member.user?.email || ''
}

function getInitial(str) {
  if (!str) return 'U'
  return str.charAt(0).toUpperCase()
}

function getFullUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${import.meta.env.VITE_API_URL || ''}${url}`
}

async function loadMembers() {
  if (!props.roomId) return
  loading.value = true
  try {
    const data = await roomsStore.fetchRoomMembers(props.roomId)
    members.value = Array.isArray(data) ? data : []
  } catch (err) {
    console.error('Failed to load room members:', err)
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.isOpen, props.roomId],
  ([newOpen, newRoomId]) => {
    if (newOpen && newRoomId) {
      loadMembers()
    }
  },
  { immediate: true }
)

function closeModal() {
  emit('close')
}

async function handleLeaveGroup() {
  const result = await Swal.fire({
    title: 'ยืนยันการออกจากกลุ่ม',
    text: `คุณต้องการออกจากกลุ่ม "${props.roomName}" ใช่หรือไม่?`,
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#d33',
    cancelButtonColor: '#3085d6',
    confirmButtonText: 'ออกจากกลุ่ม',
    cancelButtonText: 'ยกเลิก'
  })

  if (result.isConfirmed) {
    try {
      await roomsStore.leaveRoom(props.roomId)
      if (Number(chatStore.selectedRoomId) === Number(props.roomId)) {
        chatStore.selectedRoomId = null
      }
      closeModal()
      Swal.fire({
        icon: 'success',
        title: 'ออกจากกลุ่มเรียบร้อยแล้ว',
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000
      })
    } catch (err) {
      Swal.fire('เกิดข้อผิดพลาด', err.message || 'ไม่สามารถออกจากกลุ่มได้', 'error')
    }
  }
}

async function handleRemoveMember(member) {
  const userId = getMemberUserId(member)
  const username = getMemberUsername(member)

  const result = await Swal.fire({
    title: 'ยืนยันการลบสมาชิก',
    text: `คุณต้องการลบ "${username}" ออกจากกลุ่มใช่หรือไม่?`,
    icon: 'warning',
    showCancelButton: true,
    confirmButtonColor: '#d33',
    cancelButtonColor: '#3085d6',
    confirmButtonText: 'ลบออก',
    cancelButtonText: 'ยกเลิก'
  })

  if (result.isConfirmed) {
    try {
      await roomsStore.removeRoomMember(props.roomId, userId)
      await loadMembers()
      Swal.fire({
        icon: 'success',
        title: `ลบ ${username} ออกจากกลุ่มแล้ว`,
        toast: true,
        position: 'top-end',
        showConfirmButton: false,
        timer: 3000
      })
    } catch (err) {
      Swal.fire('เกิดข้อผิดพลาด', err.message || 'ไม่สามารถลบสมาชิกได้', 'error')
    }
  }
}
</script>
