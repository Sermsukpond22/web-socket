<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
    <div class="max-w-sm w-full bg-white border border-gray-300 rounded-lg p-8 shadow-sm">
      <!-- Logo Header -->
      <div class="flex flex-col items-center mb-8">
        <div class="w-12 h-12 bg-gradient-to-tr from-yellow-400 via-pink-500 to-purple-600 rounded-2xl flex items-center justify-center text-white mb-3 shadow-md">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
          </svg>
        </div>
        <h1 class="font-serif text-3xl text-gray-900 tracking-tight">เข้าสู่ระบบ</h1>
        <p class="text-sm text-gray-500 mt-1">แชทกับเพื่อนของคุณได้ทันที</p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-md border border-red-200">
        {{ errorMessage }}
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label for="email" class="sr-only">Email address</label>
          <input
            id="email"
            v-model="email"
            type="email"
            required
            placeholder="อีเมล"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400"
          />
        </div>

        <div>
          <label for="password" class="sr-only">Password</label>
          <input
            id="password"
            v-model="password"
            type="password"
            required
            placeholder="รหัสผ่าน"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400"
          />
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full bg-sky-500 hover:bg-sky-600 active:bg-sky-700 disabled:opacity-50 text-white text-sm font-semibold py-2.5 rounded-md transition duration-150 ease-in-out shadow-sm flex justify-center items-center"
        >
          <span v-if="!authStore.loading">เข้าสู่ระบบ</span>
          <svg v-else class="animate-spin h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </button>
      </form>
    </div>

    <!-- Toggle to Register -->
    <div class="max-w-sm w-full bg-white border border-gray-300 rounded-lg p-4 mt-3 text-center shadow-sm">
      <p class="text-sm text-gray-600">
        ยังไม่มีบัญชีผู้ใช้?
        <router-link to="/register" class="text-sky-500 font-semibold hover:text-sky-600 ml-1">
          สมัครสมาชิก
        </router-link>
      </p>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Swal from 'sweetalert2'

const router = useRouter()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const errorMessage = ref('')

async function handleLogin() {
  errorMessage.value = ''
  try {
    await authStore.login(email.value, password.value)
    
    await Swal.fire({
      icon: 'success',
      title: 'เข้าสู่ระบบสำเร็จ!',
      text: 'ยินดีต้อนรับกลับมา',
      timer: 1500,
      showConfirmButton: false
    })

    router.push('/')
  } catch (err) {
    errorMessage.value = err.message || 'เข้าสู่ระบบไม่สำเร็จ โปรดตรวจสอบอีเมลหรือรหัสผ่าน'
  }
}
</script>
