/**
 * AppPanel - Matches original renderPanel() output exactly
 */
export default {
  name: 'AppPanel',
  props: {
    id: { type: String, required: true },
    title: { type: String, required: true },
    count: { type: [Number, String], default: undefined },
    fillHeight: { type: Boolean, default: false },
    collapsed: { type: Boolean, default: false }
  },
  emits: ['toggle'],
  template: `
    <div class="panel-group" :class="{ collapsed: collapsed }">
      <div class="panel-group-card card"
           :class="{ 'flex flex-col': fillHeight }"
           :style="fillHeight ? 'flex:1;min-height:0;overflow:hidden' : ''">
        <header class="panel-group-header" :data-panel="id">
          <button class="collapse-toggle" @click="$emit('toggle')">
            <i data-lucide="chevron-right" class="chevron w-4 h-4"></i>
            <span class="text-heading text-primary">{{ title }}</span>
            <span v-if="count !== undefined" class="text-caption mono px-1.5 py-0.5 ml-2 badge-muted" style="border-radius: var(--radius-sm)">{{ count }}</span>
          </button>
        </header>
        <div v-if="$slots.toolbar && !collapsed"
             class="card-header flex items-center justify-between"
             style="flex-shrink: 0">
          <slot name="toolbar" />
        </div>
        <div class="panel-group-body"
             :style="fillHeight ? 'padding: 0; flex: 1; display: flex; flex-direction: column; min-height: 0' : 'padding: 0'">
          <div :style="fillHeight ? 'border: none; border-radius: 0; flex: 1; display: flex; flex-direction: column; min-height: 0' : 'border: none; border-radius: 0'">
            <div class="panel-scroll-area"
                 :class="fillHeight ? 'scroll-panel' : ''"
                 :style="fillHeight ? 'flex: 1; overflow: auto; min-height: 0' : 'min-height: 200px; max-height: 600px; overflow: auto'">
              <slot />
            </div>
            <div v-if="$slots.footer && !collapsed" style="flex-shrink: 0">
              <slot name="footer" />
            </div>
          </div>
        </div>
      </div>
    </div>
  `
}
