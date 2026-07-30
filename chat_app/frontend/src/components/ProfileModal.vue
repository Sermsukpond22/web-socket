<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50 transition-opacity">
    <div class="bg-white dark:bg-gray-800 rounded-xl shadow-xl w-full max-w-md overflow-hidden transform transition-all p-6 relative">
      <!-- Close Button -->
      <button 
        @click="close" 
        class="absolute top-4 right-4 text-gray-400 hover:text-gray-600 dark:hover:text-gray-200 transition"
      >
        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>

      <div class="text-center">
        <!-- Avatar Section -->
        <div class="relative inline-block mt-4 mb-6 group">
          <div 
            class="w-32 h-32 rounded-full mx-auto border-4 border-gray-100 dark:border-gray-700 bg-gradient-to-tr from-sky-400 to-blue-600 text-white flex items-center justify-center font-bold text-4xl shadow-md overflow-hidden"
          >
            <img v-if="previewUrl || user.avatar_url" :src="previewUrl || getFullUrl(user.avatar_url)" alt="Avatar" class="w-full h-full object-cover" />
            <span v-else>{{ getInitial(user.display_name || user.username || user.email) }}</span>
          </div>

          <!-- Edit Avatar Overlay -->
          <label v-if="isOwnProfile && isEditing" class="absolute inset-0 bg-black bg-opacity-40 rounded-full flex flex-col items-center justify-center cursor-pointer opacity-0 group-hover:opacity-100 transition-opacity text-white text-xs">
            <svg class="w-6 h-6 mb-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 9a2 2 0 012-2h.93a2 2 0 001.664-.89l.812-1.22A2 2 0 0110.07 4h3.86a2 2 0 011.664.89l.812 1.22A2 2 0 0018.07 7H19a2 2 0 012 2v9a2 2 0 01-2 2H5a2 2 0 01-2-2V9z" />
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 13a3 3 0 11-6 0 3 3 0 016 0z" />
            </svg>
            เปลี่ยนรูป
            <input type="file" class="hidden" accept="image/*" @change="handleFileChange" />
          </label>
        </div>

        <!-- View Mode -->
        <div v-if="!isEditing">
          <h2 class="text-2xl font-bold text-gray-900 dark:text-white">{{ user.display_name || user.username || 'User' }}</h2>
          <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">{{ user.email }}</p>
          
          <div class="mt-6 text-left bg-gray-50 dark:bg-gray-700/50 p-4 rounded-lg">
            <h3 class="text-xs font-semibold text-gray-500 dark:text-gray-400 uppercase tracking-wider mb-2">เกี่ยวกับ</h3>
            <p class="text-sm text-gray-800 dark:text-gray-200 whitespace-pre-wrap">{{ user.bio || 'ยังไม่มีคำอธิบายตัวตน...' }}</p>
          </div>

          <div class="mt-8" v-if="isOwnProfile">
            <button @click="startEditing" class="w-full py-2.5 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-sky-500 transition">
              แก้ไขโปรไฟล์
            </button>
          </div>
        </div>

        <!-- Edit Mode -->
        <form v-else @submit.prevent="saveProfile" class="text-left space-y-4">
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">ชื่อที่แสดง (Display Name)</label>
            <input 
              v-model="editForm.display_name" 
              type="text" 
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:text-white"
              placeholder="ชื่อที่ต้องการให้เพื่อนเห็น"
            />
          </div>

          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300">เกี่ยวกับ (Bio)</label>
            <textarea 
              v-model="editForm.bio" 
              rows="3" 
              class="mt-1 block w-full px-3 py-2 bg-white dark:bg-gray-700 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-sky-500 focus:border-sky-500 sm:text-sm dark:text-white"
              placeholder="อธิบายความเป็นตัวคุณ..."
            ></textarea>
          </div>

          <!-- Error/Success Messages -->
          <div v-if="error" class="text-sm text-red-600 dark:text-red-400">{{ error }}</div>
          <div v-if="success" class="text-sm text-green-600 dark:text-green-400">บันทึกสำเร็จ!</div>

          <div class="flex space-x-3 pt-4">
            <button type="button" @click="cancelEditing" class="flex-1 py-2 px-4 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-gray-800 hover:bg-gray-50 dark:hover:bg-gray-700 transition">
              ยกเลิก
            </button>
            <button type="submit" :disabled="loading" class="flex-1 py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-sky-600 hover:bg-sky-700 disabled:opacity-50 transition">
              {{ loading ? 'กำลังบันทึก...' : 'บันทึก' }}
            </button>
          </div>
        </form>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'

const props = defineProps({
  isOpen: Boolean,
  user: {
    type: Object,
    required: true
  },
  isOwnProfile: Boolean
})

const emit = defineEmits(['close', 'updated'])
const authStore = useAuthStore()

const isEditing = ref(false)
const editForm = ref({ display_name: '', bio: '' })
const selectedFile = ref(null)
const previewUrl = ref(null)
const loading = ref(false)
const error = ref('')
const success = ref(false)

watch(() => props.isOpen, (val) => {
  if (val) {
    isEditing.value = false
    selectedFile.value = null
    previewUrl.value = null
    error.value = ''
    success.value = false
  }
})

function close() {
  emit('close')
}

function getInitial(str) {
  if (!str) return 'U'
  return str.charAt(0).toUpperCase()
}

function getFullUrl(url) {
  if (!url) return ''
  if (url.startsWith('http')) return url
  return `${import.meta.env.VITE_API_URL}${url}`
}

function startEditing() {
  editForm.value = {
    display_name: props.user.display_name || '',
    bio: props.user.bio || ''
  }
  isEditing.value = true
  success.value = false
  error.value = ''
}

function cancelEditing() {
  isEditing.value = false
  selectedFile.value = null
  previewUrl.value = null
}

function handleFileChange(event) {
  const file = event.target.files[0]
  if (file) {
    if (file.size > 5 * 1024 * 1024) {
      error.value = 'ไฟล์รูปภาพต้องมีขนาดไม่เกิน 5MB'
      return
    }
    selectedFile.value = file
    previewUrl.value = URL.createObjectURL(file)
  }
}

async function saveProfile() {
  loading.value = true
  error.value = ''
  success.value = false

  try {
    // 1. Update text info
    const res = await fetch(`${import.meta.env.VITE_API_URL}/api/users/profile`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${authStore.token}`
      },
      body: JSON.stringify(editForm.value)
    })
    
    if (!res.ok) throw new Error('Failed to update profile text')

    // 2. Upload avatar if selected
    if (selectedFile.value) {
      const formData = new FormData()
      formData.append('avatar', selectedFile.value)

      const avatarRes = await fetch(`${import.meta.env.VITE_API_URL}/api/users/avatar`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${authStore.token}`
        },
        body: formData
      })

      if (!avatarRes.ok) throw new Error('Failed to upload avatar')
    }

    success.value = true
    setTimeout(() => {
      isEditing.value = false
      emit('updated')
      // Ensure authStore is updated so the main UI reacts
      authStore.fetchMe()
    }, 1000)

  } catch (err) {
    error.value = err.message
  } finally {
    loading.value = false
  }
}
</script>
