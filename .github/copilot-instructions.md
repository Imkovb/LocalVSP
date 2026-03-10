# Copilot Instructions for LocalVSP

## Project Overview
- Project name: **LocalVSP (Local Virtual Server Provider)**
- Purpose: Lightweight self-hosted platform to deploy Docker Compose repositories from a web UI.
- Primary backend: **Go** (`management-ui/`)
- UI approach: Server-rendered templates + HTMX patterns (`management-ui/web/templates/`)
- Infra orchestration: Docker Compose (`docker-compose.yml` at root and per test apps)

## Key Repository Areas
- `management-ui/main.go`: service entrypoint
- `management-ui/internal/handlers/`: HTTP handlers and route behavior
- `management-ui/internal/docker/`: Docker/compose interactions
- `management-ui/internal/git/`: Git clone/pull operations
- `management-ui/internal/i18n/`: localization support
- `management-ui/web/templates/`: HTML templates for dashboard/help/logs/settings
- `_testapp/`: sample deployable apps (`hello-html`, `hello-python`, `hello-world`)
- `wiki/`: operational documentation and troubleshooting

## Architecture & Behavior Expectations
- Keep the app lightweight and operationally simple (target Linux VMs and Raspberry Pi class devices).
- Prefer straightforward, maintainable logic over abstract frameworks.
- Repository deployment flow should remain:
  1) clone/pull repo
  2) run `docker compose up -d` in repo folder
  3) surface status/logs clearly in UI
- Assume Traefik labels in user repos drive external routing.

## Coding Guidelines
- Keep changes scoped to the requested feature or fix.
- Preserve existing file layout and naming conventions.
- For Go changes:
  - use idiomatic Go error handling and explicit return paths
  - avoid unnecessary dependencies
  - keep handlers thin; place command/runtime logic in `internal/docker` or `internal/git` when applicable
- For template changes:
  - keep markup consistent with existing templates
  - avoid introducing new frontend frameworks
- Do not remove or alter existing deployment/test samples in `_testapp/` unless requested.

## Validation Checklist
- Build/check Go service after backend changes.
- Verify impacted template pages still render.
- If docker-related logic changes, validate against a compose-based sample app in `_testapp/`.
- Update `wiki/` docs when behavior or setup steps change.

## Test Server Access (Provided by Project Owner)
- Keep machine-specific test server credentials out of the repository.
- Pass temporary access details through local environment variables, untracked notes, or script parameters.
- Treat any server credentials used for testing as private operational data, not project documentation.

## Security Handling Note
- Do not commit live credentials, tokens, or private host details.
- If a workflow requires secrets, prefer runtime prompts, environment variables, or ignored local files.
