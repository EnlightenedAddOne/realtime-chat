# Realtime Chat

一个基于 Go + Vue 3 的实时通讯项目（单仓库 Monorepo），包含：
- 后端服务：`go-im-server`（HTTP API + WebSocket + PostgreSQL）
- 前端界面：`vue-chat-ui`（Vue 3 + Vite）
- 桌面打包：Electron Windows 安装程序（支持安装路径选择）

## 项目亮点

- 实时通信：WebSocket 长连接，支持私聊与群聊实时收发。
- 音视频通话：基于 WebRTC（STUN）实现点对点呼叫。
- 多媒体消息：支持文本、图片、语音、视频、文件消息。
- 文件上传与下载：文件可上传到服务端并在客户端下载到本地。
- 可配置清理策略：上传目录支持按配置定时清理过期文件。
- 一套代码多端运行：浏览器 + Electron 桌面端。
- 清晰分层架构：后端采用分层设计，前端采用 Composition API + Pinia。

## 技术栈

### 后端（`go-im-server`）
- Go 1.25
- Gin（HTTP 路由）
- Gorilla WebSocket（实时通信）
- GORM + PostgreSQL（数据持久化）
- JWT（认证）
- YAML 配置

### 前端（`vue-chat-ui`）
- Vue 3 + Vite
- Pinia（状态管理）
- Vue Router
- Axios
- Element Plus
- Electron + electron-builder（桌面安装包）

## 项目结构

```text
realtime-chat/
├─ go-im-server/                 # Go 后端
│  ├─ cmd/                       # 入口（main.go, server.go）
│  ├─ config/                    # 配置（config.example.yaml, config.go）
│  ├─ internal/
│  │  ├─ api/                    # HTTP/WS Handler
│  │  ├─ model/                  # 数据模型
│  │  ├─ repository/             # 数据访问层
│  │  ├─ service/                # 业务服务（含上传清理任务）
│  │  └─ ws/                     # WebSocket Hub/Client
│  ├─ pkg/                       # 通用工具（jwt/logger/email/security）
│  └─ deploy/                    # 部署包与脚本
├─ vue-chat-ui/                  # Vue 前端 + Electron
│  ├─ electron/                  # Electron 主进程
│  ├─ src/
│  │  ├─ api/                    # Axios 封装
│  │  ├─ store/                  # Pinia（user/chat/call）
│  │  ├─ utils/                  # ws/mediaUrl 等工具
│  │  └─ views/                  # Login/Chat 页面
│  ├─ .env.*                     # 前端环境配置
│  └─ release/                   # Electron 打包输出
└─ ref_chat_repo/                # 参考代码（非运行主工程）
```

## 核心功能

- 认证与用户
  - 用户注册、登录（JWT）
  - 用户资料与头像修改

- 好友体系
  - 搜索用户
  - 发送/处理好友请求
  - 删除好友

- 会话与消息
  - 私聊/群聊会话列表
  - 历史消息查询
  - 已读状态同步
  - 消息类型：
    - `1` 文本
    - `2` 图片
    - `3` 语音
    - `4` 视频
    - `5` 文件

- 群组能力
  - 创建群组、搜索群组、申请入群
  - 邀请入群、审批入群请求
  - 群成员管理、管理员管理、群主转让、解散群

- 实时能力
  - WebSocket 消息实时收发
  - WebRTC 音视频通话信令转发（offer/answer/candidate）

- 文件能力
  - `/api/upload` 上传图片/音视频/文档/压缩包等
  - 前端支持下载图片、视频、文件到本地
  - 后端可定时清理云端过期文件



## 界面展示

| 注册界面 | 登录界面 |
| :---: | :---: |
| <img width="100%" alt="注册界面" src="https://github.com/user-attachments/assets/9f5e5c16-7dfd-4b2f-a83e-87ed4500d270" /> | <img width="100%" alt="登录界面" src="https://github.com/user-attachments/assets/4af4b90c-3b55-4d58-bef6-d3c9f91c75b0" /> |

| 亮色主题下的清新绿配色 | 亮色主题下的粉蓝渐变配色 |
| :---: | :---: |
| <img width="100%" alt="亮色主题下的清新绿配色" src="https://github.com/user-attachments/assets/2eba9548-eb13-4c0b-a86c-06bc83357b21" /> | <img width="100%" alt="亮色主题下的粉蓝渐变配色" src="https://github.com/user-attachments/assets/11abe9be-ebc2-407e-8403-5888af8e6ff7" /> |

| 暗色主题界面一 | 暗色主题界面二 |
| :---: | :---: |
| <img width="100%" alt="暗色主题界面一" src="https://github.com/user-attachments/assets/a04a93b0-0b07-437c-951c-29b7745acccd" /> | <img width="100%" alt="暗色主题界面二" src="https://github.com/user-attachments/assets/6c633446-4121-4aa8-9a66-7ea45cde704c" /> |

