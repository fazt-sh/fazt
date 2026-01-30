# fazt auth - Authentication Management

Configure OAuth providers and manage users.

## Provider Configuration

### fazt auth providers

List configured OAuth providers.

```bash
fazt @<peer> auth providers
```

Output:
```
PROVIDER     STATUS     CLIENT ID
-----------------------------------------------------------------
google       enabled    125525550164-dhc2...
github       disabled   -
discord      disabled   -
microsoft    disabled   -
```

### fazt auth provider

Configure an OAuth provider.

```bash
fazt @<peer> auth provider <name> \
  --client-id <id> \
  --client-secret <secret>
```

**Enable/disable:**
```bash
fazt @<peer> auth provider <name> --enable
fazt @<peer> auth provider <name> --disable
```

**Supported providers:**
- `google` - Google OAuth 2.0
- `github` - GitHub OAuth
- `discord` - Discord OAuth
- `microsoft` - Microsoft (personal accounts)

**Example - Google OAuth:**
```bash
fazt @<peer> auth provider google \
  --client-id "123456789.apps.googleusercontent.com" \
  --client-secret "GOCSPX-xxxxx" \
  --enable
```

## User Management

### fazt auth users

List all users.

```bash
fazt @<peer> auth users
```

### fazt auth user

Show or modify a user.

```bash
# Show user details
fazt @<peer> auth user <id>

# Change role
fazt @<peer> auth user <id> --role admin

# Delete user
fazt @<peer> auth user <id> --delete
```

**Roles:**
- `owner` - Full access, can manage other users
- `admin` - Administrative access
- `user` - Regular user (default)

## Invites

Create invite codes for new users.

### fazt auth invite

```bash
fazt @<peer> auth invite --role <role>
fazt @<peer> auth invite --role admin --expiry 7  # 7 day expiry
```

### fazt auth invites

List all invite codes.

```bash
fazt @<peer> auth invites
```

## OAuth Setup Checklist

1. **Create OAuth credentials** in provider console (e.g., Google Cloud Console)

2. **Configure callback URL** in provider:
   ```
   https://your-domain.com/auth/callback/google
   ```

3. **Add provider to fazt:**
   ```bash
   fazt @<peer> auth provider google \
     --client-id <id> \
     --client-secret <secret> \
     --enable
   ```

4. **Verify:**
   ```bash
   fazt @<peer> auth providers
   ```

See `patterns/google-oauth.md` for detailed Google OAuth setup guide.

## Important Notes

- OAuth callback URLs require HTTPS (production domain)
- Local development cannot use OAuth directly (no valid callback URL)
- See `references/auth-integration.md` for local testing strategies
