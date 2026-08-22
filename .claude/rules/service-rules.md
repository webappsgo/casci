# Service Rules (PART 24, 25)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Reuse reserved/well-known system UIDs/GIDs for the service account, even if they appear available on the current system — other services may claim them later
- Escalate privileges implicitly — escalation is detected per-OS and applied only at defined moments (install, privileged port bind), never silently elsewhere
- Run the service's child/worker processes as root after startup

## CRITICAL - ALWAYS DO
- Detect privilege-escalation needs per OS (Linux, macOS, BSD, Windows) using the documented detection logic
- Create a dedicated `{internal_name}` system user/group for the service, choosing UID/GID via the defined selection logic and skipping the reserved list
- Ship service templates for every supported init system: systemd, OpenRC, SysVinit, runit, rc.d (FreeBSD), launchd (macOS), Windows Service
- Provide install/uninstall/disable logic that is idempotent and OS-appropriate
- Provide accurate `--help` output for service, maintenance, shell, update, and CLI-admin subcommands

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| What account runs the service? | Dedicated `{internal_name}` system user/group, never root after startup | PART 24, System User Requirements |
| Which init systems are supported? | systemd, OpenRC, SysVinit, runit, rc.d, launchd, Windows Service | PART 25, Service Templates |
| Can we reuse an "available" low UID? | No — the reserved/well-known UID list is off-limits regardless of current availability | PART 24, Reserved/Well-Known UIDs |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `{internal_name}` | The project's internal system-user/group name for its service account |
| Escalation detection | Per-OS logic determining whether the current invocation needs elevated privileges |
| Service templates | Init-system-specific unit/script files generated for install |

## QUICK REFERENCE
- One dedicated non-root service account, safe UID/GID selection.
- Full init-system template coverage across Linux/BSD/macOS/Windows.
- Install/uninstall/disable are OS-aware and idempotent.
- Help output defined and consistent for every service-related subcommand.

---
For complete details, see AI.md PART 24, 25
