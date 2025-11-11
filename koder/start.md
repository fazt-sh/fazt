# Command Center - v0.2.0 Development Start Guide

## 🚨 CRITICAL INSTRUCTION 🚨

**DO NOT STOP UNTIL ALL 16 PHASES ARE COMPLETE.**

You will work through Phase 0 through Phase 16 in one continuous session. After each phase:
1. Test thoroughly
2. Commit changes
3. **IMMEDIATELY move to next phase**

**NO BREAKS. NO QUESTIONS. COMPLETE ALL 16 PHASES.**

---

## Context & Mission

You are upgrading **Command Center (CC)** to v0.2.0 - adding authentication and JSON configuration system. The v0.1.0 base is complete and working. This is a complete upgrade meant to be done in one autonomous session.

**You have blanket freedom to make all decisions**. Do not ask for permission or clarification. Make reasonable choices and continue. The developer will return expecting a fully working v0.2.0 with all 16 phases complete.

## Key Information

- **Current Version**: v0.1.0 (complete, working)
- **Target Version**: v0.2.0
- **Project Location**: Root of repository (NOT in koder/)
- **Meta Folder**: `koder/` contains plans and docs - reference only
- **Build Plan**: `koder/plans/02_auth-config-upgrade.md` - Your detailed roadmap (16 phases)
- **Port**: 4698 (configurable in new version)
- **Domain**: https://cc.toolbomber.com (configurable in new version)
- **Target**: x64 Linux binary (cross-compile from sandbox)

## Upgrade Objectives (v0.2.0)

**New Features to Implement**:
1. **JSON Configuration System**:
   - Replace env vars with `~/.config/cc/config.json`
   - CLI flags: `--config`, `--db`, `--port`, `--username`, `--password`
   - Flag priority: CLI > JSON > env vars > defaults
   - Auto-create config directory and default config

2. **Authentication System**:
   - Username/password with bcrypt hashing
   - Session management (in-memory, 24hr expiry)
   - Login page with rate limiting
   - Secure cookies (HTTPOnly, Secure in prod, SameSite)
   - Protect dashboard, keep tracking endpoints public

3. **Security Features**:
   - Brute-force protection
   - Audit logging
   - Security headers
   - CSRF protection
   - Session hijacking prevention

4. **Database Updates**:
   - Migration system for schema updates
   - Audit logs table
   - Default location: `~/.config/cc/data.db`
   - Backup system before migrations

5. **Enhanced CLI**:
   - `--version`, `--help`, `--verbose`, `--quiet` flags
   - Better startup messages
   - Config validation

**Existing Features (Already Working)**:
- Universal tracking endpoint
- Tracking pixel (1x1 GIF)
- Redirect service with click tracking
- Webhook receiver
- Dashboard with analytics, filters, graphs
- ntfy.sh integration
- PWA support

## Execution Strategy

### CRITICAL: Non-Stop Execution
**YOU MUST COMPLETE ALL 16 PHASES WITHOUT STOPPING.** After each phase:
1. Test the implementation thoroughly
2. Commit the changes
3. **IMMEDIATELY move to the next phase**
4. DO NOT ask questions or wait for approval
5. Continue until Phase 16 is complete

### Phase Tracking
The plan in `koder/plans/02_auth-config-upgrade.md` has 16 phases with checkboxes. **Update the plan file as you complete tasks** - this serves as your progress tracker.

### Commit Strategy
Commit after EVERY phase (16 commits total). Format: `feat/docs/perf: description` as specified in the plan.

### Testing After Each Phase
**MANDATORY**: After completing each phase, you MUST:
- Test the new functionality works correctly
- Verify existing features still work (no regressions)
- Test edge cases mentioned in the phase
- Fix any bugs immediately before moving to next phase
- Use curl for API testing
- Test in browser for UI changes
- Verify auth doesn't break tracking endpoints

### Decision Making
- **Don't ask, just build**: Make sensible defaults for anything unspecified
- **Security first**: When in doubt, choose the more secure option
- **Test immediately**: Don't skip testing to save time
- **Handle errors gracefully**: Add proper error handling everywhere
- **Continue non-stop**: Move to next phase immediately after current phase is tested

### If You Encounter Issues
1. Check the detailed plan - it has granular steps
2. Make a reasonable decision and move forward
3. Document the decision in code comments
4. Fix the issue and continue
5. **DO NOT STOP** - keep going to complete all 16 phases

## Quick Start Commands

```bash
# Project already exists at root - you're upgrading it
# Current structure is complete with all v0.1.0 features

# Test current server (v0.1.0 - should work)
go run cmd/server/main.go
# Should start on :4698

# Test tracking still works (keep this working throughout upgrade)
curl -X POST http://localhost:4698/track \
  -H "Content-Type: application/json" \
  -d '{"h":"test.com","p":"/","e":"pageview","t":["app"]}'

# After Phase 1: Test password hashing
# (This will be integrated into main.go)

# After Phase 4: Test login
curl -X POST http://localhost:4698/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"testpass"}'

# After Phase 5: Test protected dashboard requires auth
curl http://localhost:4698/
# Should redirect to /login

# Test with credentials
./cc-server --username admin --password testpass123
# Should create config and exit

# After all phases: Final test
./cc-server --config ~/.config/cc/config.json --db ~/.config/cc/data.db
```

