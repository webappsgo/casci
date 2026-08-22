# Docker Rules (PART 27)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Place `Dockerfile` or `docker-compose.yml` in the project root — always under `docker/`
- Modify `ENTRYPOINT` or `CMD` — all customization goes through `entrypoint.sh`
- Include `build:` or `version:` keys in docker-compose files
- Use `.env`-only variables with no inline fallback, or create `.env`/`.env.example`/`.env.sample` files for compose
- Use list-style environment entries (`- KEY=value`) in compose — always map style (`KEY: value`)
- Set `MODE`/`DEBUG` in production compose — the binary defaults to production; only `docker-compose.dev.yml`/`docker-compose.test.yml` may set these
- Commit runtime `volumes/` created by local compose runs
- Run `docker compose` from inside the project directory
- Push `:dev` or `:test` image tags to the production registry
- Skip redacting credentials in any mode, including `MODE=debug` — always redacted, no exceptions
- AI assistants: use `docker-compose.yml` (dev) or the production compose file — those are human-use only; AI/automated testing uses `docker-compose.test.yml`

## CRITICAL - ALWAYS DO
- Use `docker/` for Dockerfiles/compose files, `docker/rootfs/` for build-time overlay files, and `entrypoint.sh` for container startup
- Use a multi-stage Dockerfile build — no pre-built binaries copied in
- Name databases `server.db` and `users.db` consistently across all projects
- Include OCI meta labels and multi-arch image annotations
- Hardcode all compose environment variables with sane, working defaults (map style); the stack must work with zero env file present
- Keep build-time `docker/rootfs/` and runtime `./volumes/` conceptually and physically separate

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Where do Docker files live? | `docker/` (Dockerfile, compose), `docker/rootfs/` (overlay), never project root | PART 27, Docker Directory Structure |
| Can `.env` files be used? | No — never; all vars hardcoded with sane defaults in compose | PART 27, Environment Variables |
| Which compose file do AI/automated tests use? | `docker-compose.test.yml` — dev/production compose files are human-use only | PART 27, Docker Compose (Test) |
| What are the DB filenames? | Always `server.db` and `users.db` | PART 27, Container Paths |
| Does debug mode relax credential redaction? | Never — credentials stay redacted in every mode | PART 27, Environment Variables (Entrypoint) |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| `entrypoint.sh` | The sole customization point for container startup — `ENTRYPOINT`/`CMD` are never edited directly |
| `docker/rootfs/` | Build-time filesystem overlay baked into the image |
| `./volumes/` | Runtime-only bind-mount directory next to the compose file, never committed |
| `docker-compose.test.yml` | The only compose file AI/automated testing may use |

## QUICK REFERENCE
- Multi-stage build, OCI labels, multi-arch annotations, `server.db`/`users.db` naming.
- Compose: no `build:`/`version:`, map-style hardcoded env vars, no `.env` files.
- AI never touches dev/production compose files — test compose only.
- Production images never carry `:dev`/`:test` tags.

---
For complete details, see AI.md PART 27
