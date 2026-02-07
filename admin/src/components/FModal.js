import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FModal',
  props: {
    open: { type: Boolean, required: true },
    title: { type: String, required: true },
    subtitle: { type: String, default: '' },
    icon: { type: String, default: '' },
    maxWidth: { type: String, default: 'max-w-md' },
  },
  emits: ['update:open', 'close'],
  setup(props, { emit }) {
    useIcons()

    function close() {
      emit('update:open', false)
      emit('close')
    }

    return { close }
  },
  template: `
    <div v-if="open" class="fixed inset-0 z-50 flex items-center justify-center modal-backdrop" @click="close">
      <div class="modal w-full p-6 relative" :class="maxWidth" @click.stop>
        <div v-if="icon || title" class="text-center mb-6">
          <div v-if="icon" class="w-12 h-12 mx-auto mb-4 flex items-center justify-center" style="background:var(--accent-soft);border-radius:var(--radius-md)">
            <i :data-lucide="icon" class="w-6 h-6" style="color:var(--accent)"></i>
          </div>
          <h2 class="text-title" style="color:var(--text-1)">{{ title }}</h2>
          <p v-if="subtitle" class="text-caption mt-1" style="color:var(--text-3)">{{ subtitle }}</p>
        </div>
        <slot></slot>
        <div v-if="$slots.footer" class="flex gap-2 mt-6">
          <slot name="footer"></slot>
        </div>
        <button class="absolute top-4 right-4 p-1 btn-ghost" style="color:var(--text-3);border-radius:var(--radius-sm)" @click="close">
          <i data-lucide="x" class="w-5 h-5"></i>
        </button>
      </div>
    </div>
  `
}
