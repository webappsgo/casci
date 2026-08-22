# Binary Rules (PART 7, 8, 33)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Embed security databases in the binary — they download on first run and are kept updated by the built-in scheduler
- Let child processes run as root — credentials are applied at spawn time, never inherited as root
- Hardcode `localhost`, `127.0.0.1`, `0.0.0.0`, `[::1]`, or any static host/IP for display
- Display a bare route (`GET /api/`) without the full URL
- Include `:80` (HTTP) or `:443` (HTTPS) in a displayed URL — always stripped
- Construct URLs with raw, unencoded user input — always use the reusable URL-encoding helper
- Use `strconv.ParseBool()` for any boolean/truthy flag — use `config.ParseBool()` / `config.IsTruthy()`
- Implement UI mode flags (`--gui`, `--tui`, `--cli`) — display mode is auto-detected, never user-selected
- Attempt GUI launch over SSH/mosh — remote sessions always use TUI or CLI
- Use OS system directories for CLI runtime state — use the project's own config/data directory structure
- Skip the CLI client on any project — it is REQUIRED for every project, no include/skip decision

## CRITICAL - ALWAYS DO
- Detect stale PID files on startup (crash/`kill -9` can leave one behind) and handle them
- Support `NO_COLOR` across ALL binaries (server, CLI, agent)
- Build the URL as `{proto}://{fqdn}:{port}/path`, detecting `{proto}`, `{fqdn}`, `{port}` from request context, reverse-proxy headers preferred
- Provide shell completions for the CLI — built-in, non-negotiable
- Ship a companion CLI for every server project; treat the agent binary as a separate per-project decision (needed only for on-host monitoring/remote-management use cases)
- Apply smart user/org context detection from token scope for CLI commands (non-negotiable)
- When PART 34/35/36 (Multi-user/Organizations/Custom Domains) is adopted, flip its marker from OPTIONAL to REQUIRED — NON-NEGOTIABLE in the project's own `AI.md`, and apply every requirement in the PART in full

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Is a CLI required? | Yes, for every project, always | PART 33, CLI Is Required for All Projects |
| Is an agent required? | Only if the project needs on-host monitoring/remote management — otherwise skip it entirely | PART 33, Agent Binary |
| How to parse CLI booleans? | `config.ParseBool()` / `config.IsTruthy()`, never `strconv.ParseBool()` | PART 33, Boolean Handling |
| How is display mode (CLI/TUI/GUI) chosen? | Auto-detected from environment — never an explicit user flag | PART 33, Modes |
| How are URLs built for display? | `{proto}://{fqdn}:{port}/path`, ports 80/443 stripped, never hardcoded host | PART 8, URL & FQDN Detection |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| Server binary | Main long-running service binary |
| CLI (client) | Required companion binary for user interaction with the server |
| Agent | Optional binary that runs directly on a target system (not containerized) for monitoring/remote management |
| Display mode hierarchy | Auto-detection logic choosing CLI/TUI/GUI based on environment, never a flag |

## QUICK REFERENCE
- Single static binary, embedded assets, external security data downloaded at runtime.
- Children of privileged processes drop to non-root before running.
- Every user-facing URL built from proxy-aware `{proto}/{fqdn}/{port}` detection, never hardcoded.
- All boolean flags parsed via `config.ParseBool()`/`config.IsTruthy()`.
- CLI is mandatory; agent is optional and evaluated per project.
- Remote/SSH sessions never attempt GUI — TUI or CLI only.

---
For complete details, see AI.md PART 7, 8, 33
