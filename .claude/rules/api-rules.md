# API Rules (PART 13, 14, 15)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Expose sensitive data in `/server/healthz` or `/api/{api_version}/server/healthz` — healthz is public
- Keep legacy/removed endpoints from a project's own history — delete them, never keep for "compatibility"
- End a route with a trailing slash
- Reach for JavaScript when an HTML/CSS-native mechanism exists
- Do client-side what belongs server-side (validation, business logic, rendering that can be server-rendered)
- Manually edit generated OpenAPI JSON or GraphQL schema files
- Put Swagger/GraphQL files anywhere but `src/swagger/` and `src/graphql/` — never project root
- Guess at a third-party/compatibility API's shape — always look up the actual documentation
- Include stack traces, internal IPs, DB structure, or "username vs password" disclosure in an error `message`
- Hardcode the startup banner's "last update" date — always derive from `{build_datetime}`/`BuildEpoch`
- Use icons, ASCII art, or special characters in log output — logs are always raw text
- Include `:80`/`:443` in a displayed FQDN/URL

## CRITICAL - ALWAYS DO
- Use SemVer for versioning; source `--version` output per the defined priority order
- Follow the mandatory route rules: plural resource names, no trailing slash, org routes slug-required, user routes ID-not-required
- Apply content negotiation by `Accept` header / extension, `.txt` always plain text at priority 1
- Format every response type (JSON, TXT, HTML, YAML, Go, CSS, JS) per its formatting rules, consistent indentation throughout a file
- Use the standard response envelope for single item / action / error / paginated (default 250 items) responses
- Support the built-in Let's Encrypt challenge types and DNS-01 provider configuration as specified
- Keep overlay-network URLs (Tor/I2P) as `http://` always, regardless of clearnet TLS config

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Are legacy endpoints kept for compatibility? | No — delete them; only genuine external-API compatibility endpoints are kept, and those follow the RFC/documented spec exactly | PART 13, Legacy vs Compatibility |
| Where do Swagger/GraphQL files live? | `src/swagger/`, `src/graphql/` only, relative to `src/`, never root | PART 14, API Types |
| Default pagination size? | 250 items | PART 14, Response Standards |
| What's the `.txt` content-negotiation priority? | Priority 1 — always plain text | PART 14, Content Negotiation |
| Is healthz data ever sensitive? | Never — it's a public endpoint, public-safe fields only | PART 13, Security: Public Info Only |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `{api_version}` | Versioned API path segment, e.g. `/api/{api_version}/...` |
| `{build_datetime}` | Template variable derived from embedded `BuildEpoch`, used for "last update" display |
| Content negotiation | Response format chosen by `Accept` header or file extension |
| Legacy endpoint | An old/removed route from the project's own history — always deleted, never kept |
| Compatibility endpoint | A route implementing a documented external/RFC API for interop — kept and spec-accurate |

## QUICK REFERENCE
- Healthz: public, HTML/JSON/plain-text formats, public-safe fields only.
- Routes: plural, no trailing slash, slug for orgs, ID-free for user-scoped.
- JS is last resort; prefer server-side/HTML-native.
- Swagger/GraphQL generated only, live under `src/`.
- Error `message` field is always safe to show end users — no internals.
- Let's Encrypt DNS-01 + FQDN resolution built in; dev TLDs handled specially.

---
For complete details, see AI.md PART 13, 14, 15
