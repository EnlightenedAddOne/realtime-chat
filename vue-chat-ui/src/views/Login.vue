<template>
  <div class="login-container">
    <!-- Sophisticated animated background -->
    <div class="mesh-loader">
      <div class="mesh-orb orb-1"></div>
      <div class="mesh-orb orb-2"></div>
      <div class="mesh-orb orb-3"></div>
    </div>
    
    <div class="glass-card">
      <!-- Overlay Box (Sliding Panel) - Desktop Only -->
      <div class="overlay-panel" :class="{ 'slide-active': !isLogin }">
        <div class="overlay-content">
          <div class="avatar-ring">
            <div class="avatar-inner">
              <img :src="currentImage" alt="Avatar" />
            </div>
          </div>
          <h1 class="welcome-heading">{{ isLogin ? '你好，朋友' : '加入社区' }}</h1>
          <p class="welcome-sub">{{ isLogin ? '欢迎回来！开始聊天吧~' : '创建账号，连接精彩世界' }}</p>
        </div>
        <!-- Enhanced decorative elements -->
        <div class="deco-circle c1"></div>
        <div class="deco-circle c2"></div>
        <div class="deco-circle c3"></div>
        <div class="deco-ring r1"></div>
        <div class="deco-ring r2"></div>
        <div class="deco-dots"></div>
      </div>

      <!-- Mobile Tab Switcher -->
      <div class="mobile-tabs">
        <div 
          class="mobile-tab" 
          :class="{ active: isLogin }" 
          @click="isLogin = true"
        >
          <span>登录</span>
          <div class="tab-indicator" v-if="isLogin"></div>
        </div>
        <div 
          class="mobile-tab" 
          :class="{ active: !isLogin }" 
          @click="isLogin = false"
        >
          <span>注册</span>
          <div class="tab-indicator" v-if="!isLogin"></div>
        </div>
      </div>

      <!-- Register Form -->
      <div class="form-container register-container" :class="{ 'form-active': !isLogin }">
        <div class="form-wrapper">
          <div class="form-header">
            <h2>创建账号</h2>
            <p>请填写以下信息完成注册</p>
          </div>
          
          <el-form :model="registerForm" :rules="registerRules" ref="registerFormRef" class="modern-form" label-width="0">
            <!-- Email input -->
            <el-form-item prop="email">
              <el-input 
                v-model="registerForm.email" 
                placeholder="邮箱地址" 
                :prefix-icon="Message"
                class="premium-input"
              >
                <template #append>
                  <el-button 
                    @click="sendVerifyCode" 
                    :disabled="codeCountdown > 0"
                    class="code-btn"
                  >
                    {{ codeCountdown > 0 ? `${codeCountdown}s` : '获取验证码' }}
                  </el-button>
                </template>
              </el-input>
            </el-form-item>
            
            <!-- Verification code -->
            <el-form-item prop="code">
              <el-input 
                v-model="registerForm.code" 
                placeholder="邮箱验证码" 
                :prefix-icon="CircleCheck"
                maxlength="6"
                class="premium-input"
              />
            </el-form-item>
            
            <!-- Password -->
            <el-form-item prop="password">
              <el-input 
                v-model="registerForm.password" 
                type="password" 
                placeholder="密码 (6-20位，仅英文数字)" 
                :prefix-icon="Lock" 
                show-password
                class="premium-input"
              />
            </el-form-item>
            
            <!-- Password strength -->
            <div class="pwd-strength" v-if="registerForm.password">
              <div class="strength-track">
                <div class="strength-fill" 
                     :style="{ width: passwordStrength.percent + '%', background: getStrengthColor(passwordStrength.level) }">
                </div>
              </div>
              <span class="strength-label" :style="{ color: getStrengthColor(passwordStrength.level) }">
                {{ passwordStrength.text }}
              </span>
            </div>
            
            <!-- Confirm Password -->
            <el-form-item prop="confirmPassword">
              <el-input 
                v-model="registerForm.confirmPassword" 
                type="password" 
                placeholder="确认密码" 
                :prefix-icon="Lock" 
                show-password
                class="premium-input"
              />
            </el-form-item>
          </el-form>

          <div class="action-area">
            <button class="primary-btn register-btn" @click="handleRegister" :disabled="loading">
              <span class="btn-text">{{ loading ? '注册中...' : '立即注册' }}</span>
              <span class="btn-icon" v-if="!loading">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M5 12h14M12 5l7 7-7 7"/>
                </svg>
              </span>
            </button>
            <div class="switch-mode desktop-only" @click="toggleMode">
              已有账号？ <span>去登录</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Login Form -->
      <div class="form-container login-container-inner" :class="{ 'form-active': isLogin }">
        <div class="form-wrapper">
          <div class="form-header">
            <h2>欢迎回来</h2>
            <p>请输入您的账号信息</p>
          </div>
          
          <el-form :model="loginForm" :rules="loginRules" ref="loginFormRef" class="modern-form" label-width="0">
            <el-form-item prop="email">
              <el-input 
                v-model="loginForm.email" 
                placeholder="邮箱地址" 
                :prefix-icon="Message"
                class="premium-input"
                @keyup.enter="handleLogin"
              />
            </el-form-item>
            <el-form-item prop="password">
              <el-input 
                v-model="loginForm.password" 
                type="password" 
                placeholder="密码" 
                :prefix-icon="Lock" 
                show-password 
                class="premium-input"
                @keyup.enter="handleLogin"
              />
            </el-form-item>
          </el-form>

          <div class="action-area">
            <button class="primary-btn login-btn" @click="handleLogin" :disabled="loading">
              <span class="btn-text">{{ loading ? '登录中...' : '立即登录' }}</span>
              <span class="btn-icon" v-if="!loading">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
                  <path d="M15 3h4a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2h-4M10 17l5-5-5-5M15 12H3"/>
                </svg>
              </span>
            </button>
            <div class="switch-mode desktop-only" @click="toggleMode">
              还没有账号？ <span>去注册</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Message, CircleCheck } from '@element-plus/icons-vue'
