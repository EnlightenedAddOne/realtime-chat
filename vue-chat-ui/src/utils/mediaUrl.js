import { config } from '../config/index.js'

const ABSOLUTE_URL_PATTERN = /^(data:|blob:|https?:\/\/|wss?:\/\/|file:\/\/)/i

export function resolveMediaURL(url) {
  if (!url || typeof url !== 'string') return url

  const value = url.trim()
  if (!value || ABSOLUTE_URL_PATTERN.test(value)) return value

  const path = value.startsWith('/') ? value : `/${value}`
  try {
    return new URL(path, config.API_BASE_URL).toString()
  } catch {
    return value
  }
}

export function normalizeUserAvatar(user) {
  if (!user || typeof user !== 'object') return user
  return {
    ...user,
    avatar_url: resolveMediaURL(user.avatar_url)
  }
}

export function normalizeMessageMedia(message) {
  if (!message || typeof message !== 'object') return message

  const normalized = { ...message }
  if ([2, 3, 4, 5].includes(Number(normalized.msg_type))) {
    normalized.content = resolveMediaURL(normalized.content)
  }

  if (normalized.sender && typeof normalized.sender === 'object') {
    normalized.sender = {
      ...normalized.sender,
      avatar_url: resolveMediaURL(normalized.sender.avatar_url)
    }
  }

  return normalized
}