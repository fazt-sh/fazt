import { ref, onMounted, onBeforeUnmount } from 'vue'
import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FilterDropdown',
  props: {
    label: { type: String, required: true },
    options: { type: Array, required: true },
    modelValue: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    useIcons()

    const isOpen = ref(false)
    const btnRect = ref(null)

    function toggle(event) {
      if (isOpen.value) {
        isOpen.value = false
        return
      }
      btnRect.value = event.currentTarget.getBoundingClientRect()
      isOpen.value = true
    }

    function select(value) {
      emit('update:modelValue', value)
      isOpen.value = false
    }

    function closeOnOutsideClick() {
      isOpen.value = false
    }

    onMounted(() => {
      document.addEventListener('click', closeOnOutsideClick)
    })

    onBeforeUnmount(() => {
      document.removeEventListener('click', closeOnOutsideClick)
    })

    return { isOpen, btnRect, toggle, select }
  },
  template: `
    <div class="relative hide-mobile">
      <button class="btn btn-secondary btn-sm flex items-center gap-1.5" style="padding: 4px 8px" @click.stop="toggle($event)">
        <span class="text-caption">{{ label }}</span>
        <i data-lucide="chevron-down" class="w-3.5 h-3.5" style="color:var(--text-3)"></i>
      </button>
      <div v-if="isOpen" class="dropdown fixed z-50" :style="{ top: (btnRect?.bottom + 4) + 'px', left: btnRect?.left + 'px', minWidth: '140px' }">
        <div v-for="opt in options" :key="opt.value"
             class="dropdown-item px-4 py-2 cursor-pointer text-caption"
             @click.stop="select(opt.value)">
          {{ opt.label }}
        </div>
      </div>
    </div>
  `
}