import { authAPI } from '../api'
import { useUserStore } from '../store/user'
import { connectWebSocket } from '../utils/ws'
import loginAvatar from '../assets/login-avatar.png'
import registerAvatar from '../assets/register-avatar.png'

const router = useRouter()
const userStore = useUserStore()

const imgLogin = loginAvatar
const imgRegister = registerAvatar
const currentImage = computed(() => isLogin.value ? imgLogin : imgRegister)

const isLogin = ref(true)
const loading = ref(false)
const codeCountdown = ref(0)

const loginFormRef = ref(null)
const registerFormRef = ref(null)

const loginForm = reactive({
  email: '',
  password: ''
})

const registerForm = reactive({
  email: '',
  code: '',
  password: '',
  confirmPassword: ''
})

// Password strength calculation
const passwordStrength = computed(() => {
  const pwd = registerForm.password
  if (!pwd) return { level: '', percent: 0, text: '' }
  
  const isAlnum = /^[A-Za-z0-9]+$/.test(pwd)
  if (!isAlnum) return { level: 'weak', percent: 35, text: '仅支持英文数字' }
  if (pwd.length < 6) return { level: 'weak', percent: 35, text: '至少6位' }
  if (pwd.length < 10) return { level: 'medium', percent: 70, text: '可用' }
  return { level: 'strong', percent: 100, text: '良好' }
})

// Helper for strength color in new design
const getStrengthColor = (level) => {
  if (level === 'weak') return '#ff4757';
  if (level === 'medium') return '#ffa502';
  if (level === 'strong') return '#2ed573';
  return '#e1e1e1';
}

const validateEmail = (rule, value, callback) => {
  const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/
  if (!value) {
    callback(new Error('请输入邮箱'))
  } else if (!emailRegex.test(value)) {
    callback(new Error('请输入正确的邮箱格式'))
  } else {
    callback()
  }
}

