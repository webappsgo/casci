# Configuration Rules (PART 5, 6, 12)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Place YAML comments inline — always above the setting
- Use `strconv.ParseBool()` directly — use `config.ParseBool()` for consistent boolean handling across the app
- Construct file paths from user input without validating the result stays within bounds
- Imply or auto-enable `debug` mode — it is explicit opt-in only (`--mode debug` / `MODE=debug`)
- Disable authentication or any security check via debug mode or debug flag, in any mode including production
- Bypass admin authentication for any reason, in any mode
- Leave credentials (keys, tokens, passwords, secrets) unredacted in debug/diagnostic output

## CRITICAL - ALWAYS DO
- Normalize and validate every filesystem/URL path through the path-security functions before use
- Use `config.ParseBool()` for all boolean config parsing
- Treat the database as the configuration source of truth in cluster mode; `server.yml` acts as cache/backup and syncs from DB
- Let the server self-heal from recoverable errors; only genuinely critical errors enter maintenance mode
- Serve a maintenance dashboard at `/server/{admin_path}` and correct `/server/healthz` status while in maintenance mode
- Keep debug mode's effect scoped to verbosity/diagnostics only — never security behavior
- Redact credentials in all debug/diagnostic endpoints and logs, no exceptions

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Config source of truth? | Database in cluster mode; `server.yml` is cache/backup, synced from DB | PART 5, Configuration Storage |
| How to parse booleans? | `config.ParseBool()`, never `strconv.ParseBool()` | PART 5, Boolean Handling |
| How is debug mode enabled? | Explicit only: `--mode debug` / `MODE=debug`, or `--debug` / `DEBUG=true` flag | PART 6, Debug Mode |
| Does debug mode weaken security? | Never — auth and security checks stay fully enforced in every mode | PART 6, Debug Flag |
| Default port behavior? | Random high port (64xxx range) unless explicitly configured | PART 12, Port Rules |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| Production mode | Default operational state, no debug diagnostics |
| Development mode | Local-dev-oriented state, still security-enforced |
| Debug mode (`MODE=debug`) | Explicit opt-in state that defaults the debug flag on |
| Debug flag (`--debug`/`DEBUG=true`) | Enables diagnostics/pprof endpoints regardless of mode; never touches auth |
| Maintenance mode | Auto-entered on critical/unrecoverable error; serves limited dashboard + healthz |

## QUICK REFERENCE
- Six operational states total: production, development, debug, plus maintenance overlay and mode/debug flag combinations.
- Config file is `server.yml`; DB is authoritative in cluster mode.
- Booleans: `config.ParseBool()` everywhere.
- Debug endpoints (`/debug/*`, pprof, expvar) exist only when `--debug`/`DEBUG=true`, and never bypass auth.
- Privileged port binding (<1024) uses one-time escalation at service-install time, not per-request.

---
For complete details, see AI.md PART 5, 6, 12
