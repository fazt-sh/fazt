# Comments Service

## Summary

Attach comments to any entity. Works like a generic feedback system -
comments on products, posts, items, whatever your app has. Supports
threading, author tracking, and basic moderation.

## How It Works

Comments are attached to a **target** - any string identifier your app uses:

```
Target                      Comments
──────────────────────────────────────
"product:123"        →      [comment, comment, ...]
"post:my-blog-post"  →      [comment, comment, ...]
"order:456"          →      [comment, comment, ...]
```

Your app defines what targets mean. The service just stores and retrieves.

## Usage

### Add Comment

```javascript
const comment = await fazt.services.comments.add('product:123', {
  body: 'Great product!',
  author: 'user@example.com',     // Optional
  authorName: 'John',             // Optional
  meta: { rating: 5 }             // Optional, any JSON
});

// Returns:
// {
//   id: 'cmt_abc123',
//   target: 'product:123',
//   body: 'Great product!',
//   author: 'user@example.com',
//   authorName: 'John',
//   meta: { rating: 5 },
//   createdAt: '2024-01-15T10:30:00Z',
//   status: 'visible'
// }
```

### List Comments

```javascript
const comments = await fazt.services.comments.list('product:123', {
  limit: 20,
  order: 'desc',               // 'asc' | 'desc'
  status: 'visible'            // 'visible' | 'hidden' | 'all'
});
```

### Threading (Replies)

```javascript
// Reply to a comment
const reply = await fazt.services.comments.add('product:123', {
  body: 'Thanks for the feedback!',
  author: 'admin@example.com',
  parentId: 'cmt_abc123'       // Parent comment ID
});

// Get threaded comments
const threaded = await fazt.services.comments.list('product:123', {
  threaded: true               // Returns nested structure
});

// Returns:
// [
//   {
//     id: 'cmt_abc123',
//     body: 'Great product!',
//     replies: [
//       { id: 'cmt_def456', body: 'Thanks for the feedback!' }
//     ]
//   }
// ]
```

## Moderation

### Hide/Show Comments

```javascript
// Hide a comment (soft delete)
await fazt.services.comments.hide('cmt_abc123');

// Show again
await fazt.services.comments.show('cmt_abc123');

// Permanent delete
await fazt.services.comments.delete('cmt_abc123');
```

### Approval Queue

Optional moderation mode - comments start hidden:

```javascript
// Enable approval mode for a target prefix
await fazt.services.comments.configure({
  requireApproval: ['product:*', 'post:*']
});

// List pending comments
const pending = await fazt.services.comments.list('product:123', {
  status: 'pending'
});

// Approve
await fazt.services.comments.approve('cmt_abc123');
```

## HTTP Endpoint

### Submit Comment (Public)

```
POST /_services/comments/{target}
Content-Type: application/json

{
  "body": "Great product!",
  "authorName": "John",
  "email": "john@example.com"
}
```

### Get Comments (Public)

```
GET /_services/comments/{target}?limit=20&order=desc
```

Returns only visible comments.

## Spam Protection

Integrates with captcha service:

```html
<form action="/_services/comments/product:123" method="POST">
  <textarea name="body" required></textarea>
  <input name="authorName" placeholder="Your name">

  <!-- Captcha integration -->
  <input name="_captcha" type="hidden" value="{{captchaToken}}">

  <!-- Honeypot -->
  <input name="_hp" style="display:none">

  <button type="submit">Submit</button>
</form>
```

Also supports:
- Rate limiting (per IP)
- Origin check (same as forms)

## JS API

```javascript
fazt.services.comments.add(target, options)
// options: { body, author, authorName, meta, parentId }
// Returns: comment object

fazt.services.comments.list(target, options?)
// options: { limit, offset, order, status, threaded }
// Returns: [comments]

fazt.services.comments.get(id)
// Returns: comment object

fazt.services.comments.update(id, options)
// options: { body, meta }
// Returns: updated comment

fazt.services.comments.delete(id)
// Permanent delete

fazt.services.comments.hide(id)
fazt.services.comments.show(id)
fazt.services.comments.approve(id)

fazt.services.comments.count(target)
// Returns: number

fazt.services.comments.configure(options)
// options: { requireApproval: ['prefix:*'] }
```

## Storage

```sql
CREATE TABLE svc_comments (
    id TEXT PRIMARY KEY,
    app_uuid TEXT NOT NULL,
    target TEXT NOT NULL,
    parent_id TEXT,
    body TEXT NOT NULL,
    author TEXT,
    author_name TEXT,
    meta TEXT,                   -- JSON
    status TEXT DEFAULT 'visible',  -- visible, hidden, pending
    created_at INTEGER,
    updated_at INTEGER
);

CREATE INDEX idx_comments_target ON svc_comments(app_uuid, target, status);
CREATE INDEX idx_comments_parent ON svc_comments(parent_id);
```

## Limits

| Limit | Default |
|-------|---------|
| `maxBodyLength` | 10,000 chars |
| `maxMetaSize` | 4 KB |
| `maxPerTarget` | 10,000 |
| `maxDepth` | 5 (threading levels) |
| `ratePerMinutePerIP` | 5 |

## CLI

```bash
# List comments for target
fazt services comments list product:123

# Show pending moderation
fazt services comments pending

# Approve comment
fazt services comments approve cmt_abc123

# Delete comment
fazt services comments delete cmt_abc123

# Export comments
fazt services comments export product:123 > comments.json
```

## Example: Product Reviews

```javascript
// api/reviews.js - Submit review
module.exports = async (req) => {
  const productId = req.params.id;

  const comment = await fazt.services.comments.add(`product:${productId}`, {
    body: req.json.review,
    authorName: req.json.name,
    meta: { rating: req.json.rating }
  });

  return { json: comment };
};
```

```javascript
// api/reviews/[id].js - Get reviews
module.exports = async (req) => {
  const productId = req.params.id;

  const reviews = await fazt.services.comments.list(`product:${productId}`, {
    limit: 50,
    order: 'desc'
  });

  // Calculate average rating
  const avgRating = reviews.reduce((sum, r) => sum + r.meta.rating, 0) / reviews.length;

  return {
    json: {
      reviews,
      averageRating: avgRating,
      totalReviews: reviews.length
    }
  };
};
```

## Example: Headless API for Mobile App

```javascript
// api/comments.js
module.exports = async (req) => {
  const { target } = req.query;

  if (req.method === 'GET') {
    const comments = await fazt.services.comments.list(target, {
      threaded: true
    });
    return { json: comments };
  }

  if (req.method === 'POST') {
    // Validate user from JWT/session
    const user = await validateUser(req);

    const comment = await fazt.services.comments.add(target, {
      body: req.json.body,
      author: user.id,
      authorName: user.name,
      parentId: req.json.parentId
    });

    return { json: comment };
  }
};
```
