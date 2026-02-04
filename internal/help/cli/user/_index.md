---
command: "user"
description: "User management commands"
syntax: "fazt user <command> [options]"
version: "0.24.7"
updated: "2026-02-04"

examples:
  - title: "List all users"
    command: "fazt user list"
    description: "List users on the local instance"
  - title: "List users with pagination"
    command: "fazt user list --limit 50"
    description: "List first 50 users"
  - title: "Show user status"
    command: "fazt user status --email user@example.com"
    description: "Show user details and apps they have data in"
  - title: "Set user role"
    command: "fazt user set-role --email user@example.com --role admin"
    description: "Promote user to admin role"

related:
  - command: "app"
    description: "App management commands"
  - command: "alias"
    description: "Alias management commands"
---

# fazt user

Manage users - list, view status, and set roles.

## Commands

- `list` - List all users with pagination
- `status` - Show user status with app data (requires `--email` or `--id`)
- `set-role` - Set a user's role (requires `--email` or `--id`, and `--role`)

## Options

### Common Options
- `--email <email>` - User email address
- `--id <id>` - User ID (fazt_usr_xxx format)

### List Options
- `--app <app-id>` - Filter users by app (users with data in this app)
- `--offset <n>` - Skip first n results (default: 0)
- `--limit <n>` - Max results to return (default: 20)

### Set-Role Options
- `--role <role>` - Role to set: `user`, `admin`, or `owner`

## RBAC Rules

- **owner**: Can set any role on any user
- **admin**: Can set user/admin roles, but NOT owner; cannot modify owners
- **user**: Cannot manage users

## Remote Execution

```bash
fazt @zyt user list
fazt @zyt user status --email user@example.com
fazt @zyt user set-role --email user@example.com --role admin
```

## API Endpoints

- `GET /api/users` - List users (paginated)
- `GET /api/users/{id}/status` - User status with app data
- `POST /api/users/role` - Set user role

All endpoints require admin/owner role (session auth) or API key auth.
