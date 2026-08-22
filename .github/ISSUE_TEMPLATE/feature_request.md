---
name: Feature request
about: Propose a new capability or an improvement to casci
title: "[feature] "
labels: ["enhancement"]
assignees: []
---

<!-- Check IDEA.md before filing — the feature may already be specified and pending. -->

## Problem to Solve

What problem are you hitting today? Describe the situation, not the solution.

## Desired Behavior

What should casci do instead? Be concrete about the user-visible behavior — routes,
flags, configuration keys, UI, or CLI output.

## Use Case and Target User

Who needs this and in what workflow? For example: a self-hosted operator running
casci for a single team, an administrator managing many projects, a CI consumer
calling the API, or an agent running builds on a remote node.

## Constraints and Non-Goals

What this feature explicitly should NOT do, and any limits it must respect
(no new external service dependencies, no CGO, must work with zero configuration on
first run, must work offline, and so on).

## Compatibility Expectations

casci aims for **feature parity** with the systems it replaces. Route and API parity
is only pursued when explicitly requested.

- [ ] Feature parity is enough — behavior should be equivalent, paths may differ
- [ ] Route/API parity is required — specify the exact external paths or API shape
      that must be matched and why

Does this change any existing behavior, default, route, config key, or database
schema?

- [ ] No — purely additive
- [ ] Yes — describe the breaking change and the migration path:

## Alternatives Considered

Other approaches you evaluated and why they fall short.

## Additional Context

Links, prior art, upstream documentation, or examples from other systems.
