/**
 * zap/commands
 * Command palette (Cmd+K) system
 */

import { atom, computed } from './atom.js'
import { list } from './map.js'

/**
 * @typedef {Object} Command
 * @property {string} id - Unique command ID
 * @property {string} title - Display title
 * @property {string} [description] - Optional description
 * @property {string} [shortcut] - Keyboard shortcut display
 * @property {string} [icon] - Lucide icon name
 * @property {string} [group] - Group name
 * @property {string[]} [keywords] - Search keywords
 * @property {Function} action - Command action
 * @property {boolean} [hidden] - Hidden from palette
 */

/**
 * Create command system
 * @param {Object} [options]
 * @param {string} [options.placeholder] - Search placeholder
 */
export function createCommands(options = {}) {
  const { placeholder = 'Type a command or search...' } = options

  /** @type {Map<string, Command>} */
  const commandsMap = new Map()

  // State
  const isOpen = atom(false)
  const query = atom('')
  const selectedIndex = atom(0)
  const recentCommands = list([])

  /**
   * Get all registered commands
   */
  function getAll() {
    return Array.from(commandsMap.values())
  }

  /**
   * Filter commands by query
   * @param {string} q
   */
  function filter(q) {
    const commands = getAll().filter(cmd => !cmd.hidden)
    if (!q) return commands

    const lower = q.toLowerCase()
    return commands.filter(cmd => {
      const searchText = [
        cmd.title,
        cmd.description || '',
        ...(cmd.keywords || [])
      ].join(' ').toLowerCase()
      return searchText.includes(lower)
    })
  }

  // Computed filtered commands
  const filteredCommands = computed(
    () => filter(query.get()),
    [query]
  )

  /**
   * Register a command
   * @param {Command} command
   */
  function register(command) {
    commandsMap.set(command.id, command)
  }

  /**
   * Register multiple commands
   * @param {Command[]} commands
   */
  function registerAll(commands) {
    commands.forEach(register)
  }

  /**
   * Unregister a command
   * @param {string} id
   */
  function unregister(id) {
    commandsMap.delete(id)
  }

  /**
   * Execute a command
   * @param {string} id
   */
  function execute(id) {
    const command = commandsMap.get(id)
    if (command) {
      // Add to recent (max 5)
      const recent = recentCommands.get().filter(r => r !== id)
      recentCommands.set([id, ...recent].slice(0, 5))

      // Execute action
      command.action()

      // Close palette
      close()
    }
  }

  /**
   * Execute selected command
   */
  function executeSelected() {
    const commands = filteredCommands.get()
    const index = selectedIndex.get()
    if (commands[index]) {
      execute(commands[index].id)
    }
  }

  /**
   * Open palette
   */
  function open() {
    isOpen.set(true)
    query.set('')
    selectedIndex.set(0)
  }

  /**
   * Close palette
   */
  function close() {
    isOpen.set(false)
    query.set('')
    selectedIndex.set(0)
  }

  /**
   * Toggle palette
   */
  function toggle() {
    if (isOpen.get()) {
      close()
    } else {
      open()
    }
  }

  /**
   * Move selection
   * @param {number} delta
   */
  function moveSelection(delta) {
    const commands = filteredCommands.get()
    const current = selectedIndex.get()
    const next = Math.max(0, Math.min(commands.length - 1, current + delta))
    selectedIndex.set(next)
  }

  /**
   * Handle keyboard navigation
   * @param {KeyboardEvent} e
   */
  function handleKeyDown(e) {
    if (!isOpen.get()) return

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault()
        moveSelection(1)
        break
      case 'ArrowUp':
        e.preventDefault()
        moveSelection(-1)
        break
      case 'Enter':
        e.preventDefault()
        executeSelected()
        break
      case 'Escape':
        e.preventDefault()
        close()
        break
    }
  }

  // Global keyboard shortcut (Cmd/Ctrl + K)
  function handleGlobalKeyDown(e) {
    if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
      e.preventDefault()
      toggle()
    }
  }

  // Set up global listener
  if (typeof window !== 'undefined') {
    window.addEventListener('keydown', handleGlobalKeyDown)
  }

  return {
    // State (reactive)
    isOpen,
    query,
    selectedIndex,
    filteredCommands,
    recentCommands,

    // Config
    placeholder,

    // Methods
    register,
    registerAll,
    unregister,
    execute,
    executeSelected,
    open,
    close,
    toggle,
    moveSelection,
    handleKeyDown,

    // Get command by ID
    get(id) {
      return commandsMap.get(id)
    },

    // Get all commands
    getAll,

    // Cleanup
    destroy() {
      if (typeof window !== 'undefined') {
        window.removeEventListener('keydown', handleGlobalKeyDown)
      }
    }
  }
}

/**
 * Default global command instance
 */
let globalCommands = null

/**
 * Get or create global commands
 * @param {Object} [options]
 */
export function getCommands(options) {
  if (!globalCommands || options) {
    globalCommands = createCommands(options)
  }
  return globalCommands
}
