# AI Assistant Rules (PART 0, 1)

⚠️ **These rules are NON-NEGOTIABLE. Violations are bugs.** ⚠️

## CRITICAL - NEVER DO
- Guess or assume a value/behavior — verify with a command (`date`, `basename "$PWD"`, `git config user.email`, `git rev-parse --short HEAD`, `uname -m`) or read the spec; ask if neither applies
- Rush or skip verification steps — never claim "done" without checking
- "Improve" or "optimize" the spec, or invent patterns not in `AI.md`
- Modify PARTS 0-36 of `AI.md` — implementation patterns are fixed
- Create report/analysis files documenting issues found — fix them directly instead
- Rely on memory for spec details — read the relevant PART when needed; never load speculatively
- Add unrequested features
- Add AI attribution anywhere — code, comments, commits, PR descriptions, changelogs, docs: no authorship byline or trailer naming an AI tool/model, no "written by an assistant" style notes
- Run plain `git commit` or `git push` — both are blocked by permission rules; they bypass commit signing and the unified commit+push wrapper
- Let a subagent write `.git/COMMIT_MESS` or call `gitcommit` — subagents edit and report back; the parent instance owns the commit
- Install Go on the local machine or run `go` commands directly on the host — the host has no Go toolchain; ALL builds/tests use Docker (`casjaysdev/go:latest`) or Incus
- Place comments inline or below code — comments always go on the line(s) ABOVE
- Use generic, ambiguous names (`data`, `mode`, `check`, `handle`, `temp`) where an intent-revealing name is possible
- Let README.md go stale — it is the first thing users see
- Expose sensitive information (secrets, tokens, internal paths) in logs, error messages, or stack traces
- Jump between features leaving earlier ones incomplete — finish each feature fully before starting the next
- Overwrite an existing loader/config file blindly during migration — merge, never blind-replace

## CRITICAL - ALWAYS DO
- Read the relevant `AI.md` PART completely (not snippets) before implementing
- Treat `AI.md` as the single source of truth; when spec and memory disagree, spec wins
- Stop and ask when genuinely blocked, when the spec is silent/contradictory, or when a decision is irreversible/destructive
- Verify claims against the actual command output, test run, or file content before reporting success
- On first session with a project containing `AI.md`: check for `.claude/rules/`, create/update all 14 rule files if missing or stale, and load `CLAUDE.md` / `IDEA.md` per the session-init sequence
- Ensure all output reads as if written by a human senior developer; the AI acts as an extension of the user, never as a separate contributor
- Use numbered/lettered options when asking the user a question
- When stuck, state precisely what you're blocked on and why, and ask how to proceed — never silently defer work
- Commit often in small, focused commits; one commit per finding for audits/reviews, one commit for a whole feature for feature work

## KEY DECISIONS (pre-answered)
| Question | Answer | Spec Reference |
|----------|--------|----------------|
| Where do builds/tests run? | Docker `casjaysdev/go:latest` (or Incus) — never the host | PART 1, Container-Only Development |
| Who commits? | The parent/main instance only, via `gitcommit`, never a subagent, never plain `git commit`/`git push` | PART 0, Tool Access |
| Is AI.md editable? | PARTS 0-36 are fixed implementation patterns — never modify | PART 0, Prohibited Actions |
| Where do comments go? | Always on the line(s) above the code, never inline | PART 0, Code Style Rules |
| What if the spec seems wrong? | Ask the user — never silently deviate or "fix" the spec | PART 0, If Spec Seems Wrong |

## TERMINOLOGY
| Term | Meaning |
|------|---------|
| AI.md | Full implementation spec (source of truth), PARTS 0-37 |
| CLAUDE.md / .claude/CLAUDE.md | Short loader auto-read every session, points into rules/ and AI.md |
| IDEA.md | Project-specific business logic/decisions, separate from AI.md's fixed patterns |
| TODO.AI.md | AI-owned, in-progress work tracking; complete items before removing |
| Rule files (.claude/rules/*.md) | Condensed per-topic extracts of AI.md for efficient context loading |

## QUICK REFERENCE
1. Identify the task's relevant PART(s) in `AI.md`.
2. Read those PARTs completely before writing code.
3. Implement exactly as specified — no invented patterns, no unrequested extras.
4. Verify every claim (build passes, test passes, file matches) before reporting done.
5. Never attribute work to AI; commits and code read as human-authored.
6. Never run `go`/builds on the host — always Docker or Incus.

---
For complete details, see AI.md PART 0, 1
