/**
 * 应用配置文件
 * 从 .env 文件中读取环境变量
 */

export const config = {
  // API 服务器地址
  API_BASE_URL: import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080',
  
  // WebSocket 服务器地址
  WS_BASE_URL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:9527',
  
  // 应用名称
  APP_NAME: import.meta.env.VITE_APP_NAME || 'RealTime Chat',
  
  // 应用版本
  APP_VERSION: import.meta.env.VITE_APP_VERSION || '1.0.0',
  
  // 其他常用配置
  REQUEST_TIMEOUT: 10000, // 请求超时时间（毫秒）
  WS_RECONNECT_ATTEMPTS: 5, // WebSocket 重连次数
  WS_RECONNECT_INTERVAL: 3000, // WebSocket 重连间隔（毫秒）
}

// 打印配置信息（仅在开发环境）
if (import.meta.env.DEV) {
  console.log('🔧 应用配置:', {
    API_BASE_URL: config.API_BASE_URL,
    WS_BASE_URL: config.WS_BASE_URL,
    APP_NAME: config.APP_NAME,
    APP_VERSION: config.APP_VERSION,
  })
}
