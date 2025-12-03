# Release Workflow

When releasing a new version of Fazt (e.g., `v0.7.0`), follow this checklist:

## 1. Code & Tests
- [ ] Ensure all features for the release are merged to `master`.
- [ ] Run full test suite: `make test` (Must pass).
- [ ] Verify local build: `go build ./cmd/server`.

## 2. Documentation & Changelog
- [ ] Update `CHANGELOG.md` in the root with detailed notes.
- [ ] **Crucial**: Update `docs/changelog.json` for the website popup.
  - Add a new entry at the top.
  - `title`: Max 30 chars (e.g., "Database Replication").
  - `description`: Max 60 chars (e.g., "Added Litestream support for S3 backups.").
  - `created_at`: YYYY-MM-DD.

## 3. Tag & Push
- [ ] Tag the release:
  ```bash
  git tag v0.7.0
  git push origin master --tags
  ```
- [ ] This triggers the GitHub Action to build binaries and draft the release.

## 4. Verify Website
- [ ] Check `fazt.sh` (GitHub Pages) after a few minutes.
- [ ] Click the version tag in the navbar to verify the new changelog item appears.
