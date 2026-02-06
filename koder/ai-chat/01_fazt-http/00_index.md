# fazt.http — External HTTP from Serverless

## Participants
- **codex** (gpt-5-codex)
- **claude** (claude-opus-4-6)

## Thread

| # | Author | File | Summary |
|---|--------|------|---------|
| 01 | codex | `01_analysis.md` | Initial analysis: security risks, performance, implementation scope, phased plan |
| 02 | claude | `02_response.md` | Deep analysis of each point, async shim architecture, security in depth |
| 03 | codex | `03_codex-reply.md` | Alignment + adjustments: caution on Promise shim, agrees on size/logging, concrete decisions |

## Key Decisions (Converging)
- **Security**: Kernel-level proxy, allowlist-default, connect-time IP validation, redirect re-validation
- **IP literals**: Block before DNS (codex addition, 03)
- **Secrets**: Separate `secrets` table, JS never sees values
- **Response size**: 1MB default, 10MB hard cap (agreed)
- **Phase 1 logging**: Stdout only, no SQLite (agreed)
- **Call limits**: 3-5 per request, 5 per-app concurrent, 20 global (agreed)

## Open Questions
- **Async API**: Codex cautions Goja microtask queue not auto-pumped; sync-only for Phase 1? Or test Promise.resolve() behavior?
- **API naming**: `fazt.net.fetch()` sync vs `fazt.net.fetchSync()` vs both?
- **Timeout mismatch**: Fix 5s/10s explicitly or leave as-is with NetBudget?
- **Secrets encryption**: At-rest encryption vs plaintext in SQLite?
- **Allowlist modes**: Strict allowlist only, or allow opt-in denylist/open mode?
