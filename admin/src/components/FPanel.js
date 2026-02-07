import { computed } from 'vue'
import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FPanel',
  props: {
    title: { type: String, required: true },
    count: { type: [Number, String], default: undefined },
    collapsed: { type: Boolean, default: false },
    mode: { type: String, default: 'fill' },
    height: { type: String, default: null },
  },
  emits: ['update:collapsed'],
  setup(props, { emit }) {
    useIcons()

    const isCollapsed = computed(() => props.collapsed)

    const panelStyle = computed(() => {
      if (isCollapsed.value) return 'flex: 0 0 auto;'
      if (props.mode === 'fill') return 'flex: 1 1 0; min-height: 0; display: flex; flex-direction: column;'
      if (props.mode === 'fixed') return `flex: 0 0 ${props.height}; display: flex; flex-direction: column;`
      return 'flex: 0 0 auto; display: flex; flex-direction: column;'
    })

    const bodyStyle = computed(() => {
      if (props.mode === 'fill' || props.mode === 'fixed') {
        return 'padding: 0; flex: 1; display: flex; flex-direction: column; min-height: 0;'
      }
      return ''
    })

    function toggle() {
      emit('update:collapsed', !props.collapsed)
    }

    return { isCollapsed, panelStyle, bodyStyle, toggle }
  },
  template: `
    <div class="panel-group" :class="{ collapsed: isCollapsed }">
      <div class="panel-group-card card" :style="panelStyle">
        <header class="panel-group-header">
          <button class="collapse-toggle" @click="toggle">
            <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
            <span class="text-heading text-primary">{{ title }}</span>
            <span v-if="count != null" class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">{{ count }}</span>
            <slot name="header-actions"></slot>
          </button>
        </header>
        <template v-if="!isCollapsed">
          <slot name="toolbar"></slot>
          <div class="panel-group-body" :style="bodyStyle">
            <slot></slot>
          </div>
          <slot name="footer"></slot>
        </template>
      </div>
    </div>
  `
}
