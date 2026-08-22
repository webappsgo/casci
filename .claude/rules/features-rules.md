# Features Rules (PART 18-23)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Use an external scheduler (cron, Task Scheduler, systemd timers) — the built-in scheduler handles ALL scheduled tasks, no exceptions
- Print scheduler task activity to the console — it logs to log files only
- Embed GeoIP databases in the binary — downloaded on first run, updated by the built-in scheduler
- Treat GeoIP (country/ASN/city) as the sole access-control gate — it's one risk signal among many; a blocked country still passes through every other security layer, an allowed country skips no check
- Redirect `/metrics` root alias to the versioned metrics endpoint — same handler, never a redirect (breaks scrapers)
- Accept a metrics bearer token via query string — header (`Authorization: Bearer {token}`) only
- Store the restore/backup password — always prompted on-demand, never persisted
- Show update/version status on public pages — Tier 3 info, operator concern only
- Skip checksum verification before applying a self-update — mandatory, refuse to update without it

## CRITICAL - ALWAYS DO
- Ship every account email (welcome, password reset, verification, login alert, security alert, breach notification, etc.) in the required format with its template-specific variables
- Keep the WebUI notification system available regardless of SMTP configuration — it's always on
- Run the built-in scheduler continuously — no enable/disable setting
- Distribute and lock scheduled tasks correctly in cluster mode to avoid duplicate execution
- Redact credentials in Loki/log-stream metrics output the same as log files
- Expose the required metrics categories (application, HTTP, DB, auth, plus conditional cache/scheduler/cluster/Tor/rate-limit metrics)
- Encrypt backups and never store the backup/restore password
- Force primary-admin re-authentication after a restore
- Verify update checksums before replacing the running binary

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Can we use cron for scheduled tasks? | No — never, under any circumstance; built-in scheduler only | PART 19, NEVER Use External Schedulers |
| Is GeoIP blocking sufficient alone? | No — always combined with rate limiting, auth, authz, input validation, audit logging | PART 20, GeoIP Is a Risk Signal |
| How is the metrics token passed? | `Authorization: Bearer` header only, never query string | PART 21, Authentication |
| Is the backup password stored? | Never — always prompted on-demand | PART 22, Backup Encryption |
| Are update checks shown on public pages? | Never — Tier 3 operator-only information | PART 23, Update Configuration |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| Built-in scheduler | The only permitted task scheduler; always running, cluster-aware task locking |
| GeoIP risk signal | Country/ASN/city data used to raise risk score, never as sole gate |
| `/metrics[/{service}]` | Root metrics alias — same handler as versioned endpoint, never redirected |
| Update channel / `defer_days` | Self-update configuration controlling rollout timing |

## QUICK REFERENCE
- Email: SMTP auto-detected, sane defaults, required template formats for every account event.
- Notifications: WebUI always available; email only per the notification-vs-email decision matrix.
- Scheduler: always running, built-in only, cluster task locking, logs only (never console).
- GeoIP: downloaded at runtime, never embedded, never the sole gate.
- Metrics: header-auth only, required categories, root alias never redirects.
- Backup/Restore: encrypted, password never stored, checksum-verified self-update.

---
For complete details, see AI.md PART 18-23
