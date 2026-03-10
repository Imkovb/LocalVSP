# LocalVSP Refactor Plan

## Goal

Refactor the project in safe stages instead of attempting a single large rewrite. The current codebase is small enough to improve quickly, but several packages combine unrelated concerns and there is almost no automated verification. A big-bang refactor would create avoidable deployment risk.

## Execution Status

Status date: March 8, 2026

The original plan has now been substantially implemented in the codebase. This document is no longer just a proposal. It is the completion tracker for the refactor work that has already landed and the short list of items still open.

### Completed

- Phase 0 completed: validation workflow added with `management-ui/Makefile`, spec/install flow reconciled, and first tests added.
- Phase 1 completed: `main.go` now delegates to `handlers.NewMux()`, template rendering moved into `internal/view`, and the handler monolith was split into focused files.
- Phase 2 largely completed: `internal/docker/docker.go` was split into focused files for compose, infra, projects, config, system info, auto-generation, and shared helpers.
- Phase 4 substantially completed: dashboard and settings UX were redesigned, row partials were cleaned up, and shared CSS moved into `web/static/vsp-ui.css`.
- Installer/config hardening partially completed: `.env` writes now preserve unknown keys, file writes are more defensive, and the installer no longer uses `git reset --hard`.
- Tests partially completed: baseline unit tests and router tests now exist.

### Still Open

- Runtime verification has not been executed in this environment because Go is not available on PATH here.
- A command executor interface has not yet been introduced, so Docker-heavy code is still harder to fake in unit tests than it should be.
- Streamed and non-streamed Docker deploy flows are still not fully unified behind one deployment service.
- Default credentials and first-run secret generation are still not fully hardened.
- Some inline JavaScript remains in templates and can still be moved to shared static assets.
- CI has not yet been added.

## Current Findings

### Architecture

- `management-ui/internal/handlers/handlers.go` contains routing helpers, page rendering, HTMX actions, SSE build streaming, in-memory job management, and API responses in one file.
- `management-ui/internal/docker/docker.go` mixes command execution, Docker Compose lifecycle, project auto-detection, override generation, system metrics, autostart persistence, and platform config.
- Template loading is performed in package `init()`, which makes testing and startup behavior harder to control.
- There is repeated request-derived logic such as host normalization and language/template data assembly.

### Operational Risk

- `install.sh` updates an existing checkout with `git reset --hard`, which is destructive for local changes.
- `install.sh`, the runtime behavior, and `spec.md` are not aligned. The spec describes repo-based deployment, while the current product is centered on local folder deployment from `/home/vsp`.
- `DeployDockerProject` starts its real work in a goroutine and returns immediately, so failures are detached from the caller and behavior differs from the streamed deploy path.
- Host ports are reassigned on each deploy in the Docker project flow, which can break bookmarks and external references.
- Input validation is inconsistent. There are validation helpers in the Docker package, but not all write paths use them.
- Multiple file writes ignore returned errors.

### Maintainability

- There are no Go tests in `management-ui`.
- Command execution is not abstracted, so Docker-heavy code is difficult to unit test.
- Override file generation and compose/service parsing are string-based and fragile.
- Templates are server-rendered but still contain a large amount of repeated inline style and behavior.

### UI / UX

- The dashboard is functional but visually dense. Action controls, status, filesystem path, and networking information compete for the same visual weight.
- Tables are carrying too many responsibilities at once: status display, deploy actions, logs, autostart toggles, subdomain editing, and port visibility.
- The current styling is mostly inline Tailwind classes inside templates, which makes visual changes expensive and inconsistent.
- Primary actions are not strongly prioritized. Deploy, stop, logs, rebuild, and delete all sit in the same row with limited hierarchy.
- The networking/address column is hard to scan because subdomain editing, port display, and links share the same area.
- There is no clear empty state, onboarding path, or “next action” guidance for first-time users.
- Long-running actions rely on row-level changes and side panels, but the interaction model is not yet unified across apps, sites, and infra.

### Security / Hardening

