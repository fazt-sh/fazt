# Google OAuth Setup Guide

Complete step-by-step guide for setting up Google OAuth on a fazt instance.

**Before using this guide**, check if OAuth is already configured:

```bash
fazt @<peer> auth providers
```

If you see `google enabled`, OAuth is already set up - skip this guide and
proceed with building your app. This guide is only needed for first-time setup.

## 1. Create Google Cloud Project

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Navigate to **APIs & Services > OAuth consent screen**
4. Choose "External" user type
5. Fill in:
   - App name
   - User support email
   - Developer contact email
6. Add scopes: `email`, `profile`, `openid`
7. Add test users (if in testing mode)
8. Click **Save and Continue** through remaining steps

## 2. Create OAuth Credentials

1. Go to **APIs & Services > Credentials**
2. Click **Create Credentials > OAuth client ID**
3. Application type: **Web application**
4. Name: Your app name (e.g., "Fazt Auth")
5. **Authorized redirect URIs** - Add:
   ```
   https://yourdomain.com/auth/callback/google
   ```

   **IMPORTANT**:
   - Use your actual root domain
   - Callbacks must be on root domain, not subdomains
   - All subdomains share the same callback

6. Click **Create**
7. Copy the **Client ID** and **Client Secret**

## 3. Configure Fazt Instance

```bash
# Configure the provider with credentials
fazt @<peer> auth provider google \
  --client-id YOUR_CLIENT_ID.apps.googleusercontent.com \
  --client-secret YOUR_CLIENT_SECRET \
  --enable

# Or if you have local access to the database:
fazt auth provider google \
  --client-id YOUR_CLIENT_ID.apps.googleusercontent.com \
  --client-secret YOUR_CLIENT_SECRET \
  --db /path/to/data.db

fazt auth provider google --enable --db /path/to/data.db

# Verify configuration
fazt @<peer> auth providers
```

## 4. Auth Endpoints (Automatic)

Once configured, these endpoints work on ALL apps automatically:

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/auth/session` | GET | Get current user session |
| `/auth/login/google` | GET | Initiate Google login |
| `/auth/login/google?redirect=URL` | GET | Login with redirect |
| `/auth/logout` | POST | Sign out |
| `/auth/callback/google` | GET | OAuth callback (internal) |

## 5. Test the Flow

1. Deploy any app to your fazt instance
2. Navigate to `https://your-app.yourdomain.com/auth/login/google`
3. Should redirect to Google sign-in
4. After sign-in, redirects back to your app
5. Check `/auth/session` - should return user data

## Common Issues

### redirect_uri_mismatch

**Cause**: Callback URL doesn't match Google Console settings.

**Fix**:
- Add exact URL to Google Console Authorized redirect URIs
- Remember: `https://yourdomain.com/auth/callback/google` (root domain only)

### User type: External (Testing)

**Issue**: Only test users can sign in.

**Fix**:
- Add user emails to test users list in OAuth consent screen, OR
- Publish the app (submit for verification)

### Cookies not persisting across subdomains

**Issue**: Session lost when navigating between subdomains.

**Cause**: Cookie domain not set correctly.

**Fix**: Fazt sets cookies on root domain (`.yourdomain.com`) automatically. Ensure you're using proper subdomain format.

## Integration Patterns

### Server-Side (api/main.js)

```javascript
// Check if user is logged in
var user = fazt.auth.getUser();
if (!user) {
  return {
    status: 401,
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ error: 'Sign in required' })
  };
}

// User object contains:
// - id: unique user ID
// - email: user's email
// - name: display name
// - picture: profile picture URL
// - role: 'user' for OAuth users, 'owner' for instance owner
// - provider: 'google'
```

### Client-Side (index.html)

```javascript
let currentUser = null;

// Check session on page load
async function checkAuth() {
  try {
    const res = await fetch('/auth/session');
    const json = await res.json();
    // Note: API wraps response in 'data' field
    const data = json.data || json;

    if (data.user) {
      currentUser = data.user;
      renderLoggedInUI();
    } else {
      renderLoggedOutUI();
    }
  } catch (e) {
    console.error('Auth check failed:', e);
  }
}

// Trigger sign-in with redirect back to current page
function signIn() {
  const redirect = encodeURIComponent(window.location.href);
  window.location.href = '/auth/login/google?redirect=' + redirect;
}

// Sign out
async function signOut() {
  await fetch('/auth/logout', { method: 'POST' });
  window.location.reload();
}

// Initialize
checkAuth();
```

### Avatar Button + Dropdown Pattern

```html
<style>
.avatar-btn {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  border: 1px solid #333;
  background: #1a1a1a;
  cursor: pointer;
  overflow: hidden;
  padding: 0;
}
.avatar-btn img { width: 100%; height: 100%; object-fit: cover; }
.auth-dropdown {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background: #141414;
  border: 1px solid #2a2a2a;
  border-radius: 12px;
  min-width: 200px;
  display: none;
  z-index: 100;
}
.auth-dropdown.show { display: block; }
</style>

<div id="authContainer" style="position: relative;">
  <!-- Will be populated by renderAuthUI() -->
</div>

<script>
function renderAuthUI() {
  const container = document.getElementById('authContainer');

  if (currentUser) {
    container.innerHTML = `
      <button class="avatar-btn" onclick="toggleDropdown(event)">
        <img src="${currentUser.picture}" alt="${currentUser.name}">
      </button>
      <div id="authDropdown" class="auth-dropdown">
        <div style="padding: 16px; border-bottom: 1px solid #2a2a2a;">
          <div style="font-weight: 500;">${currentUser.name}</div>
          <div style="font-size: 12px; color: #888;">${currentUser.email}</div>
        </div>
        <button onclick="signOut()" style="width: 100%; padding: 12px; text-align: left; background: none; border: none; color: inherit; cursor: pointer;">
          Sign Out
        </button>
      </div>
    `;
  } else {
    container.innerHTML = `
      <button class="avatar-btn" onclick="signIn()">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/>
          <circle cx="12" cy="7" r="4"/>
        </svg>
      </button>
    `;
  }
}

function toggleDropdown(e) {
  e.stopPropagation();
  document.getElementById('authDropdown')?.classList.toggle('show');
}

// Close dropdown when clicking outside
document.addEventListener('click', (e) => {
  if (!e.target.closest('#authContainer')) {
    document.getElementById('authDropdown')?.classList.remove('show');
  }
});
</script>
```

## Other Providers

Fazt supports multiple OAuth providers:

```bash
# GitHub
fazt @<peer> auth provider github --client-id XXX --client-secret YYY --enable

# Discord
fazt @<peer> auth provider discord --client-id XXX --client-secret YYY --enable

# Microsoft
fazt @<peer> auth provider microsoft --client-id XXX --client-secret YYY --enable
```

Callback URLs follow the same pattern:
- `https://yourdomain.com/auth/callback/github`
- `https://yourdomain.com/auth/callback/discord`
- `https://yourdomain.com/auth/callback/microsoft`
