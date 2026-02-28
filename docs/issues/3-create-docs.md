---
type: chore
tags: ["chore"]
---

# Create Docs

## Context

Create a maintainable product documentation site for the CLI using Astro + Starlight, with examples for common automation workflows and a structure that can scale across versions.

## Implementation Plan

1. Confirm scope and product naming (`akv` vs references to `cast`) and update terminology so the docs site reflects the actual CLI and command set.
2. Scaffold an Astro + Starlight documentation site under `docs/` with version-ready structure (`docs/src/content/docs/<version>/...`) and shared navigation config.
3. Migrate existing markdown docs into the new content model and define core sections:
   - getting started,
   - installation/authentication,
   - configuration profiles and vault aliases,
   - command reference for `secrets`, `keys`, `certificates`.
4. Add task-oriented guides with realistic examples (including YAML/CLI snippets) for:
   - secret lifecycle workflows,
   - remote modules/tasks concepts (if still in scope),
   - docker and deno automation examples,
   - operational playbooks for build/deploy/ETL scenarios.
5. Enable and validate built-in docs UX features:
   - local full-text search,
   - RSS feed generation,
   - sitemap/SEO metadata,
   - version switcher strategy for future releases.
6. Add docs quality checks in CI:
   - markdown/frontmatter validation,
   - link checking,
   - `astro build` verification.
7. Set up deployment to Cloudflare Pages (or equivalent static host) with preview environments for pull requests and automatic production deploys from main.
8. Document docs contribution workflow so new commands/features can be added with consistent templates and versioning rules.

## Acceptance Criteria

- Docs site is scaffolded in `docs/` with version-ready structure and navigation.
- Command reference and task-based guides include practical examples and copy-paste-ready snippets.
- Search, RSS, and static build outputs are enabled and validated.
- Hosted deployment is configured with preview and production environments.
