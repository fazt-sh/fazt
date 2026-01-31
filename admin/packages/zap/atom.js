/**
 * zap/atom
 * Reactive atoms for state management
 */

/**
 * Create a reactive atom
 * @template T
 * @param {T} initialValue
 * @returns {{ get: () => T, set: (value: T | ((prev: T) => T)) => void, subscribe: (fn: (value: T) => void) => () => void }}
 */
export function atom(initialValue) {
  let value = initialValue
  const listeners = new Set()

  return {
    /**
     * Get current value
     */
    get() {
      return value
    },

    /**
     * Set value (accepts value or updater function)
     * @param {T | ((prev: T) => T)} newValue
     */
    set(newValue) {
      const nextValue = typeof newValue === 'function'
        ? /** @type {Function} */ (newValue)(value)
        : newValue

      if (nextValue !== value) {
        value = nextValue
        listeners.forEach(fn => fn(value))
      }
    },

    /**
     * Subscribe to value changes
     * @param {(value: T) => void} fn
     * @returns {() => void} Unsubscribe function
     */
    subscribe(fn) {
      listeners.add(fn)
      // Immediately call with current value
      fn(value)
      return () => listeners.delete(fn)
    },

    /**
     * Get listener count (for debugging)
     */
    get listenerCount() {
      return listeners.size
    }
  }
}

/**
 * Create a computed atom that derives from other atoms
 * @template T
 * @param {() => T} compute
 * @param {Array<ReturnType<typeof atom>>} deps
 * @returns {{ get: () => T, subscribe: (fn: (value: T) => void) => () => void }}
 */
export function computed(compute, deps) {
  let value = compute()
  const listeners = new Set()

  // Subscribe to all dependencies
  deps.forEach(dep => {
    dep.subscribe(() => {
      const newValue = compute()
      if (newValue !== value) {
        value = newValue
        listeners.forEach(fn => fn(value))
      }
    })
  })

  return {
    get() {
      return value
    },

    subscribe(fn) {
      listeners.add(fn)
      fn(value)
      return () => listeners.delete(fn)
    }
  }
}

/**
 * Create an effect that runs when atoms change
 * @param {() => void | (() => void)} effect
 * @param {Array<ReturnType<typeof atom>>} deps
 * @returns {() => void} Cleanup function
 */
export function effect(effectFn, deps) {
  let cleanup

  const run = () => {
    if (cleanup) cleanup()
    cleanup = effectFn()
  }

  // Initial run
  run()

  // Subscribe to deps
  const unsubscribes = deps.map(dep =>
    dep.subscribe(() => run())
  )

  // Return cleanup function
  return () => {
    if (cleanup) cleanup()
    unsubscribes.forEach(unsub => unsub())
  }
}

/**
 * Batch multiple atom updates
 * @param {() => void} fn
 */
let batchDepth = 0
const pendingUpdates = new Set()

export function batch(fn) {
  batchDepth++
  try {
    fn()
  } finally {
    batchDepth--
    if (batchDepth === 0) {
      pendingUpdates.forEach(update => update())
      pendingUpdates.clear()
    }
  }
}
