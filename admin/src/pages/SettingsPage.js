import { computed } from 'vue'
import { useUIStore } from '../stores/ui.js'
import { useHealthStore } from '../stores/health.js'
import { useIcons } from '../lib/useIcons.js'
import { palettes } from '../lib/palettes.js'

export default {
  name: 'SettingsPage',
  setup() {
    useIcons()
    const uiStore = useUIStore()
    const healthStore = useHealthStore()

    const shortcuts = [
      { key: 'Cmd + K', desc: 'Open command palette' },
      { key: 'G then H', desc: 'Go to Dashboard' },
      { key: 'G then A', desc: 'Go to Apps' },
      { key: 'G then S', desc: 'Go to System' },
      { key: 'Esc', desc: 'Close dialogs' }
    ]

    const currentTheme = computed(() => uiStore.theme)
    const currentPalette = computed(() => uiStore.palette)
    const currentPaletteName = computed(() => palettes.find(p => p.id === currentPalette.value)?.name || 'Stone')

    const setTheme = (themeId) => uiStore.setTheme(themeId)
    const setPalette = (paletteId) => uiStore.setPalette(paletteId)

    return { palettes, shortcuts, currentTheme, currentPalette, currentPaletteName, setTheme, setPalette, healthStore }
  },
  template: `
    <div class="design-system-page">
      <div class="content-container">
        <div class="content-scroll">

          <!-- Header -->
          <div class="mb-4 flex-shrink-0">
            <h1 class="text-title text-primary">Settings</h1>
            <p class="text-caption text-muted">Customize your admin experience</p>
          </div>

          <!-- Content -->
          <div class="max-w-xl">
            <!-- Theme -->
            <div class="card mb-4">
              <div class="card-header">
                <span class="text-heading text-primary">Theme</span>
              </div>
              <div class="card-body">
                <div class="flex gap-2">
                  <button class="pill" :class="{ active: currentTheme === 'light' }" @click="setTheme('light')">
                    <i data-lucide="sun" class="w-3.5 h-3.5"></i>
                    Light
                  </button>
                  <button class="pill" :class="{ active: currentTheme === 'dark' }" @click="setTheme('dark')">
                    <i data-lucide="moon" class="w-3.5 h-3.5"></i>
                    Dark
                  </button>
                </div>
              </div>
            </div>

            <!-- Palette -->
            <div class="card mb-4">
              <div class="card-header">
                <span class="text-heading text-primary">Color Palette</span>
              </div>
              <div class="card-body">
                <div class="flex gap-3">
                  <div v-for="p in palettes" :key="p.id"
                       class="swatch" :class="{ active: currentPalette === p.id }"
                       :title="p.name"
                       :style="'background: linear-gradient(135deg, ' + p.colors[0] + ' 50%, ' + p.colors[1] + ' 50%)'"
                       @click="setPalette(p.id)"></div>
                </div>
                <div class="text-caption text-muted mt-3">
                  Current: <span class="text-secondary">{{ currentPaletteName }}</span>
                </div>
              </div>
            </div>

            <!-- Keyboard Shortcuts -->
            <div class="card mb-4">
              <div class="card-header">
                <span class="text-heading text-primary">Keyboard Shortcuts</span>
              </div>
              <div class="card-body">
                <div class="space-y-2">
                  <div v-for="s in shortcuts" :key="s.key" class="flex items-center justify-between">
                    <span class="text-caption text-muted">{{ s.desc }}</span>
                    <kbd class="kbd">{{ s.key }}</kbd>
                  </div>
                </div>
              </div>
            </div>

            <!-- About -->
            <div class="card">
              <div class="card-header">
                <span class="text-heading text-primary">About</span>
              </div>
              <div class="card-body">
                <div class="flex items-center gap-3 mb-3">
                  <div class="w-10 h-10 flex items-center justify-center" style="background: var(--accent); border-radius: var(--radius-md)">
                    <i data-lucide="zap" class="w-5 h-5 text-white"></i>
                  </div>
                  <div>
                    <div class="text-heading text-primary">Fazt Admin</div>
                    <div class="text-caption text-muted">v{{ healthStore.data.version || '0.27.0' }}</div>
                  </div>
                </div>
                <p class="text-caption text-muted">
                  Sovereign compute. Single Go binary + SQLite database that runs anywhere.
                </p>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  `
}
