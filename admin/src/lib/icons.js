/**
 * Lucide icon helper
 * Re-renders Lucide icons after Vue DOM updates
 */
import { nextTick } from 'vue'

export function refreshIcons() {
  nextTick(() => {
    if (window.lucide) window.lucide.createIcons()
  })
}