- Default credentials are provisioned in the installer.
- Secrets and config are written by replacing the full `.env` file instead of preserving unknown keys and writing atomically.
- The management UI currently relies on trust of the local environment; there is no explicit CSRF or auth hardening strategy documented.

## Target End State

### Backend shape

- `internal/httpapp` or `internal/server` owns router construction and middleware.
- `internal/handlers/pages`, `internal/handlers/api`, and `internal/handlers/actions` separate HTTP concerns by surface area.
- `internal/deploy` owns project deploy flows and build job orchestration.
- `internal/dockerexec` or `internal/runtime` owns Docker and Compose command execution behind an interface.
- `internal/projects` owns project detection, port allocation, subdomain/autostart config, and override generation.
- `internal/platform` owns system info and platform config.
- `internal/view` owns template loading, shared view models, and render helpers.

### Quality bar

- Critical package logic covered by unit tests.
- HTTP handlers verified with `httptest`.
- Smoke checks run automatically in CI for the Go module.
- Installer and docs reflect the same product behavior as the UI.

## Refactor Strategy

## Phase 0: Stabilize Before Moving Code

Status: completed

- Add a minimal verification baseline for `management-ui`.
- Introduce a small `Makefile` or documented commands for `test`, `build`, and `fmt`.
- Add at least one smoke test for router startup and one unit test package to prove the test harness works.
- Document the current deployment model clearly and resolve the repo-vs-folder mismatch.

Implemented:

- Added `management-ui/Makefile` with `fmt`, `test`, `build`, and `verify` targets.
- Added initial Go test coverage in `internal/docker/docker_test.go` and `internal/handlers/router_test.go`.
- Updated `spec.md` to match the real folder-based deployment flow.

Exit criteria:

- A contributor can run one documented command to validate backend changes.
- The spec and installer docs describe the same deployment flow as the product.

## Phase 1: Split the HTTP Layer

Status: completed

- Create a router constructor instead of wiring routes directly in `main.go`.
- Move template loading/render helpers into a dedicated package.
- Extract shared request helpers: host normalization, language detection wrapping, HTMX response helpers.
- Split `handlers.go` into at least:
  - page handlers
  - fragment/API handlers
  - build handlers
  - response/render helpers

Implemented:

- Added `handlers.NewMux()` and reduced `main.go` to lifecycle/config wiring.
- Added `internal/view/view.go` for template loading/rendering.
- Split the old handler monolith into:
  - `common.go`
  - `pages.go`
  - `fragments.go`
  - `actions.go`
  - `build_jobs.go`
  - `router.go`

Exit criteria:

- `handlers.go` no longer exists as a monolith.
- `main.go` is reduced to configuration, dependency wiring, and server lifecycle.

## Phase 2: Split Docker and Project Logic

Status: partially completed

- Introduce an executor interface around command execution.
- Move these concerns out of `docker.go` into focused files or packages:
  - compose operations
  - infra service status/actions
  - local HTML site lifecycle
  - local Docker project lifecycle
  - config persistence
  - autostart/subdomain persistence
  - project auto-detection and file generation
  - system info collection
- Replace stringly-typed helper clusters with typed services.

Implemented:

- Split the old Docker monolith into:
  - `common.go`
  - `compose.go`
  - `system.go`
  - `infra.go`
  - `projects.go`
  - `config.go`
  - `autogen.go`
  - `files.go`
- Centralized shared file/config helpers and atomic write paths.
- Improved package-level separation so each file now has one dominant responsibility.

Remaining gap:

- The command executor abstraction is still pending.

Exit criteria:

- Docker-related code can be tested with fakes instead of shelling out in every test.
- Each file/package has one primary responsibility.

## Phase 3: Make Deployments Deterministic

Status: partially completed

- Unify streamed and non-streamed deploy behavior behind one deploy service.
- Stop using detached goroutines for core deploy execution unless explicitly managed by a job manager.
- Persist assigned host ports instead of always rotating them on redeploy.
- Validate names and subdomains at every write boundary.
- Make override generation deterministic and easier to verify with tests.
- Improve error propagation so the UI can distinguish validation errors, Docker failures, and post-start health failures.

