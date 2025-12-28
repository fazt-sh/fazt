# Harness Apps

## Summary

A Harness is a specialized Fazt app designed to host autonomous AI agents.
It can observe the system, execute code, and evolve its own logic—all within
the safety of the Fazt sandbox.

## Concept

Traditional AI agents are ephemeral. A Harness gives them:
- **Persistent Memory**: Via `fazt.storage`
- **Self-Modification**: Via `fazt.git`
- **System Access**: Via `fazt.kernel`
- **Scheduled Wakeup**: Via JS-Cron

## The Harness-as-App Model

```
my-harness/
├── app.json
├── api/
│   ├── main.js         # Main agent loop
│   ├── tools.js        # Custom tools
│   └── memory.js       # Memory management
```

```json
{
  "name": "architect-agent",
  "permissions": [
    "ai:complete",
    "storage:ds",
    "kernel:deploy",
    "kernel:status"
  ],
  "cron": [
    { "schedule": "*/15 * * * *", "handler": "api/main.js" }
  ]
}
```

## Native SDK

### fazt.git

Version control for VFS:

```javascript
// Commit current state
await fazt.git.commit('Added new feature');

// View history
const history = await fazt.git.log(10);

// Diff against previous
const changes = await fazt.git.diff();

// Rollback to commit
await fazt.git.rollback('abc123');
```

### fazt.kernel

System-level operations:

```javascript
// Deploy a new app
await fazt.kernel.deploy('new-app', [
    { path: 'index.html', content: '<h1>Hello</h1>' },
    { path: 'api/main.js', content: 'module.exports = ...' }
]);

// Get system status
const status = await fazt.kernel.status();

// List all apps
const apps = await fazt.kernel.apps.list();
```

## Self-Evolution

A harness can modify its own code:

```javascript
// api/main.js
module.exports = async function(request) {
    const task = await getNextTask();

    // Agent writes new capability
    const newCode = await fazt.ai.complete(
        `Write a JS function that ${task.description}`
    );

    // Save to own VFS
    await fazt.fs.write('/api/tools/' + task.name + '.js', newCode);

    // Commit the change
    await fazt.git.commit(`Added tool: ${task.name}`);
};
```

## Permission Model

### Graceful Degradation

Harnesses request elevated permissions in `app.json`. If denied:
- Can still operate internally
- Cannot affect other apps
- Cannot modify system routing

### Blast Radius

Failures are contained:
- Harness can only modify its own VFS
- Cannot access other apps' storage
- Kernel operations are audit-logged

## Dual Interface

### Headless Mode

Agent runs via cron, processes tasks silently.

### Visual Cockpit

Dashboard shows:
- Token usage
- Tool call history
- "Thought" streams
- Active deployments

## Example: Code Review Agent

```javascript
// api/main.js
module.exports = async function(request) {
    // Get recent deploys
    const deploys = await fazt.kernel.deploys.recent(10);

    for (const deploy of deploys) {
        // Get code
        const files = await fazt.kernel.apps.files(deploy.uuid);

        // Review with AI
        const review = await fazt.ai.complete(
            `Review this code for security issues:\n${files}`
        );

        // Store review
        await fazt.storage.ds.insert('reviews', {
            app: deploy.uuid,
            review: review.text,
            timestamp: Date.now()
        });
    }
};
```