const validateCode = (rule, value, callback) => {
  if (!value) {
    callback(new Error('请输入验证码'))
  } else if (!/^\d{6}$/.test(value)) {
    callback(new Error('验证码为6位数字'))
  } else {
    callback()
  }
}

const validatePassword = (rule, value, callback) => {
  const pwdRegex = /^[A-Za-z0-9]{6,20}$/
  if (!value) {
    callback(new Error('请输入密码'))
  } else if (!pwdRegex.test(value)) {
    callback(new Error('密码需为6-20位，仅支持英文和数字'))
  } else {
    callback()
  }
}

const registerRules = {
  password: [
    { required: true, validator: validatePassword, trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    {
      validator: (rule, value, callback) => {
        if (value !== registerForm.password) {
          callback(new Error('两次输入密码不一致'))
        } else {
          callback()
        }
      },
      trigger: 'blur'
    }
  ],
  email: [
    { required: true, validator: validateEmail, trigger: 'blur' }
  ],
  code: [
    { required: true, validator: validateCode, trigger: 'blur' }
  ]
}

const loginRules = {
  email: [
    { required: true, validator: validateEmail, trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

function toggleMode() {
  isLogin.value = !isLogin.value
  // Reset forms
  if (loginFormRef.value) loginFormRef.value.resetFields()
  if (registerFormRef.value) registerFormRef.value.resetFields()
  // Reset code countdown
  codeCountdown.value = 0
}

// Send verification code
async function sendVerifyCode() {
  if (!registerForm.email) {
    ElMessage.warning('请先输入邮箱')
    return
  }
  
  const emailRegex = /^[\w.-]+@[\w.-]+\.\w+$/
  if (!emailRegex.test(registerForm.email)) {
    ElMessage.warning('请输入正确的邮箱格式')
    return
  }
  
  try {
    loading.value = true
    await authAPI.sendCode(registerForm.email, 'register')
    ElMessage.success('验证码已发送到您的邮箱')
    
    // Start countdown
    codeCountdown.value = 60
    const timer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0) {
        clearInterval(timer)
      }
    }, 1000)
  } catch (e) {
    // Error already handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleLogin() {
  if (!loginFormRef.value) return

  const valid = await loginFormRef.value.validate().catch(() => false)
  if (!valid) {
    ElMessage.warning('请先填写完整的登录信息')
    return
  }

  loading.value = true
  try {
    const email = loginForm.email.trim()
    const password = loginForm.password
    const res = await authAPI.login(email, password)
    const token = res.token || res.data?.token || (res.data && res.data.token)
    const user = res.user || res.data?.user || (res.data && res.data.user)

    if (token) {
      userStore.setToken(token)
      userStore.setUserInfo(user)
      ElMessage.success('登录成功')
      connectWebSocket()
      router.push('/chat')
    } else {
      throw new Error('Invalid response')
    }
  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.message || '登录失败')
  } finally {
    loading.value = false
  }
}

async function handleRegister() {
  if (!registerFormRef.value) return

  const valid = await registerFormRef.value.validate().then(() => true).catch((fields) => {
    const firstError = fields && Object.values(fields)[0]?.[0]?.message
    ElMessage.warning(firstError || '请先完善并修正注册信息')
    return false
  })
  if (!valid) {
    return
  }

  loading.value = true
  try {
    await authAPI.register({
      email: registerForm.email.trim(),
      code: registerForm.code.trim(),
      password: registerForm.password
    })
    ElMessage.success('注册成功，请登录')
    toggleMode() // Switch to login view
  } catch (e) {
    console.error(e)
    ElMessage.error(e.response?.data?.message || '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* =========================================
   Base & Background (Mesh Gradient)
   ========================================= */
.login-container {
  position: relative;
  min-height: 100vh;
  width: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  overflow: hidden;
  background-color: #f3f4f6;
  font-family: 'Inter', -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
}

.mesh-loader {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  filter: blur(80px);
  overflow: hidden;
}

.mesh-orb {
  position: absolute;
  border-radius: 50%;
  opacity: 0.6;
  animation: floatOrb 20s infinite ease-in-out;
}

.orb-1 {
  width: 60vw;
  height: 60vw;
  background: radial-gradient(circle, #667eea 0%, #764ba2 100%);
  top: -20%;
  left: -20%;
  animation-duration: 25s;
}

.orb-2 {
  width: 50vw;
  height: 50vw;
  background: radial-gradient(circle, #f093fb 0%, #f5576c 100%);
  bottom: -10%;
  right: -10%;
  animation-duration: 30s;
  animation-direction: reverse;
}

.orb-3 {
  width: 40vw;
  height: 40vw;
  background: radial-gradient(circle, #4facfe 0%, #00f2fe 100%);
  top: 40%;
  left: 40%;
  animation-duration: 35s;
}

@keyframes floatOrb {
  0% { transform: translate(0, 0) rotate(0deg); }
  33% { transform: translate(50px, 80px) rotate(10deg); }
  66% { transform: translate(-30px, 20px) rotate(-5deg); }
  100% { transform: translate(0, 0) rotate(0deg); }
}

/* =========================================
   Glass Card
   ========================================= */
.glass-card {
  position: relative;
  width: 1100px;
  min-height: 640px;
  background: rgba(255, 255, 255, 0.65);
  backdrop-filter: blur(24px);
  -webkit-backdrop-filter: blur(24px);
  border: 1px solid rgba(255, 255, 255, 0.4);
  border-radius: 24px;
  box-shadow: 
    0 25px 50px -12px rgba(0, 0, 0, 0.1),
    0 0 0 1px rgba(255, 255, 255, 0.2) inset;
  display: flex;
  overflow: hidden;
  z-index: 10;
  transition: all 0.5s ease;
}

/* =========================================
   Overlay Panel (The "Curtain")
   ========================================= */
.overlay-panel {
  position: absolute;
  top: 0;
  left: 0;
  width: 50%;
  height: 100%;
  z-index: 100;
  background: linear-gradient(145deg, #667eea 0%, #764ba2 50%, #f093fb 100%);
  backdrop-filter: blur(30px);
  -webkit-backdrop-filter: blur(30px);
  color: #fff;
  transition: transform 0.7s cubic-bezier(0.8, 0, 0.2, 1);
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  text-align: center;
  overflow: hidden;
}

.overlay-panel.slide-active {
  transform: translateX(100%);
  background: linear-gradient(145deg, #f093fb 0%, #f5576c 50%, #ff9a9e 100%);
}

.overlay-content {
  position: relative;
  z-index: 5;
  padding: 0 50px;
  max-width: 400px;
}

/* Avatar Ring */
.avatar-ring {
  width: 150px;
  height: 150px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.2);
  border: 3px solid rgba(255, 255, 255, 0.4);
  margin: 0 auto 30px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(5px);
  animation: pulse-ring 3s ease-in-out infinite;
}

.avatar-inner {
  width: 110px;
  height: 110px;
  border-radius: 50%;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.15);
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-ring img {
  width: 90%;
  height: 90%;
  object-fit: contain;
  transition: transform 0.5s ease;
}

@keyframes pulse-ring {
  0%, 100% { box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2), 0 0 0 0 rgba(255, 255, 255, 0.3); }
  50% { box-shadow: 0 15px 35px rgba(0, 0, 0, 0.2), 0 0 0 15px rgba(255, 255, 255, 0); }
}

.welcome-heading {
  font-size: 2.6rem;
  font-weight: 800;
  margin-bottom: 15px;
  letter-spacing: 2px;
  text-shadow: 0 4px 10px rgba(0, 0, 0, 0.15);
  animation: fadeSlideUp 0.6s ease forwards;
}

.welcome-sub {
  font-size: 1.05rem;
  opacity: 0.9;
  line-height: 1.6;
  animation: fadeSlideUp 0.6s ease 0.1s forwards;
  opacity: 0;
}

@keyframes fadeSlideUp {
  from { opacity: 0; transform: translateY(15px); }
  to { opacity: 1; transform: translateY(0); }
}

/* =========================================
   Enhanced Decorative Elements
   ========================================= */
.deco-circle {
  position: absolute;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.12);
  animation: float 8s ease-in-out infinite;
}

.deco-circle.c1 {
  width: 280px;
  height: 280px;
  top: -80px;
  right: -60px;
  animation-duration: 10s;
}

.deco-circle.c2 {
  width: 180px;
  height: 180px;
  bottom: -40px;
  left: -40px;
  animation-duration: 12s;
  animation-delay: 2s;
}

.deco-circle.c3 {
  width: 100px;
  height: 100px;
  bottom: 30%;
  right: 10%;
  animation-duration: 8s;
  animation-delay: 4s;
  background: rgba(255, 255, 255, 0.08);
}

.deco-ring {
  position: absolute;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.15);
  animation: rotate 20s linear infinite;
}

.deco-ring.r1 {
  width: 350px;
  height: 350px;
  top: 50%;
  left: -100px;
  transform: translateY(-50%);
}

.deco-ring.r2 {
  width: 200px;
  height: 200px;
  bottom: 10%;
  right: -50px;
  animation-direction: reverse;
  animation-duration: 15s;
}

.deco-dots {
  position: absolute;
  top: 20%;
  right: 15%;
  width: 80px;
  height: 80px;
  background-image: radial-gradient(rgba(255,255,255,0.3) 2px, transparent 2px);
  background-size: 12px 12px;
  animation: float 6s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0) rotate(0deg); }
  50% { transform: translateY(-20px) rotate(5deg); }
}

@keyframes rotate {
  from { transform: translateY(-50%) rotate(0deg); }
  to { transform: translateY(-50%) rotate(360deg); }
}

/* =========================================
   Forms Area
   ========================================= */
.form-container {
  position: absolute;
  top: 0;
  height: 100%;
  width: 50%;
  transition: all 0.7s cubic-bezier(0.8, 0, 0.2, 1);
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  z-index: 1;
  visibility: hidden;
}

.login-container-inner {
  right: 0;
  z-index: 2;
}

.register-container {
  left: 0;
  z-index: 1;
}

.form-active {
  opacity: 1;
  z-index: 5;
  visibility: visible;
  animation: fadeInUp 0.7s ease 0.2s both;
}

@keyframes fadeInUp {
  from { opacity: 0; transform: translateY(20px); }
  to { opacity: 1; transform: translateY(0); }
}

.form-wrapper {
  width: 100%;
  max-width: 380px;
  padding: 0 20px;
}

.form-header {
  text-align: center;
  margin-bottom: 35px;
}

.form-header h2 {
  font-size: 2rem;
  color: #1f2937;
  font-weight: 700;
  margin-bottom: 10px;
}

.form-header p {
  color: #9ca3af;
  font-size: 0.95rem;
}

/* Input Customization */
.modern-form .el-form-item {
  margin-bottom: 24px;
}

:deep(.premium-input .el-input__wrapper) {
  background: #f3f5f7;
  box-shadow: none !important;
  border-radius: 12px;
  padding: 1px 15px;
  transition: all 0.3s ease;
  height: 48px;
}

:deep(.premium-input .el-input__wrapper:hover) {
  background: #eef1f5;
}

:deep(.premium-input .el-input__wrapper.is-focus) {
  background: #fff;
  box-shadow: 0 0 0 4px rgba(245, 87, 108, 0.15) !important;
  border-color: #f5576c;
}

:deep(.el-input__inner) {
  height: 48px;
  font-size: 0.95rem;
  font-weight: 500;
  color: #374151;
}

:deep(.el-input__prefix-inner) {
  font-size: 1.1rem;
  color: #9ca3af;
}

/* Code Button */
.code-btn {
  border: none !important;
  background: transparent !important;
  color: #4facfe !important;
  font-weight: 600;
  padding: 0 10px !important;
  height: auto !important;
}

.code-btn:disabled {
  color: #9ca3af !important;
}

/* Password Strength */
.pwd-strength {
  margin: -10px 0 20px;
  display: flex;
  align-items: center;
  gap: 10px;
}

.strength-track {
  flex: 1;
  height: 4px;
  background: #e5e7eb;
  border-radius: 2px;
  overflow: hidden;
}

.strength-fill {
  height: 100%;
  border-radius: 2px;
  transition: width 0.3s ease, background 0.3s ease;
}

.strength-label {
  font-size: 0.75rem;
  font-weight: 600;
  min-width: 30px;
  text-align: right;
}

/* Action Area & Buttons */
.action-area {
  margin-top: 10px;
  text-align: center;
}

.primary-btn {
  width: 100%;
  height: 54px;
  border: none;
  border-radius: 27px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  margin-bottom: 20px;
  position: relative;
  overflow: hidden;
}

.primary-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -100%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.2), transparent);
  transition: left 0.5s;
}