Implemented:

- Host port persistence behavior was improved so redeploys do not rotate ports unnecessarily.
- Validation and config write handling were tightened in the Docker/config helpers.
- Build job behavior is now more explicit and isolated from the rest of the HTTP layer.

Remaining gap:

- Streamed and non-streamed deploy behavior are not yet fully unified behind one service.
- There is still room to improve typed error propagation for deploy failures.

Exit criteria:

- Deploying a project follows one code path.
- Redeploying does not unexpectedly change routing or host port behavior.

## Phase 4: Template and Frontend Cleanup

Status: partially completed

- Introduce a base layout/shared partial structure for navigation, modal shells, and repeated status components.
- Move large inline scripts and styles to static assets where practical.
- Standardize table row partials and build-panel behavior.
- Reduce duplicated Tailwind class blocks in templates.
- Introduce reusable UI primitives for badges, action buttons, form inputs, segmented toggles, and cards.
- Reduce visual noise by separating primary actions from destructive and secondary actions.
- Rework the address/subdomain editing UX so read-only state and edit state are clearly distinct.

Implemented:

- Added shared stylesheet `web/static/vsp-ui.css`.
- Reduced repeated inline CSS across dashboard, settings, help, and logs templates.
- Standardized more of the row/card/button styling through shared classes.
- Simplified the main management surfaces so status, actions, and address controls are easier to scan.

Remaining gap:

- Some inline JavaScript still remains in templates.
- Shared partial/layout extraction can still go further.

Exit criteria:

- Shared UI behavior is defined once.
- Template files are shorter and easier to review.

## Phase 4A: Dashboard UX Improvements

Status: substantially completed

- Rework the dashboard tables into clearer management surfaces with stronger hierarchy:
  - project/site identity first
  - health/status second
  - primary action area third
  - advanced controls hidden behind a compact secondary affordance
- Collapse low-signal metadata such as local filesystem path into expandable details instead of the primary row.
- Replace the current action clusters with a consistent scheme:
  - primary: deploy or redeploy
  - secondary: stop, logs
  - destructive: delete
- Convert autostart and subdomain controls into cleaner compact components instead of mixed inline buttons and ad-hoc forms.
- Improve build feedback with a clearer progress panel, sticky current status, and direct jump from row to active build output.
- Add empty states for no projects, no sites, and missing domain/cloudflare configuration.
- Improve responsive behavior so the dashboard remains usable on tablets and narrow laptops.
- Add clearer copy around URL exposure:
  - local port access
  - subdomain access
  - why a project has no web address yet

Implemented:

- Reworked dashboard surfaces to create clearer visual hierarchy.
- Improved primary vs secondary vs destructive action separation.
- Cleaned up subdomain/address presentation and row information density.
- Improved settings grouping and platform guidance.
- Added better empty-state and structure support in the row partials.

Remaining gap:

- Build-panel interaction can still be consolidated further.
- Additional responsive testing should happen in a real browser.

Exit criteria:

- The main dashboard is easier to scan in under a few seconds.
- Primary actions are obvious without making destructive actions too easy.
- Address management and build status require fewer clicks and less interpretation.

## UI Improvement Backlog

- Add a shared layout template with centralized theme tokens.
- Move inline CSS and JavaScript out of `dashboard.html`, `settings.html`, and `help.html` into static assets.
- Standardize status badges for `running`, `stopped`, `partial`, `unknown`, and `building`.
- Introduce a consistent button system for primary, neutral, warning, and destructive actions.
- Replace the current table action row with a compact action bar plus overflow menu for less common actions.
- Convert subdomain editing into explicit edit/save/cancel states.
- Show port and URL information as chips/links instead of mixed plain text.
- Add inline validation and clearer error messaging for subdomain/domain input.
- Add first-run onboarding hints for empty `docker` and `html` folders.
- Add a clearer build drawer or modal with active step, live output, and last result.
- Improve settings page grouping and explain Cloudflare/domain dependencies more clearly.
- Add visual feedback for background refreshes and row updates so HTMX changes feel intentional.

