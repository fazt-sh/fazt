/**
 * zap/agent
 * AI agent interface for automation
 *
 * Exposes window.__fazt_agent for AI agents to control the UI
 */

/**
 * Create agent context
 * @param {import('./router.js').createRouter} router
 * @param {import('./commands.js').createCommands} commands
 * @param {Object} [stores] - App stores
 */
export function createAgentContext(router, commands, stores = {}) {
  const context = {
    /**
     * Version
     */
    version: '0.1.0',

    /**
     * Navigation
     */
    navigation: {
      /**
       * Navigate to path
       * @param {string} path
       */
      goto: (path) => router.push(path),

      /**
       * Get current route
       */
      current: () => router.current.get(),

      /**
       * Get all routes
       */
      routes: () => router.routes.map(r => ({
        name: r.name,
        path: r.path
      }))
    },

    /**
     * Commands
     */
    commands: {
      /**
       * Execute command by ID
       * @param {string} id
       */
      execute: (id) => commands.execute(id),

      /**
       * List all commands
       */
      list: () => commands.getAll().map(c => ({
        id: c.id,
        title: c.title,
        description: c.description,
        shortcut: c.shortcut
      })),

      /**
       * Search commands
       * @param {string} query
       */
      search: (query) => {
        commands.query.set(query)
        return commands.filteredCommands.get().map(c => ({
          id: c.id,
          title: c.title
        }))
      },

      /**
       * Open command palette
       */
      open: () => commands.open(),

      /**
       * Close command palette
       */
      close: () => commands.close()
    },

    /**
     * State access
     */
    state: {
      /**
       * Get store value
       * @param {string} storeName
       */
      get: (storeName) => {
        const store = stores[storeName]
        return store ? store.get() : undefined
      },

      /**
       * Set store value
       * @param {string} storeName
       * @param {any} value
       */
      set: (storeName, value) => {
        const store = stores[storeName]
        if (store?.set) {
          store.set(value)
          return true
        }
        return false
      },

      /**
       * List available stores
       */
      list: () => Object.keys(stores)
    },

    /**
     * DOM helpers
     */
    dom: {
      /**
       * Click element by selector
       * @param {string} selector
       */
      click: (selector) => {
        const el = document.querySelector(selector)
        if (el) {
          el.click()
          return true
        }
        return false
      },

      /**
       * Fill input by selector
       * @param {string} selector
       * @param {string} value
       */
      fill: (selector, value) => {
        const el = document.querySelector(selector)
        if (el && 'value' in el) {
          el.value = value
          el.dispatchEvent(new Event('input', { bubbles: true }))
          return true
        }
        return false
      },

      /**
       * Get text content by selector
       * @param {string} selector
       */
      text: (selector) => {
        const el = document.querySelector(selector)
        return el?.textContent?.trim() || null
      },

      /**
       * Query all matching selectors
       * @param {string} selector
       */
      queryAll: (selector) => {
        return Array.from(document.querySelectorAll(selector)).map(el => ({
          tag: el.tagName.toLowerCase(),
          text: el.textContent?.trim().slice(0, 100),
          id: el.id || undefined,
          className: el.className || undefined
        }))
      },

      /**
       * Wait for element
       * @param {string} selector
       * @param {number} [timeout]
       */
      waitFor: (selector, timeout = 5000) => {
        return new Promise((resolve, reject) => {
          const el = document.querySelector(selector)
          if (el) {
            resolve(true)
            return
          }

          const observer = new MutationObserver(() => {
            const el = document.querySelector(selector)
            if (el) {
              observer.disconnect()
              resolve(true)
            }
          })

          observer.observe(document.body, {
            childList: true,
            subtree: true
          })

          setTimeout(() => {
            observer.disconnect()
            resolve(false)
          }, timeout)
        })
      }
    },

    /**
     * Utilities
     */
    utils: {
      /**
       * Wait for milliseconds
       * @param {number} ms
       */
      sleep: (ms) => new Promise(resolve => setTimeout(resolve, ms)),

      /**
       * Get page title
       */
      title: () => document.title,

      /**
       * Get current URL
       */
      url: () => window.location.href,

      /**
       * Log message
       * @param {string} message
       */
      log: (message) => console.log('[Agent]', message)
    }
  }

  // Expose to window
  if (typeof window !== 'undefined') {
    window.__fazt_agent = context
  }

  return context
}

/**
 * Check if agent context is available
 */
export function hasAgentContext() {
  return typeof window !== 'undefined' && window.__fazt_agent !== undefined
}

/**
 * Get agent context
 */
export function getAgentContext() {
  return typeof window !== 'undefined' ? window.__fazt_agent : null
}
