// Button Component
import { defineComponent } from 'vue'

export const Button = defineComponent({
  name: 'Button',

  props: {
    variant: {
      type: String,
      default: 'primary',
      validator: (v) => ['primary', 'secondary', 'danger', 'ghost'].includes(v)
    },
    size: {
      type: String,
      default: 'md',
      validator: (v) => ['sm', 'md', 'lg'].includes(v)
    },
    loading: {
      type: Boolean,
      default: false
    },
    disabled: {
      type: Boolean,
      default: false
    },
    block: {
      type: Boolean,
      default: false
    }
  },

  emits: ['click'],

  computed: {
    classes() {
      const base = 'inline-flex items-center justify-center font-medium rounded-lg transition-colors touch-feedback'

      const variants = {
        primary: 'bg-blue-500 hover:bg-blue-600 text-white',
        secondary: 'bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 text-neutral-900 dark:text-neutral-100',
        danger: 'bg-red-500 hover:bg-red-600 text-white',
        ghost: 'hover:bg-neutral-100 dark:hover:bg-neutral-800 text-neutral-700 dark:text-neutral-300'
      }

      const sizes = {
        sm: 'px-3 py-1.5 text-sm',
        md: 'px-4 py-2 text-sm',
        lg: 'px-6 py-3 text-base'
      }

      return [
        base,
        variants[this.variant],
        sizes[this.size],
        this.block ? 'w-full' : '',
        (this.disabled || this.loading) ? 'opacity-50 cursor-not-allowed' : ''
      ].filter(Boolean).join(' ')
    }
  },

  template: `
    <button
      :class="classes"
      :disabled="disabled || loading"
      @click="$emit('click', $event)"
    >
      <span v-if="loading" class="spinner mr-2"></span>
      <slot></slot>
    </button>
  `
})