## Phase 5: Installer, Config, and Security Hardening

Status: partially completed

- Remove destructive update behavior from `install.sh`.
- Preserve unknown `.env` keys and write config atomically.
- Replace default credentials with first-run prompts or one-time generated secrets.
- Review Samba, Gitea, and management UI defaults for safer out-of-box behavior.
- Add installer idempotency checks and clearer failure handling.

Implemented:

- Removed destructive `git reset --hard` update behavior from `install.sh`.
- Preserved unknown `.env` keys during updates instead of rewriting the whole file blindly.
- Improved config file write safety.

Remaining gap:

- Default credentials still exist in installer/bootstrap flows.
- First-run generated secrets and broader default-hardening are still pending.

Exit criteria:

- Re-running the installer is safe.
- Secrets/config are managed predictably.

## Phase 6: Test Coverage and Release Hygiene

Status: partially completed

- Add table-driven tests for:
  - compose file detection
  - first-service detection
  - trigger file auto-detection
  - config read/write logic
  - subdomain/name validation
- Add handler tests for key endpoints using `httptest`.
- Add a small integration smoke path for build job status and rendered fragments.
- Add CI for `go test ./...` and formatting checks.

Implemented:

- Added baseline tests for config, parsing, validation, and router behavior.
- Added a simple verification command surface via the Makefile.

Remaining gap:

- Broader handler coverage is still needed.
- CI has not yet been added.
- Full compile/test execution is still pending in an environment with Go installed.

Exit criteria:

- Refactors can be validated without manual clicking for every change.

## Recommended Order of Work

This order has been mostly executed. The remaining recommended order is:

1. Add a command executor abstraction for Docker and Compose calls.
2. Unify streamed and non-streamed deploy paths.
3. Finish moving remaining inline JavaScript into shared static assets.
4. Replace default installer credentials with first-run generated secrets or prompts.
5. Add CI for `go test ./...` and `go build ./...`.
6. Run full browser and runtime verification in a Go-enabled environment.

## Concrete TODO

- [x] Add a backend validation workflow for build, format, and test.
- [x] Create a router constructor and remove route registration from `main.go`.
- [x] Move template parsing/rendering out of handler package `init()`.
- [x] Extract shared HTTP helpers for host, language, and HTMX responses.
- [x] Split page, fragment, action, and build handlers into separate files/packages.
- [ ] Introduce a command executor interface for Docker/Compose calls.
- [x] Split `docker.go` by responsibility.
- [ ] Unify streamed and non-streamed deploy code behind one deployment service.
- [x] Persist host ports across redeploys unless explicitly changed.
- [x] Enforce validation on project names and subdomains at every input boundary.
- [x] Make `.env` updates atomic and preserve unknown keys.
- [x] Remove `git reset --hard` from installer update flow.
- [ ] Replace default credentials/secrets with first-run setup.
- [x] Add unit tests for parsing, detection, and config logic.
- [ ] Add handler tests for dashboard, settings save, and build/log endpoints.
- [ ] Move repeated inline JavaScript to static assets or shared template partials.
- [x] Move repeated inline CSS to static assets or shared template partials.
- [x] Reconcile `spec.md`, installer behavior, and UI terminology.
- [x] Define shared UI tokens and reusable component classes for buttons, badges, forms, and cards.
- [x] Split primary and destructive actions visually in dashboard rows.
- [x] Redesign the web-address/subdomain editing flow.
- [x] Add empty states and onboarding guidance for first-time users.
- [ ] Improve build panel UX and unify it across project/site actions.
- [x] Make dashboard layout more responsive and easier to scan on smaller screens.

## Remaining Refactor Slice

If continuing from the current codebase, this is the next low-risk sequence:

1. Introduce a Docker/Compose executor interface without changing handler APIs.
2. Route both deploy entry points through one deploy service.
3. Move remaining build/help inline scripts into shared static files.
4. Expand `httptest` coverage for settings save, logs, and build endpoints.
5. Replace bootstrap default credentials with first-run prompts or generated secrets.

That would close the remaining structural gaps without requiring another broad rewrite.