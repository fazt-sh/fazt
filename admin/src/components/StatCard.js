import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'StatCard',
  props: {
    label: { type: String, required: true },
    icon: { type: String, default: '' },
    value: { type: [String, Number], default: '' },
    subtitle: { type: String, default: '' },
    valueClass: { type: String, default: 'text-primary' },
    clickable: { type: Boolean, default: false },
  },
  emits: ['click'],
  setup() {
    useIcons()
  },
  template: `
    <div class="stat-card card" :class="{ 'row-clickable': clickable }" @click="clickable && $emit('click')">
      <div class="stat-card-header">
        <span class="text-micro text-muted">{{ label }}</span>
        <i v-if="icon" :data-lucide="icon" class="w-4 h-4 text-faint"></i>
      </div>
      <div class="stat-card-value text-display mono" :class="valueClass">
        <slot name="value">{{ value }}</slot>
      </div>
      <div class="stat-card-subtitle text-caption text-muted">{{ subtitle }}</div>
    </div>
  `
}