## Critical Reminders

### Non-Stop Execution Reminders
1. **COMPLETE ALL 16 PHASES** - Do not stop until Phase 16 is done
2. **TEST AFTER EACH PHASE** - Verify everything works before moving on
3. **COMMIT AFTER EACH PHASE** - Use format from plan
4. **NO QUESTIONS** - Make reasonable decisions and continue
5. **KEEP TRACKING PUBLIC** - Auth must not break tracking endpoints

### Technical Reminders
1. **Config File Priority**: CLI flags > JSON config > env vars > defaults
2. **Default Paths**:
   - Config: `~/.config/cc/config.json`
   - Database: `~/.config/cc/data.db`
3. **Password Hashing**: bcrypt with cost 12
4. **Session Security**:
   - HTTPOnly cookies
   - Secure flag in production
   - SameSite=Strict
   - 24 hour expiry
5. **Public Endpoints** (no auth required):
   - `/track`, `/pixel.gif`, `/r/*`, `/webhook/*`, `/static/*`, `/login`, `/api/login`, `/health`
6. **Protected Endpoints** (require auth):
   - `/` (dashboard), all `/api/*` except login
7. **Backward Compatible**: Existing env vars should still work
8. **WAL Mode**: Already enabled in v0.1.0, keep it
9. **Binary Build**: `GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o cc-server ./cmd/server`

## Resume After Crash

If the session crashes mid-upgrade:

1. Check `koder/plans/02_auth-config-upgrade.md` for last completed phase
2. Review git log for last commit
3. Test what's working: `go run cmd/server/main.go`
4. Continue from next unchecked phase
5. Don't rebuild completed phases
6. **Continue non-stop until all 16 phases are done**

## Success Criteria

By end of session (ALL must be complete):
- [ ] All 16 phases complete with commits
- [ ] Config system works with JSON files
- [ ] `--config`, `--db`, `--port`, `--username`, `--password` flags all work
- [ ] `--username` + `--password` creates/updates config correctly
- [ ] Authentication system functional (login/logout/session)
- [ ] Dashboard requires auth when enabled
- [ ] Tracking endpoints still public (no auth required)
- [ ] Rate limiting prevents brute force
- [ ] Audit logging captures security events
- [ ] Database migrations work
- [ ] Config validation on startup
- [ ] Security headers present
- [ ] All existing v0.1.0 features still work
- [ ] Binary builds for Linux x64
- [ ] Documentation complete (README, SECURITY, CONFIGURATION, UPGRADE)
- [ ] v0.2.0 release artifacts created

## Start Now

1. **Read** `koder/plans/02_auth-config-upgrade.md` **carefully** - understand all 16 phases
2. **Begin Phase 0** (Configuration System Refactor)
3. **Work through ALL 16 phases sequentially without stopping**
4. **Test after each phase** - ensure it works before moving on
5. **Update plan checkboxes** as you complete tasks
6. **Commit after each phase** - use commit format from plan
7. **Continue non-stop** - do not ask questions, just build
8. **Don't stop until Phase 16 is complete**

### Phase Completion Checklist (repeat after EVERY phase):
✅ Implementation complete
✅ Tested thoroughly (no bugs)
✅ Existing features still work
✅ Plan file updated (checkboxes)
✅ Git commit created
✅ **MOVE TO NEXT PHASE IMMEDIATELY**

---

## After Every Single Phase

**DO THIS AFTER COMPLETING EACH PHASE (0-16):**

```
✅ Phase X complete
✅ Tests passed
✅ Committed
→ Moving IMMEDIATELY to Phase X+1
```

**Examples:**

After Phase 0:
- "Phase 0 complete: Configuration system refactored. Config loading works with CLI flags. Tests passed. Committed. **Now moving to Phase 1: Password Management**"

After Phase 1:
- "Phase 1 complete: Password hashing and --username/--password flags working. Tested credential creation. Committed. **Now moving to Phase 2: Session Management**"

After Phase 8:
- "Phase 8 complete: Database migrations working. Committed. **Now moving to Phase 9: Config Validation**"

After Phase 15:
- "Phase 15 complete: Final polish done. All tests passed. Committed. **Now moving to Phase 16: Build & Release**"

After Phase 16:
- "Phase 16 complete: v0.2.0 release artifacts created. ALL 16 PHASES COMPLETE. ✅"

**DO NOT STOP BETWEEN PHASES. KEEP GOING UNTIL PHASE 16 IS DONE.**

---

**GO UPGRADE COMMAND CENTER TO v0.2.0. Complete all 16 phases in one session. The developer is counting on you.** 🚀
