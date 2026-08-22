---
name: Bug report
about: Report incorrect or unexpected behavior in casci
title: "[bug] "
labels: ["bug"]
assignees: []
---

<!-- Do NOT use this template for security vulnerabilities.
     Report those privately: https://github.com/webappsgo/casci/security/advisories/new -->

## Security Impact

Could this bug have security impact (authentication bypass, credential exposure,
privilege escalation, remote code execution, data leak)?

- [ ] No — this is a functional bug and is safe to discuss publicly
- [ ] Yes — **stop here**, close this issue and report privately at
      <https://github.com/webappsgo/casci/security/advisories/new>

## Version

Output of `casci --version`, or the commit SHA / image tag you are running:

```text

```

## Environment

- Platform and architecture (e.g. linux/amd64, darwin/arm64):
- Deployment mode (single binary, systemd service, Docker standard, Docker all-in-one):
- Container image tag, if applicable:
- Database backend (SQLite, PostgreSQL, ...):
- Reverse proxy in front of casci (none, nginx, Caddy, Traefik, ...):

## Expected Behavior

What you expected casci to do.

## Actual Behavior

What casci actually did.

## Steps to Reproduce

1.
2.
3.

Include the exact commands, requests, or pipeline definition needed to reproduce.
Minimal reproductions get fixed fastest.

## Logs and Screenshots

Relevant server logs, build output, or screenshots.

**Redact before pasting:** API tokens, session tokens, passwords, private keys,
webhook secrets, SSH keys, and any personal data. Replace the value and keep the key
name, for example `api_token=xxxxx`.

```text

```

## Additional Context

Anything else that helps — when it started, whether it worked in a previous version,
related issues, or workarounds you have found.
