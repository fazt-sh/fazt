import { useUIStore } from '../stores/ui.js'
import { useIcons } from '../lib/useIcons.js'
import { palettes } from '../lib/palettes.js'

export default {
  name: 'SettingsPanel',
  setup() {
    useIcons()
    const ui = useUIStore()

    return { ui, palettes }
  },
  template: `
    <div v-if="ui.settingsPanelOpen" id="settingsPanel" class="settings-panel fixed" style="z-index: 1100; bottom: 20px; left: 50%; transform: translateX(-50%);">
      <div class="p-4 flex items-center gap-6">
        <div class="flex items-center gap-3">
          <span class="text-micro" style="color:var(--text-3)">Theme</span>
          <div class="flex gap-1">
            <button class="pill" :class="{ active: ui.theme === 'light' }" @click="ui.setTheme('light')">
              <i data-lucide="sun" class="w-3.5 h-3.5 inline-block mr-1" style="vertical-align: -2px"></i>Light
            </button>
            <button class="pill" :class="{ active: ui.theme === 'dark' }" @click="ui.setTheme('dark')">
              <i data-lucide="moon" class="w-3.5 h-3.5 inline-block mr-1" style="vertical-align: -2px"></i>Dark
            </button>
          </div>
        </div>
        <div class="w-px h-6" style="background:var(--border)"></div>
        <div class="flex items-center gap-3">
          <span class="text-micro" style="color:var(--text-3)">Palette</span>
          <div class="flex gap-2">
            <div v-for="p in palettes" :key="p.id"
                 class="swatch" :class="{ active: ui.palette === p.id }"
                 :title="p.name"
                 :style="'background: linear-gradient(135deg, ' + p.colors[0] + ' 50%, ' + p.colors[1] + ' 50%)'"
                 @click="ui.setPalette(p.id)"></div>
          </div>
        </div>
      </div>
    </div>
  `
}
