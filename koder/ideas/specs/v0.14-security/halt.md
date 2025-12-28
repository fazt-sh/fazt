# Human-in-the-Loop: fazt.halt()

## Summary

`fazt.halt()` pauses execution and waits for human approval. Essential for
autonomous agents that need oversight before taking significant actions.

## Usage

```javascript
// Agent wants to deploy code
const approval = await fazt.halt(
    'Deploy new feature?',
    {
        action: 'deploy',
        app: 'blog',
        changes: diff
    }
);

if (approval.approved) {
    await fazt.kernel.deploy(...);
}
```

## Flow

1. Agent calls `fazt.halt(reason, data)`
2. Kernel pauses the process
3. Push notification sent to owner (via ntfy)
4. Owner opens `os.<domain>/approvals`
5. Reviews the request, clicks Approve/Deny
6. Process resumes with result

## Configuration

```json
{
  "halt": {
    "notify": ["ntfy:topic", "email:owner@example.com"],
    "timeout": 86400000,
    "defaultAction": "deny"
  }
}
```

## Use Cases

- AI agent wants to deploy code: require approval
- Scheduled job wants to send emails: require approval
- Harness wants to modify routing: require approval

## Dashboard

`/os/approvals` shows:
- Pending requests
- History of approvals/denials
- Who approved, when
