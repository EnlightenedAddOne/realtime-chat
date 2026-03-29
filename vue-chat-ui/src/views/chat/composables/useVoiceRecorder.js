import { computed, ref } from 'vue'

export function useVoiceRecorder(showError) {
  const isRecording = ref(false)
  const recordingDuration = ref(0)

  let mediaRecorder = null
  let audioChunks = []
  let recordingTimer = null

  const formattedDuration = computed(() => {
    const m = Math.floor(recordingDuration.value / 60).toString().padStart(2, '0')
    const s = (recordingDuration.value % 60).toString().padStart(2, '0')
    return `${m}:${s}`
  })

  function stopTimer() {
    if (!recordingTimer) return
    clearInterval(recordingTimer)
    recordingTimer = null
  }

  async function startRecording() {
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      mediaRecorder = new MediaRecorder(stream)
      audioChunks = []

      mediaRecorder.ondataavailable = (event) => {
        audioChunks.push(event.data)
      }

      mediaRecorder.start()
      isRecording.value = true
      recordingDuration.value = 0
      recordingTimer = setInterval(() => {
        recordingDuration.value++
      }, 1000)
    } catch (err) {
      console.error('Error accessing microphone:', err)
      showError?.('无法访问麦克风，请检查权限')
    }
  }

  function cancelRecording() {
    if (mediaRecorder && mediaRecorder.state !== 'inactive') {
      mediaRecorder.stop()
      mediaRecorder.stream.getTracks().forEach(track => track.stop())
    }
    stopTimer()
    isRecording.value = false
    audioChunks = []
  }

  function toggleVoice() {
    if (isRecording.value) return
    startRecording()
  }

  function stopAndGetBlob() {
    return new Promise((resolve, reject) => {
      if (!mediaRecorder || mediaRecorder.state === 'inactive') {
        reject(new Error('recorder inactive'))
        return
      }

      const duration = recordingDuration.value
      stopTimer()

      mediaRecorder.onstop = () => {
        try {
          const audioBlob = new Blob(audioChunks, { type: 'audio/webm' })
          mediaRecorder.stream.getTracks().forEach(track => track.stop())
          audioChunks = []
          isRecording.value = false
          resolve({ blob: audioBlob, duration })
        } catch (e) {
          reject(e)
        }
      }

      mediaRecorder.stop()
    })
  }

  return {
    isRecording,
    recordingDuration,
    formattedDuration,
    startRecording,
    cancelRecording,
    toggleVoice,
    stopAndGetBlob,
    stopTimer
  }
}
