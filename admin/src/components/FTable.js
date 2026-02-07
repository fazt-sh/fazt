import { useIcons } from '../lib/useIcons.js'

export default {
  name: 'FTable',
  props: {
    columns: { type: Array, required: true },
    rows: { type: Array, required: true },
    rowKey: { type: String, default: 'id' },
    clickable: { type: Boolean, default: true },
    emptyIcon: { type: String, default: 'inbox' },
    emptyTitle: { type: String, default: 'No data' },
    emptyMessage: { type: String, default: '' },
    loading: { type: Boolean, default: false },
  },
  emits: ['row-click'],
  setup() {
    useIcons()
  },
  template: `
    <div style="flex: 1; display: flex; flex-direction: column; min-height: 0;">
      <!-- Loading -->
      <div v-if="loading" class="flex items-center justify-center p-8">
        <div class="text-caption text-muted">Loading...</div>
      </div>
      <!-- Empty state -->
      <div v-else-if="rows.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
        <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
          <i :data-lucide="emptyIcon" class="w-6 h-6"></i>
        </div>
        <div class="text-heading text-primary mb-1">{{ emptyTitle }}</div>
        <div v-if="emptyMessage" class="text-caption text-muted">{{ emptyMessage }}</div>
        <slot name="empty"></slot>
      </div>
      <!-- Table -->
      <div v-else class="scroll-panel" style="flex: 1; overflow: auto; min-height: 0">
        <div class="table-container" style="overflow-x: visible">
          <table style="width: 100%; min-width: 0">
            <thead class="sticky" style="top: 0; background: var(--bg-1)">
              <tr style="border-bottom: 1px solid var(--border-subtle)">
                <th v-for="col in columns" :key="col.key"
                    class="px-3 py-2 text-left text-micro text-muted"
                    :class="{ 'hide-mobile': col.hideOnMobile }"
                    :style="col.width ? 'width:' + col.width : ''">
                  {{ col.label }}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(row, index) in rows" :key="row[rowKey] || index"
                  :class="clickable ? 'row row-clickable' : 'row'"
                  style="border-bottom: 1px solid var(--border-subtle)"
                  @click="clickable && $emit('row-click', row)">
                <slot name="row" :row="row" :index="index">
                  <td v-for="col in columns" :key="col.key"
                      class="px-3 py-2"
                      :class="{ 'hide-mobile': col.hideOnMobile }">
                    <slot :name="'cell-' + col.key" :row="row" :value="row[col.key]">
                      <span class="text-caption text-muted">{{ row[col.key] }}</span>
                    </slot>
                  </td>
                </slot>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  `
}
