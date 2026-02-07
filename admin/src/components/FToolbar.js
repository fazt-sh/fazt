import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FToolbar',
  props: {
    searchPlaceholder: { type: String, default: 'Filter...' },
    modelValue: { type: String, default: '' },
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    useIcons()

    function onInput(e) {
      emit('update:modelValue', e.target.value)
    }

    return { onInput }
  },
  template: `
    <div class="card-header flex items-center justify-between" style="flex-shrink: 0">
      <div class="input toolbar-search">
        <i data-lucide="search" class="w-4 h-4 text-faint"></i>
        <input type="text" :placeholder="searchPlaceholder" :value="modelValue" @input="onInput">
      </div>
      <div class="flex items-center gap-2" style="flex-shrink: 0">
        <slot name="filters"></slot>
        <slot name="actions"></slot>
      </div>
    </div>
  `
}
