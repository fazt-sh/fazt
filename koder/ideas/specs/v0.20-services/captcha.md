# Captcha Service

## Summary

Simple challenge-response spam protection. No external dependencies,
no JavaScript required. Works with forms, APIs, and any user input.

## Challenge Types

| Type | Description | Difficulty |
|------|-------------|------------|
| `math` | Simple arithmetic (3 + 7 = ?) | Easy |
| `text` | Type distorted characters | Medium |
| `honeypot` | Hidden field trap | Invisible |

Default: `math` (simple, accessible, no JS needed)

## Usage

### Generate Challenge

```javascript
const challenge = await fazt.services.captcha.create();
// {
//   id: 'cap_abc123',
//   type: 'math',
//   question: 'What is 7 + 4?',
//   image: null,
//   expiresAt: '2024-01-15T10:35:00Z'
// }

// Text captcha (image-based)
const textChallenge = await fazt.services.captcha.create({ type: 'text' });
// {
//   id: 'cap_def456',
//   type: 'text',
//   question: 'Type the characters shown',
//   image: '/_services/captcha/cap_def456.png',
//   expiresAt: '...'
// }
```

### Verify Response

```javascript
const valid = await fazt.services.captcha.verify('cap_abc123', '11');
// true or false

// Captcha is consumed after verification (one-time use)
```

## HTML Form Integration

### Math Captcha (No JS)

```html
<form action="/_services/forms/contact" method="POST">
  <input name="email" type="email" required>
  <textarea name="message" required></textarea>

  <!-- Captcha -->
  <label>What is 7 + 4?</label>
  <input name="_captcha_id" type="hidden" value="cap_abc123">
  <input name="_captcha_answer" type="text" required>

  <button type="submit">Send</button>
</form>
```

### With Server-Side Rendering

```javascript
// api/contact-form.js
module.exports = async (req) => {
  const challenge = await fazt.services.captcha.create();

  return {
    html: `
      <form action="/_services/forms/contact" method="POST">
        <input name="email" type="email" required>
        <textarea name="message" required></textarea>

        <label>${challenge.question}</label>
        <input name="_captcha_id" type="hidden" value="${challenge.id}">
        <input name="_captcha_answer" type="text" required>

        <button type="submit">Send</button>
      </form>
    `
  };
};
```

### Text Captcha (Image)

```html
<form action="/api/signup" method="POST">
  <input name="email" type="email" required>
  <input name="password" type="password" required>

  <!-- Image captcha -->
  <img src="/_services/captcha/cap_def456.png" alt="Captcha">
  <input name="_captcha_id" type="hidden" value="cap_def456">
  <input name="_captcha_answer" type="text" placeholder="Type characters">

  <button type="submit">Sign Up</button>
</form>
```

## HTTP Endpoints

### Create Challenge

```
POST /_services/captcha
Content-Type: application/json

{ "type": "math" }
```

Response:
```json
{
  "id": "cap_abc123",
  "type": "math",
  "question": "What is 7 + 4?",
  "expiresAt": "2024-01-15T10:35:00Z"
}
```

### Get Image (text captcha only)

```
GET /_services/captcha/{id}.png
```

Returns PNG image with distorted text.

### Verify

```
POST /_services/captcha/{id}/verify
Content-Type: application/json

{ "answer": "11" }
```

Response:
```json
{ "valid": true }
```

## Forms Service Integration

Forms service automatically checks for captcha fields:

```html
<form action="/_services/forms/contact" method="POST">
  <!-- Regular fields -->
  <input name="email" type="email">

  <!-- Captcha fields (auto-validated by forms service) -->
  <input name="_captcha_id" type="hidden" value="cap_abc123">
  <input name="_captcha_answer" type="text">

  <button type="submit">Send</button>
</form>
```

If captcha fails, form submission is rejected with 400.

## JS API

```javascript
fazt.services.captcha.create(options?)
// options: { type, expiresIn }
// type: 'math' | 'text' | 'honeypot'
// Returns: { id, type, question, image?, expiresAt }

fazt.services.captcha.verify(id, answer)
// Returns: boolean
// Captcha is consumed after verification

fazt.services.captcha.image(id)
// Returns: PNG buffer (for text captcha)
```

## Headless API Usage

For APIs (mobile apps, SPAs):

```javascript
// api/captcha.js - Get new captcha
module.exports = async (req) => {
  if (req.method === 'POST') {
    const challenge = await fazt.services.captcha.create({
      type: req.json.type || 'math'
    });
    return { json: challenge };
  }
};

// api/signup.js - Verify on submit
module.exports = async (req) => {
  // Verify captcha first
  const valid = await fazt.services.captcha.verify(
    req.json.captchaId,
    req.json.captchaAnswer
  );

  if (!valid) {
    return { status: 400, json: { error: 'Invalid captcha' } };
  }

  // Proceed with signup...
};
```

## Math Captcha Details

Generates simple arithmetic:
- Addition: `3 + 7 = ?`
- Subtraction: `12 - 5 = ?`
- Multiplication: `4 × 3 = ?`

Constraints:
- Numbers 1-20
- Results always positive
- No division (avoids decimals)

## Text Captcha Details

Generates distorted image:
- 5-6 alphanumeric characters
- Excludes ambiguous: 0/O, 1/l/I
- Wavy distortion, noise lines
- 200x60 pixels, PNG format

Pure Go implementation using `image` package.

## Storage

```sql
CREATE TABLE svc_captcha (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    type TEXT NOT NULL,
    question TEXT,
    answer TEXT NOT NULL,        -- Hashed
    created_at INTEGER,
    expires_at INTEGER,
    used INTEGER DEFAULT 0
);

CREATE INDEX idx_captcha_expires ON svc_captcha(expires_at);
```

Expired captchas auto-cleaned.

## Limits

| Limit | Default |
|-------|---------|
| `expiresIn` | 5 minutes |
| `maxPendingPerIP` | 10 |
| `ratePerMinutePerIP` | 20 |

## CLI

```bash
# Generate test captcha
fazt services captcha create --type math

# View pending captchas
fazt services captcha list

# Clean expired
fazt services captcha cleanup
```

## Example: Protected Signup

```javascript
// api/signup.js
module.exports = async (req) => {
  if (req.method === 'GET') {
    // Render signup form with captcha
    const captcha = await fazt.services.captcha.create();

    return {
      html: `
        <form method="POST">
          <input name="email" type="email" required>
          <input name="password" type="password" required>

          <p>${captcha.question}</p>
          <input name="captchaId" type="hidden" value="${captcha.id}">
          <input name="captchaAnswer" required>

          <button type="submit">Sign Up</button>
        </form>
      `
    };
  }

  if (req.method === 'POST') {
    // Verify captcha
    const valid = await fazt.services.captcha.verify(
      req.form.captchaId,
      req.form.captchaAnswer
    );

    if (!valid) {
      return { status: 400, html: '<p>Wrong answer, try again</p>' };
    }

    // Create user...
    await fazt.storage.ds.insert('users', {
      email: req.form.email,
      password: hashPassword(req.form.password)
    });

    return { redirect: '/welcome' };
  }
};
```

## Accessibility

Math captcha is accessible:
- Screen reader friendly (plain text question)
- No images required
- Keyboard navigable

For text captcha, provide audio alternative:
```javascript
const challenge = await fazt.services.captcha.create({
  type: 'text',
  audio: true
});
// challenge.audio = '/_services/captcha/cap_abc123.mp3'
```