.primary-btn:hover::before {
  left: 100%;
}

.login-btn {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  box-shadow: 0 8px 25px -5px rgba(245, 87, 108, 0.45);
  color: white;
  font-size: 1.05rem;
  font-weight: 600;
  letter-spacing: 1px;
}

.login-btn:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 35px -5px rgba(245, 87, 108, 0.5);
}

.register-btn {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  box-shadow: 0 8px 25px -5px rgba(245, 87, 108, 0.45);
  color: white;
  font-size: 1.05rem;
  font-weight: 600;
  letter-spacing: 1px;
}

.register-btn:hover {
  transform: translateY(-3px);
  box-shadow: 0 12px 35px -5px rgba(245, 87, 108, 0.5);
}

.btn-text {
  position: relative;
  z-index: 1;
}

.btn-icon {
  display: flex;
  align-items: center;
  position: relative;
  z-index: 1;
  transition: transform 0.3s ease;
}

.primary-btn:hover .btn-icon {
  transform: translateX(4px);
}

.primary-btn:active {
  transform: scale(0.98);
}

.primary-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
  transform: none;
}

.switch-mode {
  color: #6b7280;
  font-size: 0.9rem;
  cursor: pointer;
  transition: color 0.2s;
}

.switch-mode span {
  color: #f5576c;
  font-weight: 600;
  margin-left: 5px;
}

