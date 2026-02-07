import { onMounted, onUpdated } from 'vue'
import { refreshIcons } from './icons.js'

export function useIcons() {
  onMounted(() => refreshIcons())
  onUpdated(() => refreshIcons())
}
