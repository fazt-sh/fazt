/**
 * PanelToolbar - Matches original renderToolbar() output exactly
 */
export default {
  name: 'PanelToolbar',
  props: {
    searchId: { type: String, default: 'filter-input' },
    searchPlaceholder: { type: String, default: 'Filter...' },
    modelValue: { type: String, default: '' }
  },
  emits: ['update:modelValue', 'search'],
  methods: {
    onInput(e) {
      this.$emit('update:modelValue', e.target.value)
      this.$emit('search', e.target.value)
    }
  },
  template: `
    <div class="input toolbar-search">
      <i data-lucide="search" class="w-4 h-4 text-faint"></i>
      <input type="text" :id="searchId" :placeholder="searchPlaceholder" :value="modelValue" @input="onInput">
    </div>
    <div v-if="$slots.actions" class="flex items-center gap-2" style="flex-shrink: 0">
      <slot name="actions" />
    </div>
  `
}
