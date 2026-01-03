# Forms Service

## Summary

A dumb bucket for form submissions. You design the HTML, the service
catches and stores whatever is POSTed. Zero configuration required.

## How It Works

```
┌──────────────┐      POST       ┌──────────────┐
│  Your HTML   │ ──────────────► │ /_services/forms/ │
│    Form      │                 │   {name}     │
└──────────────┘                 └──────┬───────┘
                                        │
                                        ▼
                                 ┌──────────────┐
                                 │   Storage    │
                                 │  (JSON blob) │
                                 └──────────────┘
```

## Usage

### 1. Create Your Form (Any HTML)

```html
<form action="/_services/forms/contact" method="POST">
  <label>Email</label>
  <input name="email" type="email" required>

  <label>Message</label>
  <textarea name="message"></textarea>

  <label>Priority</label>
  <select name="priority">
    <option value="low">Low</option>
    <option value="high">High</option>
  </select>

  <button type="submit">Send</button>
</form>
```

### 2. Service Stores Everything

Whatever fields you include get stored as JSON:

```json
{
  "email": "user@example.com",
  "message": "Hello there",
  "priority": "high",
  "_submitted": "2024-01-15T10:30:00Z",
  "_ip": "192.168.1.1"
}
```

### 3. Query Submissions

```javascript
// In serverless handler
const submissions = await fazt.services.forms.list('contact');

// With options
const recent = await fazt.services.forms.list('contact', {
  limit: 10,
  order: 'desc'
});

// Single submission
const sub = await fazt.services.forms.get('contact', submissionId);

// Delete
await fazt.services.forms.delete('contact', submissionId);
```

## Redirect After Submit

Use hidden field convention:

```html
<form action="/_services/forms/contact" method="POST">
  <input type="hidden" name="_next" value="/thanks.html">
  <!-- ... other fields ... -->
</form>
```

Service redirects to `_next` after successful submission.
Field is not stored.

## Spam Protection

### Automatic (No Setup)

| Protection       | How                                         |
| ---------------- | ------------------------------------------- |
| **Origin check** | Rejects POST from different domains         |
| **Rate limit**   | 10 submissions/minute per IP (configurable) |

### Optional Honeypot

```html
<form action="/_services/forms/contact" method="POST">
  <!-- Hidden field - bots fill it, humans don't -->
  <input name="_hp" style="display:none" tabindex="-1" autocomplete="off">

  <!-- Real fields -->
  <input name="email" type="email">
  <button type="submit">Send</button>
</form>
```

If `_hp` has a value, submission is silently rejected.
Field is not stored.

## Reserved Fields

Fields starting with `_` are special:

| Field        | Purpose                     |
| ------------ | --------------------------- |
| `_next`      | Redirect URL after submit   |
| `_hp`        | Honeypot (reject if filled) |
| `_submitted` | Auto-added timestamp        |
| `_ip`        | Auto-added client IP        |

## HTTP Endpoint

### Submit Form

```
POST /_services/forms/{name}
Content-Type: application/x-www-form-urlencoded

email=user@example.com&message=Hello
```

**Success (no `_next`):**
```
HTTP 200 OK
Content-Type: application/json

{ "id": "sub_abc123", "status": "ok" }
```

**Success (with `_next`):**
```
HTTP 303 See Other
Location: /thanks.html
```

**Errors:**

| Code | Reason              |
| ---- | ------------------- |
| 403  | Origin check failed |
| 429  | Rate limit exceeded |

## JS API

```javascript
fazt.services.forms.list(name, options?)
// options: { limit, offset, order }
// Returns: [{ ...fields, _submitted, _ip, _id }, ...]

fazt.services.forms.get(name, id)
// Returns: { ...fields, _submitted, _ip, _id }

fazt.services.forms.delete(name, id)
// Returns: { deleted: true }

fazt.services.forms.count(name)
// Returns: number

fazt.services.forms.clear(name)
// Deletes all submissions for this form
```

## Storage

```sql
CREATE TABLE svc_forms (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    form_name TEXT NOT NULL,
    data TEXT NOT NULL,          -- JSON blob
    ip TEXT,
    submitted_at INTEGER NOT NULL
);

CREATE INDEX idx_forms_app_name ON svc_forms(app_uuid, form_name);
CREATE INDEX idx_forms_submitted ON svc_forms(submitted_at DESC);
```

## Limits

| Limit                    | Default |
| ------------------------ | ------- |
| `maxSubmissionsPerForm`  | 10,000  |
| `maxFieldSize`           | 64 KB   |
| `maxFieldsPerSubmission` | 100     |
| `ratePerMinutePerIP`     | 10      |
| `retentionDays`          | 90      |

## CLI

```bash
# List forms with submission counts
fazt services forms list

# View submissions
fazt services forms show contact --limit 10

# Export as CSV
fazt services forms export contact > submissions.csv

# Clear old submissions
fazt services forms purge contact --older-than 30d
```

## Example: Contact Form

**HTML (your design):**
```html
<!DOCTYPE html>
<html>
<head>
  <title>Contact</title>
  <link rel="stylesheet" href="https://unpkg.com/@picocss/pico@1/css/pico.min.css">
</head>
<body>
  <main class="container">
    <h1>Contact Us</h1>
    <form action="/_services/forms/contact" method="POST">
      <input name="_hp" style="display:none" tabindex="-1">
      <input type="hidden" name="_next" value="/thanks.html">

      <label>Name
        <input name="name" required>
      </label>

      <label>Email
        <input name="email" type="email" required>
      </label>

      <label>Message
        <textarea name="message" required></textarea>
      </label>

      <button type="submit">Send Message</button>
    </form>
  </main>
</body>
</html>
```

**Serverless handler (query submissions):**
```javascript
// api/admin/submissions.js
module.exports = async (req) => {
  // Only owner can view
  if (!fazt.security.isOwner()) {
    return { status: 401 };
  }

  const submissions = await fazt.services.forms.list('contact', {
    limit: 50,
    order: 'desc'
  });

  return { json: submissions };
};
```
