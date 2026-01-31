/**
 * zap/map
 * Object state with key-level updates
 */

/**
 * Create a reactive map (object state)
 * @template {Object} T
 * @param {T} initialValue
 * @returns {{ get: () => T, getKey: <K extends keyof T>(key: K) => T[K], set: (value: T) => void, setKey: <K extends keyof T>(key: K, value: T[K]) => void, subscribe: (fn: (value: T) => void) => () => void, subscribeKey: <K extends keyof T>(key: K, fn: (value: T[K]) => void) => () => void }}
 */
export function map(initialValue) {
  let value = { ...initialValue }
  const listeners = new Set()
  const keyListeners = new Map()

  function notifyAll() {
    listeners.forEach(fn => fn(value))
  }

  function notifyKey(key) {
    const keySet = keyListeners.get(key)
    if (keySet) {
      keySet.forEach(fn => fn(value[key]))
    }
  }

  return {
    /**
     * Get entire object
     */
    get() {
      return value
    },

    /**
     * Get specific key
     * @template {keyof T} K
     * @param {K} key
     * @returns {T[K]}
     */
    getKey(key) {
      return value[key]
    },

    /**
     * Set entire object
     * @param {T} newValue
     */
    set(newValue) {
      const oldValue = value
      value = { ...newValue }
      notifyAll()

      // Notify individual key listeners for changed keys
      for (const key of Object.keys(value)) {
        if (oldValue[key] !== value[key]) {
          notifyKey(key)
        }
      }
    },

    /**
     * Set specific key
     * @template {keyof T} K
     * @param {K} key
     * @param {T[K]} keyValue
     */
    setKey(key, keyValue) {
      if (value[key] !== keyValue) {
        value = { ...value, [key]: keyValue }
        notifyAll()
        notifyKey(key)
      }
    },

    /**
     * Subscribe to entire object changes
     * @param {(value: T) => void} fn
     * @returns {() => void}
     */
    subscribe(fn) {
      listeners.add(fn)
      fn(value)
      return () => listeners.delete(fn)
    },

    /**
     * Subscribe to specific key changes
     * @template {keyof T} K
     * @param {K} key
     * @param {(value: T[K]) => void} fn
     * @returns {() => void}
     */
    subscribeKey(key, fn) {
      if (!keyListeners.has(key)) {
        keyListeners.set(key, new Set())
      }
      const keySet = keyListeners.get(key)
      keySet.add(fn)
      fn(value[key])
      return () => keySet.delete(fn)
    },

    /**
     * Update multiple keys at once
     * @param {Partial<T>} updates
     */
    update(updates) {
      const oldValue = value
      value = { ...value, ...updates }
      notifyAll()

      for (const key of Object.keys(updates)) {
        if (oldValue[key] !== value[key]) {
          notifyKey(key)
        }
      }
    },

    /**
     * Reset to initial value
     */
    reset() {
      this.set(initialValue)
    }
  }
}

/**
 * Create a list (array state with helpers)
 * @template T
 * @param {T[]} initialValue
 */
export function list(initialValue = []) {
  let value = [...initialValue]
  const listeners = new Set()

  function notify() {
    listeners.forEach(fn => fn(value))
  }

  return {
    get() {
      return value
    },

    set(newValue) {
      value = [...newValue]
      notify()
    },

    subscribe(fn) {
      listeners.add(fn)
      fn(value)
      return () => listeners.delete(fn)
    },

    push(item) {
      value = [...value, item]
      notify()
    },

    pop() {
      const item = value[value.length - 1]
      value = value.slice(0, -1)
      notify()
      return item
    },

    shift() {
      const item = value[0]
      value = value.slice(1)
      notify()
      return item
    },

    unshift(item) {
      value = [item, ...value]
      notify()
    },

    remove(predicate) {
      value = value.filter((item, i) => !predicate(item, i))
      notify()
    },

    update(index, item) {
      if (index >= 0 && index < value.length) {
        value = [...value.slice(0, index), item, ...value.slice(index + 1)]
        notify()
      }
    },

    find(predicate) {
      return value.find(predicate)
    },

    get length() {
      return value.length
    }
  }
}
