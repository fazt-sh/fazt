// Settings Page Template
import { defineComponent, ref, computed } from 'vue'
import { loadSettings, updateSetting } from '../lib/settings.js'
import { getSessionUrl, generateNewSession } from '../lib/session.js'

export const SettingsPage = defineComponent({
  name: 'SettingsPage',

  setup() {
    const settings = ref(loadSettings())

    const themes = [
      { value: 'system', label: 'System' },
      { value: 'light', label: 'Light' },
      { value: 'dark', label: 'Dark' }
    ]

    function setTheme(theme) {
      settings.value = updateSetting('theme', theme)
    }

    function copySessionUrl() {
      const url = getSessionUrl()
      navigator.clipboard.writeText(url)
      // Could add toast notification here
    }

    function resetSession() {
      if (confirm('Start a new session? This will clear your current data view.')) {
        generateNewSession()
        window.location.reload()
      }
    }

    return {
      settings,
      themes,
      setTheme,
      copySessionUrl,
      resetSession
    }
  },

  template: `
    <div class="h-full flex flex-col">
      <!-- Header -->
      <header class="flex-none bg-white dark:bg-neutral-900 border-b border-neutral-200 dark:border-neutral-800 px-4 py-4 safe-top">
        <div class="max-w-4xl mx-auto flex items-center">
          <router-link
            to="/"
            class="p-2 -ml-2 rounded-lg hover:bg-neutral-100 dark:hover:bg-neutral-800"
          >
            <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 19l-7-7 7-7"></path>
            </svg>
          </router-link>
          <h1 class="text-xl font-semibold ml-2">Settings</h1>
        </div>
      </header>

      <!-- Content -->
      <main class="flex-1 overflow-y-auto">
        <div class="max-w-4xl mx-auto px-4 py-6">
          <!-- Theme Section -->
          <section class="mb-8">
            <h2 class="text-sm font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wide mb-3">
              Appearance
            </h2>
            <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden">
              <div class="p-4">
                <label class="block text-sm font-medium mb-2">Theme</label>
                <div class="flex gap-2">
                  <button
                    v-for="theme in themes"
                    :key="theme.value"
                    @click="setTheme(theme.value)"
                    :class="[
                      'flex-1 py-2 px-4 rounded-lg text-sm font-medium transition-colors',
                      settings.theme === theme.value
                        ? 'bg-blue-500 text-white'
                        : 'bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700'
                    ]"
                  >
                    {{ theme.label }}
                  </button>
                </div>
              </div>
            </div>
          </section>

          <!-- Session Section -->
          <section class="mb-8">
            <h2 class="text-sm font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wide mb-3">
              Session
            </h2>
            <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden">
              <div class="p-4 border-b border-neutral-200 dark:border-neutral-800">
                <label class="block text-sm font-medium mb-2">Share Session</label>
                <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-3">
                  Copy this URL to share your session with others
                </p>
                <button
                  @click="copySessionUrl"
                  class="w-full py-2 px-4 bg-neutral-100 dark:bg-neutral-800 hover:bg-neutral-200 dark:hover:bg-neutral-700 rounded-lg text-sm font-medium transition-colors"
                >
                  Copy Session URL
                </button>
              </div>
              <div class="p-4">
                <label class="block text-sm font-medium mb-2">New Session</label>
                <p class="text-sm text-neutral-500 dark:text-neutral-400 mb-3">
                  Start fresh with a new session
                </p>
                <button
                  @click="resetSession"
                  class="w-full py-2 px-4 bg-red-50 dark:bg-red-900/20 hover:bg-red-100 dark:hover:bg-red-900/30 text-red-600 dark:text-red-400 rounded-lg text-sm font-medium transition-colors"
                >
                  Reset Session
                </button>
              </div>
            </div>
          </section>

          <!-- About Section -->
          <section>
            <h2 class="text-sm font-medium text-neutral-500 dark:text-neutral-400 uppercase tracking-wide mb-3">
              About
            </h2>
            <div class="bg-white dark:bg-neutral-900 border border-neutral-200 dark:border-neutral-800 rounded-lg overflow-hidden">
              <div class="p-4">
                <p class="text-sm text-neutral-500 dark:text-neutral-400">
                  Built with Fazt - sovereign compute for individuals
                </p>
              </div>
            </div>
          </section>
        </div>
      </main>
    </div>
  `
})
