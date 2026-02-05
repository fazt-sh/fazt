/**
 * DataTable - Matches original renderTable() output exactly
 */
export default {
  name: 'DataTable',
  props: {
    columns: { type: Array, required: true },
    data: { type: Array, required: true },
    rowKey: { type: String, default: 'id' },
    clickable: { type: Boolean, default: false },
    emptyIcon: { type: String, default: 'inbox' },
    emptyTitle: { type: String, default: 'No data' },
    emptyMessage: { type: String, default: '' }
  },
  emits: ['row-click'],
  computed: {
    visibleColumns() {
      return this.columns.filter(col => !col.showOnMobile)
    }
  },
  template: `
    <div v-if="data.length === 0" class="flex flex-col items-center justify-center p-8 text-center" style="min-height: 200px">
      <div class="icon-box mb-3" style="width:48px;height:48px;opacity:0.5">
        <i :data-lucide="emptyIcon" class="w-6 h-6"></i>
      </div>
      <div class="text-heading text-primary mb-1">{{ emptyTitle }}</div>
      <div v-if="emptyMessage" class="text-caption text-muted">{{ emptyMessage }}</div>
    </div>
    <div v-else class="table-container" style="overflow-x: visible">
      <table style="width: 100%; min-width: 0">
        <thead class="sticky" style="top: 0; background: var(--bg-1)">
          <tr style="border-bottom: 1px solid var(--border-subtle)">
            <th v-for="col in visibleColumns"
                :key="col.key"
                class="px-3 py-2 text-left text-micro text-muted"
                :class="{ 'hide-mobile': col.hideOnMobile }"
                :style="col.width ? 'width: ' + col.width : ''">
              {{ col.label }}
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in data"
              :key="row[rowKey]"
              :class="clickable ? 'row row-clickable' : 'row'"
              style="border-bottom: 1px solid var(--border-subtle)"
              @click="clickable && $emit('row-click', row)">
            <td v-for="col in visibleColumns"
                :key="col.key"
                class="px-3 py-2"
                :class="{ 'hide-mobile': col.hideOnMobile }">
              <slot :name="'cell-' + col.key" :row="row" :value="row[col.key]">
                {{ row[col.key] }}
              </slot>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  `
}
