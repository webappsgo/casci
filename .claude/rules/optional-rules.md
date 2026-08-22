# Optional Feature Rules (PART 34, 35, 36)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

These three PARTs describe OPTIONAL features. A project may skip them entirely — but once a project implements one, that entire PART becomes NON-NEGOTIABLE and must be followed exactly.

## CRITICAL - NEVER DO
- Implement multi-user, organizations, or custom domains "partially" — once started, the whole PART applies
- Implement organizations without PART 34 multi-user already in place
- Confuse Server Admin accounts (PART 17, `admins` table, always required) with Regular Users (PART 34, `users` table, optional)
- Let an admin set or view a user's password, 2FA secret, or private data
- Treat `server.orgs.creation.mode` as a per-org setting — it is server-level policy, distinct from per-org `visibility`
- Allow a user and an org to hold the same slug — users and orgs share one namespace, and collisions are blocked at creation
- Allow reserved names to be registered by either users or orgs
- Build a JS-dependent flow first — the server-rendered form → POST → redirect path must work with JavaScript disabled before any JS enhancement
- Use routing records (CNAME or A/AAAA) as proof of domain ownership — only the `_verify.{domain}` TXT record proves it
- Expose a private user's bio, location, website, email, or activity to org members via org-scoped visibility
- Store user accounts, tokens, or sessions in `server.db` — those belong in `users.db`

## CRITICAL - ALWAYS DO
- Default to admin-only mode; multi-user is opt-in via `users.enabled`
- Gate new regular-user account creation through exactly one setting: `users.registration.mode` (`open` default, `invite`, `admin_only`, `disabled`)
- Apply `auto_register` and the username collision rules when an OIDC/LDAP/SAML identity first creates a local user record
- Treat usernames and emails as case-insensitive: store lowercase, compare lowercase
- Accept user ID (all digits), username (no `@`), or email (contains `@`) as the login identifier, detected by format
- Keep the two-database split: `server.db` for admin credentials and server state, `users.db` for user accounts, tokens, and sessions
- Gate org creation through `server.orgs.creation.mode` (`open` default, `invite`, `admin_only`, `disabled`), assigning the initial owner per that mode
- Validate org slugs exactly like usernames: 2-39 chars, lowercase alphanumeric plus hyphens, start/end alphanumeric, no consecutive hyphens, not reserved, not taken
- Keep custom domains disabled by default (`server.features.custom_domains.enabled: false`) and enforce the configured per-user/per-org limits, reserved domains, and blocked patterns
- Verify domain ownership with a TXT record at `_verify.{domain}` matching the issued token, then instruct CNAME (subdomains) or A/AAAA (apex) for routing
- Select the ACME challenge by the defined logic: DNS-01 when a DNS provider is configured or the domain is a wildcard, otherwise TLS-ALPN-01, otherwise HTTP-01
- Return the defined domain error codes with their specified HTTP statuses (`DOMAIN_EXISTS` 409, `DOMAIN_RESERVED` 400, `TXT_RECORD_MISSING` 400, `SSL_ISSUANCE_FAILED` 500, and the rest)
- Give server admins the full domain control set: list all, suspend, unsuspend, force delete, force SSL renewal, view audit log

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Is multi-user required? | No — admin-only is the default; Server Admin is always required, Regular Users are optional | PART 34, Overview |
| What is the default registration mode? | `open` — anyone can self-register | PART 34, Registration Mode Definitions |
| Can an admin set a user's password? | No — only the user can, via invite link or reset | PART 34, Admin Permission Reminder |
| How many databases? | Two: `server.db` and `users.db`, for isolation, security, and independent backup | PART 34, Database Architecture |
| Do organizations require multi-user? | Yes — PART 34 must be implemented first | PART 35, header note |
| Who may create organizations by default? | Any authenticated user (`mode: open`), who becomes the initial owner | PART 35, Organization Creation Modes |
| Can a username and an org slug match? | No — shared namespace; a username `acme` blocks org slug `acme` | PART 35, Shared Namespace |
| Are custom domains on by default? | No — `enabled: false` | PART 36, Feature Configuration |
| What proves domain ownership? | A TXT record at `_verify.{domain}` containing the verification token | PART 36, Verification Flow |
| Subdomain or custom domain first? | Start with subdomains; add custom domains only when users request them | PART 36, Custom Domains vs Subdomains |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| Server Admin | PART 17 administrative account in the `admins` table; required in every project |
| Regular User | PART 34 end-user account in the `users` table; optional |
| Registration mode | `users.registration.mode` — controls creation of new regular-user accounts only, not login or profile visibility |
| Creation mode | `server.orgs.creation.mode` — server-level policy for who may create an organization |
| Organization | Canonical internal term; may be labeled team/workspace/group in user-facing copy, but routes and schema stay `org`/`orgs` |
| Org-scoped visibility | A private user opting to expose username, display name, and avatar to fellow org members only |
| `_verify.{domain}` | The TXT record name where a custom-domain owner publishes the verification token |

## QUICK REFERENCE
- All three PARTs are opt-in; implementing one makes it fully binding.
- PART 34: `users.enabled` + `registration.mode` (open/invite/admin_only/disabled); two DBs; case-insensitive username/email; login by ID, username, or email.
- PART 35: requires PART 34; `server.orgs.creation.mode` (open/invite/admin_only/disabled); slug rules identical to usernames; one shared user/org namespace.
- PART 36: disabled by default; TXT ownership verification then CNAME/A-AAAA routing; challenge order DNS-01 → TLS-ALPN-01 → HTTP-01; admins can suspend, delete, and force renewal.
- Every flow in all three PARTs is built no-JS-first.

---
For complete details, see AI.md PART 34, 35, 36
