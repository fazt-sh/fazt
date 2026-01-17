# /fazt-app - Build Fazt Apps

Build and deploy apps to fazt instances with Claude.

## Usage

```
/fazt-app <description>
/fazt-app "build a pomodoro tracker with task persistence"
```

## Context

You are building an app for fazt.sh - a single-binary personal cloud.

### App Structure

```
my-app/
├── manifest.json      # Required: {"name": "my-app"}
├── index.html         # Entry point
├── main.js            # ES6 module entry
├── components/        # Vue components (plain JS, not SFC)
└── api/               # Serverless functions
    └── data.js        # → GET/POST /api/data
```

### Frontend Stack (Zero Build)

```html
<script type="importmap">
{
  "imports": {
    "vue": "https://unpkg.com/vue@3/dist/vue.esm-browser.js"
  }
}
</script>
<script src="https://cdn.tailwindcss.com"></script>
<script type="module" src="main.js"></script>
```

### Vue Components (Plain JS, Not SFC)

```javascript
// components/Timer.js
import { ref, computed } from 'vue'

export default {
  props: ['duration'],
  setup(props) {
    const remaining = ref(props.duration)
    const formatted = computed(() => {
      const m = Math.floor(remaining.value / 60)
      const s = remaining.value % 60
      return `${m}:${s.toString().padStart(2, '0')}`
    })
    return { remaining, formatted }
  },
  template: `
    <div class="text-4xl font-mono">{{ formatted }}</div>
  `
}
```

### Serverless Functions

```javascript
// api/tasks.js
function handler(req) {
  const db = fazt.storage.kv

  if (req.method === 'GET') {
    const tasks = db.get('tasks') || []
    return { status: 200, body: JSON.stringify(tasks) }
  }

  if (req.method === 'POST') {
    const tasks = db.get('tasks') || []
    const newTask = JSON.parse(req.body)
    tasks.push({ ...newTask, id: Date.now() })
    db.set('tasks', tasks)
    return { status: 201, body: JSON.stringify(newTask) }
  }

  return { status: 405, body: 'Method not allowed' }
}
```

### User Identification Pattern

For apps that need user-specific data without auth:

```javascript
// Get or create user ID from query string
function getUserId() {
  const params = new URLSearchParams(location.search)
  let id = params.get('u')
  if (!id) {
    id = crypto.randomUUID().split('-')[0]
    history.replaceState(null, '', `?u=${id}`)
  }
  return id
}

// Use in API calls
fetch(`/api/tasks?user=${getUserId()}`)
```

## Location Behavior

Where to create app files depends on context:

| Scenario | Location |
|----------|----------|
| In fazt repo, no flag | `servers/zyt/{app}/` |
| Not in fazt repo, no flag | `/tmp/fazt-{app}-{hash}/` |
| `--in <dir>` specified | `<dir>/{app}/` |
| `--tmp` flag | `/tmp/fazt-{app}-{hash}/` |

### Detection Logic

```
1. Check if --in or --tmp flag provided → use that
2. Check if cwd is fazt repo (has CLAUDE.md with "fazt")
   → Yes: use servers/zyt/{app}/
   → No: use /tmp/fazt-{app}-{hash}/
```

### Examples

```bash
# In fazt repo (persisted, can commit to git)
/fazt-app "pomodoro tracker"
→ Creates: servers/zyt/pomodoro/

# Explicit directory
/fazt-app "game" --in ~/projects/games
→ Creates: ~/projects/games/game/

# Force temp (throwaway)
/fazt-app "quick test" --tmp
→ Creates: /tmp/fazt-quick-test-a1b2c3/
```

## Workflow

1. Determine location (see Location Behavior above)
2. Create app folder with manifest.json
3. Write all files (html, js, api)
4. Deploy: `fazt app deploy <folder> --to local`
5. Report URL: `http://{name}.192.168.64.3:8080`
6. For edits: modify files in same folder, redeploy

## Deployment

```bash
# To local instance
fazt app deploy servers/zyt/myapp/ --to local

# To production
fazt app deploy servers/zyt/myapp/ --to zyt
```

## Design Guidelines

- Modern, clean UI with Tailwind
- Dark mode support (respect prefers-color-scheme)
- Mobile-first responsive design
- Minimal dependencies (prefer native APIs)
- No build step required

## CLI Reference

| Command | Purpose | Example |
|---------|---------|---------|
| `fazt app list [peer]` | See deployed apps | `fazt app list zyt` |
| `fazt app deploy <dir> --to <peer>` | Push local folder | `fazt app deploy ./myapp --to zyt` |
| `fazt app install <url> --to <peer>` | Install from GitHub | `fazt app install github:user/repo --to zyt` |
| `fazt app upgrade <app>` | Update git-sourced app | `fazt app upgrade myapp` |
| `fazt app pull <app> --to <dir>` | Download app locally | `fazt app pull myapp --to ./local` |
| `fazt app info <app>` | View app details | `fazt app info myapp` |
| `fazt app remove <app> --from <peer>` | Delete an app | `fazt app remove myapp --from zyt` |

### Typical Workflows

**Build & Deploy:**
```bash
/fazt-app "todo list"        # Claude creates app
fazt app deploy ./todo --to local   # Test locally
fazt app deploy ./todo --to zyt     # Push to production
```

**Fork & Modify:**
```bash
fazt app pull existingapp --to ./local   # Download
# ... edit files ...
fazt app deploy ./local --to zyt         # Redeploy
```

**Install from GitHub:**
```bash
fazt app install github:someone/cool-app --to zyt
fazt app upgrade cool-app   # Later, check for updates
```

## Instructions

When the user invokes `/fazt-app`:

1. Parse the description to understand what app to build
2. Determine the target location based on the rules above
3. Create the app with:
   - manifest.json with app name
   - index.html with proper imports
   - main.js with Vue app setup
   - Any needed components in components/
   - API endpoints in api/ if persistence needed
4. Deploy to local peer if running
5. Report the URL where the app can be accessed
