import axios from 'axios'
import { ElMessage } from 'element-plus'
import { config } from '../config/index.js'

const api = axios.create({
  baseURL: config.API_BASE_URL,
  timeout: config.REQUEST_TIMEOUT
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    return response.data
  },
  error => {
    ElMessage.error(error.response?.data?.message || '请求失败')
    return Promise.reject(error)
  }
)

// Auth API
export const authAPI = {
  // 发送验证码
  sendCode(email, type = 'register') {
    return api.post('/api/auth/send-code', { email, type })
  },
  
  // 验证验证码
  verifyCode(email, code, type = 'register') {
    return api.post('/api/auth/verify-code', { email, code, type })
  },
  
  // 注册
  register(data) {
    return api.post('/api/register', data)
  },
  
	// 登录
	login(email, password) {
		return api.post('/api/login', { email, password })
	}
}

export default api
