// Card Component
import { defineComponent } from 'vue'

export const Card = defineComponent({
  name: 'Card',

  props: {
    padding: {
      type: String,
      default: 'md',
      validator: (v) => ['none', 'sm', 'md', 'lg'].includes(v)
    },
    hoverable: {
      type: Boolean,
      default: false
    }
  },

  computed: {
    classes() {
      const base = 'bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg'

      const paddings = {
        none: '',
        sm: 'p-3',
        md: 'p-4',
        lg: 'p-6'
      }

      const hover = this.hoverable
        ? 'transition-shadow hover:shadow-md cursor-pointer'
        : ''

      return [base, paddings[this.padding], hover].filter(Boolean).join(' ')
    }
  },

  template: `
    <div :class="classes">
      <slot></slot>
    </div>
  `
})
