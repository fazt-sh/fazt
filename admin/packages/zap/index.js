/**
 * zap
 * Framework-agnostic state & navigation library
 *
 * Usage:
 *   import { atom, map, createRouter, createCommands } from './packages/zap/index.js'
 *
 *   // Reactive state
 *   const count = atom(0)
 *   count.subscribe(v => console.log(v))
 *   count.set(1)
 *
 *   // Object state
 *   const user = map({ name: '', email: '' })
 *   user.setKey('name', 'Alice')
 *
 *   // Routing
 *   const router = createRouter({ mode: 'hash', routes })
 *   router.push('/apps')
 *
 *   // Commands (Cmd+K)
 *   const commands = createCommands()
 *   commands.register({ id: 'home', title: 'Dashboard', action: () => router.push('/') })
 */

// Atoms (reactive primitives)
export { atom, computed, effect, batch } from './atom.js'

// Maps (object state)
export { map, list } from './map.js'

// Router
export { createRouter } from './router.js'

// Commands
export { createCommands, getCommands } from './commands.js'

// Agent
export { createAgentContext, hasAgentContext, getAgentContext } from './agent.js'