| 设置界面 |
| :---: |
| <img width="100%" alt="设置界面" src="https://github.com/user-attachments/assets/cf511359-70ee-4125-8ed9-db2e78b7fc63" /> |

## 架构说明

### 后端架构（分层）
- `api`：接收请求、参数校验、返回响应
- `service`：业务编排（如上传清理 worker）
- `repository`：数据库读写
- `model`：领域模型
- `ws`：连接管理与消息广播

消息链路：
1. 客户端通过 `/ws` 建立连接。
2. Hub 接收消息并持久化（`messages`）。
3. 更新会话摘要与未读数（`conversations`）。
4. 转发给接收方/群成员。

### 前端架构
- `views/Chat.vue`：聊天主页面
- `views/chat/composables/*`：按功能拆分逻辑（消息、群组、好友、语音等）
- `store/chat.js`：消息与会话状态
- `store/call.js`：WebRTC 通话状态
- `utils/ws.js`：WebSocket 连接与消息分发
- `utils/mediaUrl.js`：统一处理资源 URL（兼容 Electron `file://` 场景）

## 配置说明

### 后端配置（`go-im-server/config/config.yaml`）
（*注：代码库中提供的是 `config.example.yaml`，请在本地复制并重命名为 `config.yaml` 后使用。*）

关键字段：

```yaml
server:
  port: 8080
  mode: debug

upload:
  path: ./uploads
  max_size: 10
  cleanup_enabled: true
  cleanup_interval_hours: 24
  cleanup_max_age_hours: 168
```

说明：
- `max_size`：上传文件大小上限（MB）
- `cleanup_enabled`：是否开启定时清理
- `cleanup_interval_hours`：清理执行间隔
- `cleanup_max_age_hours`：文件最大保留时长（超过即删除）

### 前端配置（`vue-chat-ui/.env.*`）

```env
VITE_API_BASE_URL=http://localhost:8080
VITE_WS_BASE_URL=ws://localhost:8080
VITE_APP_NAME=RealTime Chat
VITE_APP_VERSION=1.0.0
```

## 本地开发

### 1) 启动后端

```bash
cd go-im-server
go test ./...
go run ./cmd
```

默认健康检查：`GET /health`

### 2) 启动前端

```bash
cd vue-chat-ui
npm install
npm run dev
```

## 打包命令

### 后端打包

在 `go-im-server` 目录执行：

```bash
# Windows
go build -o deploy/im-server.exe ./cmd

# Linux amd64
GOOS=linux GOARCH=amd64 go build -o deploy/im-server ./cmd
```

PowerShell 示例：

```powershell
go build -o deploy/im-server.exe ./cmd
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -o deploy/im-server ./cmd
Remove-Item Env:GOOS
Remove-Item Env:GOARCH
```

### 前端 Web 打包

在 `vue-chat-ui` 目录执行：

```bash
npm run build
```

### 前端 Electron 打包（Windows 安装程序）

```bash
npm run electron:build
```

输出目录：`vue-chat-ui/release/`

安装器特性：
- NSIS 向导模式（`oneClick=false`）
- 支持安装时选择安装路径
- 创建桌面和开始菜单快捷方式

## 部署流程（建议）

### 后端部署（Linux）
1. 上传 `go-im-server/deploy/im-server` 二进制文件到服务器（如 `/opt/im-server`）。
2. 在同级目录创建 `config/config.yaml`（参考 `config.example.yaml` 填写数据库、端口、JWT等）。
3. 执行 `start-daemon.sh` 后台启动。
4. 用 `status.sh` 查看运行状态，用 `stop.sh` 停止。

`go-im-server/deploy/DEPLOY.md` 中提供了 systemd、Nginx、日志轮转等完整示例。

### 前端部署
- 浏览器方案：将 `vue-chat-ui/dist` 托管到静态服务器（如 Nginx）。
- 桌面方案：分发 `vue-chat-ui/release/RealTimeChat-Setup-*.exe`。

## 常用接口（摘要）

- 认证：`/api/register`, `/api/login`
- 用户：`/api/user/profile`
- 好友：`/api/friends`, `/api/friends/search`, `/api/friends/request`
- 会话：`/api/conversations`, `/api/conversations/read`
- 消息：`/api/messages`
- 上传：`/api/upload`
- 群组：`/api/groups*`, `/api/search/groups`
- 实时：`/ws`
- 静态资源：`/uploads/*`

## 安全与生产建议

- 生产环境务必替换：
  - `jwt.secret`
  - 数据库账号密码
- 建议将 `server.mode` 设为 `release`
- 建议在 Nginx 层配置 HTTPS 与反向代理
- 上传目录清理策略请结合业务保留周期配置

## 许可证

当前仓库未声明独立 LICENSE，请按团队内部规范使用。
