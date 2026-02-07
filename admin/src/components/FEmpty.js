import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FEmpty',
  props: {
    icon: { type: String, default: 'inbox' },
    title: { type: String, default: 'Nothing here' },
    message: { type: String, default: '' },
  },
  setup() {
    useIcons()
  },
  template: `
    <div class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
      <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
        <i :data-lucide="icon" class="w-6 h-6"></i>
      </div>
      <div class="text-heading text-primary mb-1">{{ title }}</div>
      <div v-if="message" class="text-caption text-muted">{{ message }}</div>
      <slot></slot>
    </div>
  `
}
