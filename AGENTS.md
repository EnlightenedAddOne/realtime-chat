# AGENTS.md - Developer Guide & Rules

This document defines the coding standards, patterns, and workflows for AI agents and developers working on the `realtime-chat` repository.

## 1. Project Overview

This is a monorepo containing a real-time chat application with a Go backend and a Vue.js frontend.

- **Backend**: `go-im-server/` (Go 1.25+, Gin, GORM, WebSocket)
- **Frontend**: `vue-chat-ui/` (Vue 3, Vite, Pinia, Element Plus)

## 2. Backend Guidelines (`go-im-server/`)

### Architecture
Follow the standard Clean Architecture / Layered pattern:
- `cmd/`: Entry points (`main.go`, `server.go`).
- `internal/api/`: HTTP handlers and routers. Receives HTTP requests, validates input, calls service.
- `internal/service/`: Business logic. Orchestrates repositories and other services.
- `internal/repository/`: Data access layer (DB operations using GORM).
- `internal/model/`: Domain entities and DTOs.
- `internal/ws/`: WebSocket logic (Hub, Client).
- `pkg/`: Public shared packages (utils like `logger`, `jwt`, `ecode`).
- `config/`: Configuration handling (`viper`).

### Code Style (Go)
- **Formatting**: Always use `gofmt` (or `goimports`).
- **Naming**:
  - Use `CamelCase` for exported identifiers, `camelCase` for private ones.
  - Acronyms should be all caps (e.g., `ServeHTTP`, `ID`, `URL`) or consistent.
  - Interface names should usually end in `er` (e.g., `Repository`, `Service`).
- **Import Organization** (use `goimports` automatically):
  ```go
  import (
      "net/http"           // stdlib
      "strings"            // stdlib

      "go-im-server/internal/repository"  // internal packages
      "go-im-server/internal/service"

      "github.com/gin-gonic/gin"  // external packages
      "gorm.io/gorm"
  )
  ```
- **Error Handling**:
  - Check errors explicitly: `if err != nil { return nil, err }`.
  - Wrap errors with context where useful: `fmt.Errorf("failed to create user: %w", err)`
  - Return HTTP errors from handlers using `c.JSON(code, gin.H{"message": err.Error()})`.
  - Define custom error types in service layer (e.g., `service.ErrUserExists`).
- **Dependency Injection**:
  - Use constructor functions (e.g., `NewAuthHandler(s *service.AuthService)`) to inject dependencies.
  - Avoid global state.
- **DB Transactions**: Use `db.Transaction()` for multi-table operations.

### Libraries
- **Web Framework**: Gin (`github.com/gin-gonic/gin`)
- **ORM**: GORM (`gorm.io/gorm`)
- **WebSockets**: Gorilla (`github.com/gorilla/websocket`)
- **Config**: Viper or simple env vars (check `config/`). Note: The actual `config.yaml` is git-ignored for security. Use `config.example.yaml` as a template.

### Database Schema (Quick Reference)
- **Users**: `id`, `username`, `password_hash`, `nickname`, `avatar_url`
- **Messages**: `id`, `sender_id`, `receiver_id`, `msg_type` (1:text, 2:image, 3:voice, 4:video), `content`, `is_read`
- **Friends**: `user_id`, `friend_id`, `status`

### Commands
Run these from `go-im-server/`:

```bash
# Run locally
go run ./cmd

# Build for Windows
go build -o deploy/im-server.exe ./cmd

# Build for Linux amd64
GOOS=linux GOARCH=amd64 go build -o deploy/im-server ./cmd

# Run all tests
go test ./...

# Run single test (IMPORTANT for debugging)
go test -v -run TestUserService_CreateUser ./internal/service/

# Run tests with coverage
go test -v -cover ./internal/service/

# Run tests matching pattern
go test -v -run "Test.*" ./internal/repository/

# Watch mode (requires gotest)
go test -v -run TestName -count=1 ./internal/package/path

# Format code
go fmt ./...
goimports -w .

# Lint (if golangci-lint installed)
golangci-lint run
```

## 3. Frontend Guidelines (`vue-chat-ui/`)

### Architecture
- `src/views/`: Page-level components (`Login.vue`, `Chat.vue`).
- `src/components/`: Reusable UI components.
- `src/store/`: Pinia state management (`user.js`, `chat.js`).
- `src/api/`: Axios instances and API request functions (`index.js`).
- `src/utils/`: Helper functions (`ws.js` for WebSocket logic).

### Code Style (Vue/JS)
- **Component Style**: Use **Composition API** with `<script setup>`.
- **Formatting**: 2 spaces indentation, semicolons optional (prefer consistent style), single quotes.
- **Naming**:
  - Components: `PascalCase.vue` (e.g., `ChatWindow.vue`).
  - Props: `camelCase`.
  - Events: `kebab-case` in templates, `camelCase` in emitters.
- **State Management**: Use Pinia. Store files in `src/store/*.js`.
- **UI Library**: Element Plus. Use `<el-*>` components.

### Example Component Pattern
```vue
<template>
  <div class="my-component">
    <el-button @click="handleClick">{{ label }}</el-button>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useUserStore } from '../store/user'

const props = defineProps({
  label: String
})

const emit = defineEmits(['submit'])
const userStore = useUserStore()

function handleClick() {
  emit('submit', userStore.userInfo)
}
</script>

<style scoped>
.my-component {
  /* styles */
}
</style>
```

### Error Handling in Vue
- Use try/catch with async/await for API calls
- Display errors with Element Plus `ElMessage.error()`
- Handle WebSocket reconnections in `utils/ws.js`
- Use Pinia for global error state management

### Commands
Run these from `vue-chat-ui/`:

```bash
# Install dependencies
npm install

# Start development server
npm run dev

# Build for production
npm run build

# Preview build
npm run preview

# Electron development (run both Vite and Electron)
npm run electron:dev

# Electron build (Windows NSIS installer)
npm run electron:build
```

## 4. General Workflows

### Git Commit Messages
Format: `type(scope): description`
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `refactor`: Code restructuring
- `style`: Formatting (no logic change)
- `chore`: Build/dep updates

Example: `feat(auth): add JWT token refresh mechanism`

### Testing
- **Backend**: Write table-driven tests for services and repositories.
- **Frontend**: Currently no automated test suite configured. Ensure manual verification of UI components.

### Implementation Rules
1. **Minimal Changes**: Only modify files necessary for the task.
2. **Type Safety**: Prefer explicit types in Go. Use proper prop validation in Vue.
3. **No Magic Numbers**: Extract constants.
4. **Comments**: Document complex logic, exported functions, and API endpoints.

## 5. Environment
- **Go Version**: 1.25.0
- **Node Version**: LTS Recommended
- **Database**: PostgreSQL (via GORM)

## 6. Critical Paths
- **Authentication**: JWT based. Handled in `internal/api/auth_handler.go` and `store/user.js`.
- **Real-time**: WebSocket implementation in `internal/ws/` and `utils/ws.js`. Ensure connection stability and reconnection logic.

---
*Generated by OpenCode Sisyphus Agent*
