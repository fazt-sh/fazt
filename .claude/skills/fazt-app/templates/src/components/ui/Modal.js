// Modal Component
import { defineComponent } from 'vue'

export const Modal = defineComponent({
  name: 'Modal',

  props: {
    show: {
      type: Boolean,
      required: true
    },
    title: {
      type: String,
      default: ''
    },
    size: {
      type: String,
      default: 'md',
      validator: (v) => ['sm', 'md', 'lg', 'full'].includes(v)
    }
  },

  emits: ['close'],

  computed: {
    sizeClasses() {
      const sizes = {
        sm: 'max-w-sm',
        md: 'max-w-md',
        lg: 'max-w-lg',
        full: 'max-w-full mx-4'
      }
      return sizes[this.size]
    }
  },

  template: `
    <teleport to="body">
      <transition name="modal">
        <div
          v-if="show"
          class="fixed inset-0 z-50 flex items-center justify-center p-4"
          @click.self="$emit('close')"
        >
          <!-- Backdrop -->
          <div
            class="absolute inset-0 bg-black/50 backdrop-blur-sm"
            @click="$emit('close')"
          ></div>

          <!-- Modal Content -->
          <div
            :class="[
              'relative w-full bg-white dark:bg-neutral-900 rounded-2xl shadow-xl',
              sizeClasses
            ]"
          >
            <!-- Header -->
            <div v-if="title || $slots.header" class="flex items-center justify-between p-4 border-b border-neutral-200 dark:border-neutral-800">
              <slot name="header">
                <h2 class="text-lg font-semibold">{{ title }}</h2>
              </slot>
              <button
                @click="$emit('close')"
                class="p-2 -mr-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800"
              >
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
                </svg>
              </button>
            </div>

            <!-- Body -->
            <div class="p-4">
              <slot></slot>
            </div>

            <!-- Footer -->
            <div v-if="$slots.footer" class="flex items-center justify-end gap-2 p-4 border-t border-neutral-200 dark:border-neutral-800">
              <slot name="footer"></slot>
            </div>
          </div>
        </div>
      </transition>
    </teleport>
  `
})
