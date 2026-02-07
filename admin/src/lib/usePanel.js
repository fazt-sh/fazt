import { ref, reactive } from 'vue'
import { useUIStore } from '../stores/ui.js'

export function usePanel(key, defaultCollapsed = false) {
  const ui = useUIStore()
  const collapsed = ref(ui.getUIState(key, defaultCollapsed))

  function toggle() {
    collapsed.value = !collapsed.value
    ui.setUIState(key, collapsed.value)
  }

  return reactive({ collapsed, toggle })
}
