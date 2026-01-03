# v0.11 - Distribution

**Theme**: App ecosystem, marketplace, and package management.

## Summary

v0.11 introduces the "Linux distro" model for app distribution. Marketplaces
are Git repositories that serve as package sources. Apps can be installed,
updated, and removed like `apt-get` packages.

## Goals

1. **Marketplace Model**: Git repos as app sources
2. **Package Manager**: `fazt app install/update/remove`
3. **Manifest System**: `app.json` for metadata and permissions
4. **Source Tracking**: Personal vs marketplace apps

## Key Changes

| Capability  | Description                       |
| ----------- | --------------------------------- |
| Marketplace | Git-based app repositories        |
| App Install | One-command installation          |
| Manifest    | `app.json` with permissions       |
| Updates     | Check and apply app updates       |
| Sources     | Track origin of each app          |
| Profiles    | Auto-detect project type & deploy |

## Documents

- `marketplace.md` - Git-based app distribution
- `manifest.md` - The `app.json` specification
- `profiles.md` - Deployment profiles (Vite, Astro, Jekyll, etc.)

## Dependencies

- v0.8 (Kernel): App UUIDs for identity
- v0.10 (Runtime): Serverless execution for app logic

## Risks

- **Trust**: Marketplace apps could be malicious
- **Versioning**: Breaking changes in updates
- **Dependency Hell**: Apps requiring specific Fazt versions
