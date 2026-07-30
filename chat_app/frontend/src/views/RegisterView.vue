<template>
  <div class="min-h-screen flex flex-col items-center justify-center bg-gray-50 px-4">
    <div class="max-w-sm w-full bg-white border border-gray-300 rounded-lg p-8 shadow-sm">
      <!-- Logo Header -->
      <div class="flex flex-col items-center mb-6">
        <div class="w-12 h-12 bg-gradient-to-tr from-yellow-400 via-pink-500 to-purple-600 rounded-2xl flex items-center justify-center text-white mb-3 shadow-md">
          <svg class="w-7 h-7" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z" />
          </svg>
        </div>
        <h1 class="font-serif text-3xl text-gray-900 tracking-tight">สร้างบัญชีผู้ใช้</h1>
        <p class="text-xs text-gray-500 mt-1 text-center">สมัครสมาชิกเพื่อแชทกับเพื่อนของคุณ</p>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMessage" class="mb-4 p-3 bg-red-50 text-red-600 text-sm rounded-md border border-red-200">
        {{ errorMessage }}
      </div>

      <!-- Register Form -->
      <form @submit.prevent="handleRegister" class="space-y-3.5">
        <div>
          <label for="username" class="sr-only">Username</label>
          <input
            id="username"
            v-model="username"
            type="text"
            required
            placeholder="ชื่อผู้ใช้ (Username)"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400"
          />
        </div>

        <div>
          <label for="displayName" class="sr-only">Display Name</label>
          <input
            id="displayName"
            v-model="displayName"
            type="text"
            required
            placeholder="ชื่อแสดงผล (Display Name)"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400"
          />
        </div>

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
            placeholder="รหัสผ่าน (อย่างน้อย 6 ตัวอักษร)"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400"
          />
        </div>

        <div>
          <label for="bio" class="sr-only">Bio (Optional)</label>
          <textarea
            id="bio"
            v-model="bio"
            placeholder="คำแนะนำตัวสั้นๆ (ไม่บังคับ)"
            rows="2"
            class="w-full px-3 py-2.5 bg-gray-50 border border-gray-300 rounded-md text-sm text-gray-900 focus:outline-none focus:ring-1 focus:ring-sky-500 focus:border-sky-500 placeholder-gray-400 resize-none"
          ></textarea>
        </div>

        <button
          type="submit"
          :disabled="authStore.loading"
          class="w-full bg-sky-500 hover:bg-sky-600 active:bg-sky-700 disabled:opacity-50 text-white text-sm font-semibold py-2.5 rounded-md transition duration-150 ease-in-out shadow-sm flex justify-center items-center"
        >
          <span v-if="!authStore.loading">สมัครสมาชิก</span>
          <svg v-else class="animate-spin h-5 w-5 text-white" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
        </button>
      </form>
    </div>

    <!-- Toggle to Login -->
    <div class="max-w-sm w-full bg-white border border-gray-300 rounded-lg p-4 mt-3 text-center shadow-sm">
      <p class="text-sm text-gray-600">
        มีบัญชีอยู่แล้ว?
        <router-link to="/login" class="text-sky-500 font-semibold hover:text-sky-600 ml-1">
          เข้าสู่ระบบ
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

const username = ref('')
const email = ref('')
const password = ref('')
const displayName = ref('')
const bio = ref('')
const errorMessage = ref('')

async function handleRegister() {
  errorMessage.value = ''
  
  // Frontend Validation
  if (!username.value.trim() || !email.value.trim() || !password.value.trim() || !displayName.value.trim()) {
    errorMessage.value = 'กรุณากรอกข้อมูลที่จำเป็นให้ครบถ้วน'
    return
  }
  if (password.value.length < 6) {
    errorMessage.value = 'รหัสผ่านต้องมีความยาวอย่างน้อย 6 ตัวอักษร'
    return
  }

  try {
    await authStore.register(username.value, email.value, password.value, displayName.value, bio.value)
    
    await Swal.fire({
      icon: 'success',
      title: 'สมัครสมาชิกสำเร็จ!',
      text: 'กรุณาเข้าสู่ระบบด้วยบัญชีใหม่ของคุณ',
      confirmButtonText: 'ตกลง',
      confirmButtonColor: '#0ea5e9'
    })
    
    router.push('/login')
  } catch (err) {
    errorMessage.value = err.message || 'สมัครสมาชิกไม่สำเร็จ โปรดลองอีกครั้ง'
  }
}
</script>