.register-container .switch-mode span {
  color: #667eea;
}

.switch-mode:hover span {
  text-decoration: underline;
}

/* =========================================
   Mobile Tabs
   ========================================= */
.mobile-tabs {
  display: none;
}

/* =========================================
   Responsive Design
   ========================================= */
@media (max-width: 1024px) {
  .login-container {
    align-items: flex-start;
    padding-top: 60px;
    background: #fff;
  }

  .glass-card {
    width: 100%;
    min-height: 100vh;
    border-radius: 0;
    border: none;
    box-shadow: none;
    background: transparent;
    flex-direction: column;
    overflow: visible;
  }

  .overlay-panel {
    display: none;
  }

  .form-container {
    position: relative;
    width: 100%;
    height: auto;
    opacity: 1;
    visibility: visible;
    transform: none;
    display: none;
    padding-top: 20px;
  }

  .form-active {
    display: flex;
    animation: slideUpMobile 0.4s ease forwards;
  }
  
  @keyframes slideUpMobile {
    from { opacity: 0; transform: translateY(10px); }
    to { opacity: 1; transform: translateY(0); }
  }

  .desktop-only {
    display: none;
  }

  /* Mobile Tabs */
  .mobile-tabs {
    display: flex;
    width: 100%;
    padding: 0 20px;
    margin-bottom: 20px;
  }

  .mobile-tab {
    flex: 1;
    text-align: center;
    padding: 15px 0;
    font-weight: 600;
    color: #9ca3af;
    position: relative;
    cursor: pointer;
    font-size: 1.1rem;
    transition: color 0.3s;
  }

  .mobile-tab.active {
    color: #f5576c;
  }

  .tab-indicator {
    position: absolute;
    bottom: 0;
    left: 50%;
    transform: translateX(-50%);
    width: 40px;
    height: 4px;
    background: linear-gradient(90deg, #f093fb, #f5576c);
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(245, 87, 108, 0.4);
  }
}
</style>
