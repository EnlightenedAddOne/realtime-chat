import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useUserStore } from './user'
import { sendMessage } from '../utils/ws'
import { ElMessage } from 'element-plus'

export const useCallStore = defineStore('call', () => {
  const userStore = useUserStore()

  // State
  const isCalling = ref(false)
  const incomingCall = ref(null) // { sender_id, sender_name, sender_avatar, offer }
  const localStream = ref(null)
  const remoteStream = ref(null)
  const peerConnection = ref(null)
  const currentPeerId = ref(null) // The ID of the person we are talking to

  // WebRTC Configuration
  const rtcConfig = {
    iceServers: [
      { urls: 'stun:stun.l.google.com:19302' }
    ]
  }

  // --- Actions ---

  // 1. Initiator: Start a call
  async function startCall(targetUser) {
    if (isCalling.value) return
    
    console.log('Starting call with:', targetUser)
    currentPeerId.value = targetUser.id
    isCalling.value = true

    try {
      // Get User Media
      localStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      
      // Create PeerConnection
      createPeerConnection()

      // Add Tracks
      localStream.value.getTracks().forEach(track => {
        peerConnection.value.addTrack(track, localStream.value)
      })

      // Create Offer
      const offer = await peerConnection.value.createOffer()
      await peerConnection.value.setLocalDescription(offer)

      // Send Offer Signal
      sendSignal('offer', { offer })

    } catch (err) {
      console.error('Failed to start call:', err)
      ElMessage.error('无法启动通话: ' + err.message)
      hangup()
    }
  }

  // 2. Receiver: Handle Incoming Signals
  async function handleSignal(payload) {
    const { type, sender_id, ...data } = payload
    console.log('Handle signal:', type, sender_id)

    // If we are already in a call with someone else, maybe auto-reject? 
    // For now, let's assume we can only handle one call.
    if (isCalling.value && currentPeerId.value !== sender_id && type === 'offer') {
       console.warn('Busy: rejecting call from', sender_id)
       // Busy signal? 
       // For now just ignore or auto-reject
       return
    }

    switch (type) {
      case 'offer':
        incomingCall.value = {
          sender_id,
          sender_name: data.sender_name, // Assuming backend sends this or we look it up
          sender_avatar: data.sender_avatar,
          offer: data.offer
        }
        break

      case 'answer':
        if (peerConnection.value) {
            const remoteDesc = new RTCSessionDescription(data.answer)
            await peerConnection.value.setRemoteDescription(remoteDesc)
        }
        break

      case 'candidate':
        if (peerConnection.value && data.candidate) {
            try {
                const candidate = new RTCIceCandidate(data.candidate)
                await peerConnection.value.addIceCandidate(candidate)
            } catch (e) {
                console.error('Error adding received ice candidate', e)
            }
        }
        break
      
      case 'reject':
        ElMessage.info('对方拒绝了通话')
        hangup()
        break

      case 'hangup':
        ElMessage.info('通话已结束')
        hangup()
        break
    }
  }

  // 3. Receiver: Accept Call
  async function acceptCall() {
    if (!incomingCall.value) return

    currentPeerId.value = incomingCall.value.sender_id
    isCalling.value = true
    
    try {
      // Get User Media
      localStream.value = await navigator.mediaDevices.getUserMedia({ video: true, audio: true })
      
      // Create PeerConnection
      createPeerConnection()

      // Add Tracks
      localStream.value.getTracks().forEach(track => {
        peerConnection.value.addTrack(track, localStream.value)
      })

      // Set Remote Description (Offer)
      const offer = new RTCSessionDescription(incomingCall.value.offer)
      await peerConnection.value.setRemoteDescription(offer)

      // Create Answer
      const answer = await peerConnection.value.createAnswer()
      await peerConnection.value.setLocalDescription(answer)

      // Send Answer Signal
      sendSignal('answer', { answer })
      
      // Clear incoming call state
      incomingCall.value = null

    } catch (err) {
      console.error('Failed to accept call:', err)
      ElMessage.error('无法接听通话')
      hangup()
    }
  }

  // 4. Receiver: Reject Call
  function rejectCall() {
    if (incomingCall.value) {
        sendSignal('reject', {}, incomingCall.value.sender_id)
        incomingCall.value = null
    }
  }

  // 5. Cleanup / Hangup
  function hangup() {
    // Send hangup signal if we were connected
    if (isCalling.value && currentPeerId.value) {
        sendSignal('hangup', {})
    }

    // Stop all tracks
    if (localStream.value) {
      localStream.value.getTracks().forEach(track => track.stop())
      localStream.value = null
    }
    
    // Close PeerConnection
    if (peerConnection.value) {
      peerConnection.value.close()
      peerConnection.value = null
    }

    remoteStream.value = null
    isCalling.value = false
    currentPeerId.value = null
    incomingCall.value = null
  }


  // --- Helpers ---

  function createPeerConnection() {
    peerConnection.value = new RTCPeerConnection(rtcConfig)

    // ICE Candidate Handler
    peerConnection.value.onicecandidate = (event) => {
      if (event.candidate) {
        sendSignal('candidate', { candidate: event.candidate })
      }
    }

    // Track Handler (Remote Stream)
    peerConnection.value.ontrack = (event) => {
      console.log('Received remote track')
      remoteStream.value = event.streams[0]
    }

    peerConnection.value.onconnectionstatechange = () => {
        console.log('Connection state:', peerConnection.value.connectionState)
        if (peerConnection.value.connectionState === 'disconnected' || 
            peerConnection.value.connectionState === 'failed' ||
            peerConnection.value.connectionState === 'closed') {
             // Maybe auto hangup?
        }
    }
  }

  function sendSignal(type, data, targetId = null) {
    const payload = {
      type: 'signal',
      receiver_id: targetId || currentPeerId.value,
      content: JSON.stringify({
         type,
         ...data,
         // Include extra info for 'offer' so receiver knows who calls
         ...(type === 'offer' ? { 
             sender_name: userStore.userInfo?.nickname || userStore.userInfo?.username,
             sender_avatar: userStore.userInfo?.avatar_url 
         } : {})
      })
    }
    sendMessage(payload)
  }

  return {
    isCalling,
    incomingCall,
    localStream,
    remoteStream,
    peerConnection,
    startCall,
    handleSignal,
    acceptCall,
    rejectCall,
    hangup
  }
})
