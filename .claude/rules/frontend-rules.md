# Frontend Rules (PART 16, 17)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Let long strings (IPv6, Tor addresses, tokens, hashes) overflow their container or break mobile layout
- Show blank space for an empty list/table/data view — every one needs a proper empty state
- Store session tokens in `localStorage` or `IndexedDB` — session lives only in an `HttpOnly`+`Secure`+`SameSite` cookie
- Use default/unstyled JavaScript UI elements (`alert()`, `confirm()`, native `<select>` styling, etc.) — always use custom styled components
- Pass user-controlled content through `template.HTML` unless it went through an approved sanitizer first
- Link to `/server/{admin_path}` from any public route (`/**`)
- Show a regular authenticated (non-admin) user the login form again or reveal that an admin panel exists
- Let `{admin_path}` embed a `server/` prefix, contain a slash, or collide with reserved paths (`api`, `static`, `assets`, `health`, `version`, …)
- Remove `administration` from the set of valid `admin_path` values — it is the permanent revert-on-error fallback
- Use generic placeholder content on `/server/about`, `/server/help`, or similar standard pages — content must come from `IDEA.md`
- Render `server.contact.admin.email` on the public `/server/contact` page — never public
- Load analytics/tracking scripts when cookie consent is declined
- Skip embedding security-relevant data that needs frequent updates — never embed it in the binary
- Use inline `<style>` attributes — no inline styles

## CRITICAL - ALWAYS DO
- Build mobile-first responsive layouts using the defined breakpoints and container rules
- Support full CRUD in every response mode (HTML, JSON, plain text) via smart content detection
- Keep session cookies `HttpOnly` + `Secure` + `SameSite` — no token-handling JavaScript
- Enable the cookie-consent banner always — the app uses cookies for sessions/preferences
- Sanitize `custom_html` (footer, etc.) before rendering — strip scripts/dangerous elements automatically
- Keep public and admin layouts (`public.tmpl` / `admin.tmpl`) strictly separated with their own header/nav/footer partials
- Match error pages to the active theme
- Enforce the admin route hierarchy strictly — no conflicting or overlapping routes
- Require password complexity, TOTP 2FA support, passkey/WebAuthn support, recovery keys, session timeout, API token security, audit logging, and rate limiting for every Server Admin account
- Prefer native HTML/CSS mechanisms over JavaScript; when JS is required, use external scripts, never inline
- Register and maintain the PWA service worker, manifest, and cache-versioning strategy when PWA support applies

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Where does the session token live? | `HttpOnly`+`Secure`+`SameSite` cookie only — never localStorage/IndexedDB | PART 16, User Sessions in PWA |
| Can public pages link to admin? | Never, under any circumstance | PART 17, Admin Panel Isolation |
| Is cookie consent optional? | No — always enabled, the app uses cookies | PART 16, Cookie Consent Banner |
| Is `{admin_path}` free-form? | Constrained: single path segment, no slash, reserved-word validated, `administration` always valid as fallback | PART 17, Configurable Admin Path |
| What powers `/server/about` and `/server/help` content? | `IDEA.md` project-specific content — never generic placeholder text | PART 17, Standard Pages |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `{admin_path}` | Configurable single path segment after `/server/`, default `administration` |
| Public layout / Admin layout | Strictly separated template trees, never cross-linked |
| Smart content detection | Server logic choosing HTML/JSON/text response by client type, all supporting full CRUD |
| Server Admin | Privileged account type distinct from regular multi-user accounts |

## QUICK REFERENCE
- Mobile-first, semantic HTML, custom-styled components (never native browser dialogs).
- Empty states everywhere; no blank space.
- Admin panel fully isolated from public routes, own layout, own auth flow.
- Cookie consent always on; declined consent blocks all tracking.
- PWA: HTTPS-required service worker, versioned cache, install/update prompts, no credentials in JS-readable storage.

---
For complete details, see AI.md PART 16, 17
