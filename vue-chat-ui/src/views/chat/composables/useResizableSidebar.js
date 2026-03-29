import { ref } from 'vue'

export function useResizableSidebar(miniSidebarWidth = 60, minWidth = 200, maxWidth = 500, defaultWidth = 250) {
  const sidebarWidth = ref(defaultWidth)
  const isResizing = ref(false)

  function startResize() {
    isResizing.value = true
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
  }

  function onMouseMove(e) {
    if (!isResizing.value) return
    const newWidth = e.clientX - miniSidebarWidth
    if (newWidth >= minWidth && newWidth <= maxWidth) {
      sidebarWidth.value = newWidth
    }
  }

  function stopResize() {
    if (!isResizing.value) return
    isResizing.value = false
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
  }

  return {
    sidebarWidth,
    isResizing,
    startResize,
    onMouseMove,
    stopResize
  }
}
