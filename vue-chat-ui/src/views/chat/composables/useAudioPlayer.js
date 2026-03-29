import { ref } from 'vue'

export function useAudioPlayer(showError) {
  const currentAudio = ref(null)
  const playingUrl = ref('')

  function stopCurrentAudio() {
    if (!currentAudio.value) return
    currentAudio.value.pause()
    currentAudio.value = null
    playingUrl.value = ''
  }

  function playVoice(url) {
    if (!url) return

    if (currentAudio.value && playingUrl.value === url) {
      stopCurrentAudio()
      return
    }

    if (currentAudio.value) {
      currentAudio.value.pause()
      currentAudio.value = null
    }

    const audio = new Audio(url)
    currentAudio.value = audio
    playingUrl.value = url

    audio.play().catch(() => {
      showError?.('播放失败')
      stopCurrentAudio()
    })

    audio.onended = () => {
      stopCurrentAudio()
    }

    audio.onerror = () => {
      showError?.('音频加载失败')
      stopCurrentAudio()
    }
  }

  return {
    currentAudio,
    playingUrl,
    playVoice,
    stopCurrentAudio
  }
}
