# Security Policy

## Supported Versions

Security fixes land on the latest release only. The `latest` GitHub release and the
`ghcr.io/webappsgo/casci:latest` container image are the supported artifacts.
Older tags, `beta`, `daily`, and `devel` builds receive no backported fixes — upgrade
to the newest release before reporting an issue.

| Channel | Supported |
|---------|-----------|
| Latest stable release (`v*` tag, `:latest` image) | Yes |
| `beta` branch / `:beta` image | Best effort, no guarantee |
| `daily` release / `:devel` image | No |
| Any older stable release | No |

## Reporting a Vulnerability

**Do not open a public issue for a vulnerability.** Public bug reports are for
functional defects only.

Report privately through GitHub private vulnerability reporting — this is the primary
and preferred channel:

<https://github.com/webappsgo/casci/security/advisories/new>

You can also reach it from the repository's **Security** tab → **Report a vulnerability**.

If you are looking at a mirror of this repository rather than
<https://github.com/webappsgo/casci>, report on the GitHub origin repository first.

A running casci instance also publishes its own contact paths:

- `/.well-known/security.txt` — machine-readable contact and policy pointers
- `/server/security` — the instance's security page
- `/server/contact?security_id=...` — contact form carrying a security reference id

The address published in `/.well-known/security.txt` is a secondary/CC contact only.
It is never the primary reporting path — always start with private vulnerability
reporting on GitHub.

## What to Include

- Affected version or commit (`casci --version`)
- Platform and deployment mode (binary, systemd, Docker standard, Docker all-in-one)
- A description of the impact and the attack preconditions
- Reproduction steps or a proof of concept
- Any logs or output, with tokens, passwords, and personal data redacted

## Disclosure Process

1. You report privately via the advisory link above.
2. We acknowledge the report and open a private advisory draft.
3. We confirm the issue, determine severity, and agree a fix timeline with you.
4. A fix is prepared in a private fork attached to the advisory.
5. The fix ships in a new stable release, and the advisory is published with credit
   to the reporter unless you ask to remain anonymous.

Please keep the report private until the advisory is published.
