/**
 * Settings Page
 */

import { theme, palette, setTheme, setPalette } from '../stores/app.js'

const palettes = [
  { id: 'stone', name: 'Stone', colors: ['#faf9f7', '#d97706'] },
  { id: 'slate', name: 'Slate', colors: ['#f8fafc', '#0284c7'] },
  { id: 'oxide', name: 'Oxide', colors: ['#faf8f8', '#e11d48'] },
  { id: 'forest', name: 'Forest', colors: ['#f7faf8', '#059669'] },
  { id: 'violet', name: 'Violet', colors: ['#faf9fc', '#7c3aed'] }
]

/**
 * Render settings page
 * @param {HTMLElement} container
 * @param {Object} ctx
 */
export function render(container, ctx) {
  function update() {
    const currentTheme = theme.get()
    const currentPalette = palette.get()

    container.innerHTML = `
      <div class="flex flex-col h-full overflow-hidden">
        <!-- Header -->
        <div class="mb-4 flex-shrink-0">
          <h1 class="text-title text-primary">Settings</h1>
          <p class="text-caption text-muted">Customize your admin experience</p>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-auto scroll-panel">
          <div class="max-w-xl">
            <!-- Theme -->
            <div class="card mb-4">
              <div class="card-header">
                <span class="text-heading text-primary">Theme</span>
              </div>
              <div class="card-body">
                <div class="flex gap-2">
                  <button class="pill ${currentTheme === 'light' ? 'active' : ''}" data-theme="light">
                    <i data-lucide="sun" class="w-3.5 h-3.5"></i>
                    Light
                  </button>
                  <button class="pill ${currentTheme === 'dark' ? 'active' : ''}" data-theme="dark">
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
                  ${palettes.map(p => `
                    <button class="swatch ${currentPalette === p.id ? 'active' : ''}" data-palette="${p.id}" title="${p.name}" style="background: linear-gradient(135deg, ${p.colors[0]} 50%, ${p.colors[1]} 50%)"></button>
                  `).join('')}
                </div>
                <div class="text-caption text-muted mt-3">
                  Current: <span class="text-secondary">${palettes.find(p => p.id === currentPalette)?.name || 'Stone'}</span>
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
                  ${[
                    { key: 'Cmd + K', desc: 'Open command palette' },
                    { key: 'G then H', desc: 'Go to Dashboard' },
                    { key: 'G then A', desc: 'Go to Apps' },
                    { key: 'G then S', desc: 'Go to System' },
                    { key: 'Esc', desc: 'Close dialogs' }
                  ].map(shortcut => `
                    <div class="flex items-center justify-between">
                      <span class="text-caption text-muted">${shortcut.desc}</span>
                      <kbd class="kbd">${shortcut.key}</kbd>
                    </div>
                  `).join('')}
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
                    <div class="text-caption text-muted">v0.17.0</div>
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
    `

    // Re-render Lucide icons
    if (window.lucide) {
      window.lucide.createIcons()
    }

    // Theme handlers
    container.querySelectorAll('[data-theme]').forEach(btn => {
      btn.addEventListener('click', () => {
        setTheme(btn.dataset.theme)
      })
    })

    // Palette handlers
    container.querySelectorAll('[data-palette]').forEach(btn => {
      btn.addEventListener('click', () => {
        setPalette(btn.dataset.palette)
      })
    })
  }

  // Subscribe to theme changes
  const unsubTheme = theme.subscribe(update)
  const unsubPalette = palette.subscribe(update)

  // Initial render
  update()

  // Return cleanup function
  return () => {
    unsubTheme()
    unsubPalette()
  }
}
