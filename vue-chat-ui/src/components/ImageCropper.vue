<template>
  <el-dialog
    v-model="visible"
    title="调整并裁剪头像"
    width="500px"
    class="custom-dialog"
    append-to-body
    @opened="initCropper"
    @closed="destroyCropper"
  >
    <div class="cropper-container">
      <img ref="imageRef" :src="imageUrl" class="cropper-img" crossorigin="anonymous" />
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="visible = false" round>取消</el-button>
        <el-button type="primary" :loading="loading" @click="handleCrop" round>确认</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, watch } from 'vue'
import Cropper from 'cropperjs'
import 'cropperjs/dist/cropper.css'

const props = defineProps({
  modelValue: Boolean,
  imageUrl: String,
  loading: Boolean
})

const emit = defineEmits(['update:modelValue', 'crop'])

const visible = ref(props.modelValue)
watch(() => props.modelValue, (val) => {
  visible.value = val
})
watch(visible, (val) => {
  emit('update:modelValue', val)
})

const imageRef = ref(null)
let cropper = null

function initCropper() {
  if (imageRef.value) {
    cropper = new Cropper(imageRef.value, {
      aspectRatio: 1, // 1:1 for avatar
      viewMode: 1,
      dragMode: 'move',
      autoCropArea: 0.8,
      restore: false,
      guides: true,
      center: true,
      highlight: false,
      cropBoxMovable: true,
      cropBoxResizable: true,
      toggleDragModeOnDblclick: false,
    })
  }
}

function destroyCropper() {
  if (cropper) {
    cropper.destroy()
    cropper = null
  }
}

function handleCrop() {
  if (!cropper) return
  const canvas = cropper.getCroppedCanvas({
    width: 300,
    height: 300,
    imageSmoothingEnabled: true,
    imageSmoothingQuality: 'high',
  })
  
  canvas.toBlob((blob) => {
    emit('crop', blob)
  }, 'image/jpeg', 0.9)
}
</script>

<style scoped>
.cropper-container {
  width: 100%;
  height: 350px;
  background-color: #f7f9fc;
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.cropper-img {
  max-width: 100%;
  max-height: 100%;
  display: block;
}
</style>