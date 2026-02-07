import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FPagination',
  props: {
    currentPage: { type: Number, required: true },
    totalPages: { type: Number, required: true },
  },
  emits: ['page-change'],
  setup(props, { emit }) {
    useIcons()

    function goFirst() { emit('page-change', 1) }
    function goPrev() { if (props.currentPage > 1) emit('page-change', props.currentPage - 1) }
    function goNext() { if (props.currentPage < props.totalPages) emit('page-change', props.currentPage + 1) }
    function goLast() { emit('page-change', props.totalPages) }

    return { goFirst, goPrev, goNext, goLast }
  },
  template: `
    <div v-if="totalPages > 1" class="flex items-center gap-2">
      <button class="btn btn-secondary btn-sm" :disabled="currentPage === 1" title="First page" @click="goFirst">
        <i data-lucide="chevrons-left" class="w-3.5 h-3.5"></i>
      </button>
      <button class="btn btn-secondary btn-sm" :disabled="currentPage === 1" title="Previous" @click="goPrev">
        <i data-lucide="chevron-left" class="w-3.5 h-3.5"></i>
      </button>
      <span class="text-caption text-muted px-2">Page {{ currentPage }} of {{ totalPages }}</span>
      <button class="btn btn-secondary btn-sm" :disabled="currentPage === totalPages" title="Next" @click="goNext">
        <i data-lucide="chevron-right" class="w-3.5 h-3.5"></i>
      </button>
      <button class="btn btn-secondary btn-sm" :disabled="currentPage === totalPages" title="Last page" @click="goLast">
        <i data-lucide="chevrons-right" class="w-3.5 h-3.5"></i>
      </button>
    </div>
  `
}
