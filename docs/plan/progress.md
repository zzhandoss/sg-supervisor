# Progress

## Status Legend

- `not started`
- `in progress`
- `blocked`
- `done`

## 2026-03-07

### done

- reviewed product specs in `docs/specs`
- confirmed owner decisions for setup UX, platform scope, release model, licensing baseline, and adapter policy
- inspected `school-gate` and `dahua-terminal-adapter` repositories to ground the plan in actual artifacts and runtime facts
- created project tracking structure in `docs/plan`
- scaffolded Go supervisor foundation with config, layout, manifest, licensing, activation request, and local control API
- validated the foundation with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added service catalog support and in-memory child-process supervision with license gating for core services
- added persistent update manifest import/storage flow for package metadata
- extended the control API with service lifecycle and package import endpoints
- upgraded service catalog to concrete packaged command groups for `api`, `device-service`, `bot`, `worker`, and `dahua-terminal-adapter`
- added active package selection flow on top of imported manifests
- revalidated the expanded runtime/update slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added zip update bundle import with `manifest.json` + staged `payload/` extraction
- added bundle apply flow that copies staged payload into `install/core`, `install/runtime`, and `install/adapters/*` with pre-apply backup into `backups/<packageId>/`
- extended CLI and control API for bundle import and package apply
- revalidated the bundle update slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added graceful package-apply sequencing: supervisor now snapshots running services, stops them, waits for shutdown, applies the package, and restarts previously running services
- active package status now records backup path and restart/stop summary
- revalidated the update sequencing slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added service health/readiness policy driven by configured HTTP probes and static-asset presence checks
- grounded default health probes in known product endpoints for `api`, `device-service`, and `dahua-terminal-adapter`; services without a reliable probe surface as `unknown`
- revalidated the health/readiness slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added detached Ed25519 package signature verification for update imports
- bundle contract now requires `manifest.json` plus `manifest.sig`; standalone manifest import requires sibling `<manifest>.sig`
- package signing key can be configured explicitly through `packageSigningPublicKeyBase64`, with fallback to the existing `publicKeyBase64`
- revalidated the package-signing slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `internal/servicehost` with rendered Linux `systemd` unit content and Windows PowerShell service scripts
- exposed service-host artifact generation through CLI and control API for packaging/integration workflows
- revalidated the service-host slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `install-package` flow that combines package apply with service-host artifact rendering
- added `repair` flow that re-ensures supervisor-managed state and re-renders service-host artifacts
- added `uninstall` flow with `keep-state` and `full-wipe` modes that remove only supervisor-managed directories
- exposed install/repair/uninstall through CLI and control API
- revalidated the maintenance slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added persisted setup-state storage under `config/setup-state.json`
- setup progress is now structured as `required` vs `optional` fields, with `license` as the only blocking field and `telegram-bot` as resumable optional configuration
- `status` now returns detailed setup progress alongside the existing `setupRequired` boolean
- exposed setup-state mutation through `POST /api/v1/setup/fields` and `set-setup-field`
- revalidated the setup-state slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added deterministic rollback plans for bundle apply operations so supervisor can restore replaced targets and remove newly created targets
- package apply now restores the previous active package when post-update service restart fails and a backup exists
- active-package state is restored or cleared after rollback, depending on whether a previous active package existed
- revalidated the rollback slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added platform-specific service-host execution plans for Windows and Linux install/uninstall flows
- `install-package`, `repair`, and `uninstall` now execute rendered service-host artifacts through a runner abstraction instead of only returning hints
- made Windows service install/uninstall scripts idempotent enough for repeatable repair and uninstall flows
- revalidated the service-host execution slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added persistent `updates/last-operation.json` reporting for package apply outcomes, rollback outcomes, and resulting active package state
- `status` now exposes `lastUpdate`, so rollback and failed restart outcomes remain operator-visible after the command that triggered them has exited
- revalidated the update-reporting slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added structured uninstall issue reporting for service deregistration, service stop/wait failures, and filesystem removal failures
- `uninstall` now returns partial report data alongside failure so CLI and HTTP API can surface completed work and remaining issues explicitly
- `POST /api/v1/uninstall` now returns partial uninstall data on conflict instead of dropping already completed steps behind a plain error
- revalidated the uninstall-reporting slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added internal packaging manifests for `windows` and `linux` under `runtime/packaging/<platform>/install-manifest.json`
- packaging manifests now reuse generated service-host artifacts and platform-specific install/uninstall actions instead of duplicating service integration logic
- `install-package` and `repair` now refresh packaging manifests as part of their normal flow
- revalidated the packaging-wiring slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `assemble-package` flow that builds platform staging output under `build/<platform>` from the generated install manifest
- packaging assembly now copies the supervisor binary, install root, and platform-specific service-host artifacts into a deterministic staging layout
- revalidated the packaging-assembly slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `build-distribution` flow on top of assembled staging output
- Linux distribution automation now produces a controlled-install `tar.gz` package with generated `install.sh` and `uninstall.sh`
- Windows distribution automation now generates WiX MSI inputs and attempts local MSI build when `candle.exe` and `light.exe` are available; otherwise it emits ready-to-build inputs plus a warning
- revalidated the distribution-automation slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `build-release` flow that wraps distribution output into versioned release directories under `releases/v<version>/<platform>`
- release orchestration now produces stable artifact names, `release.json`, and `SHA256SUMS.txt`
- Windows release fallback now emits a versioned `wix-inputs.zip` when a local MSI cannot be built
- revalidated the release-orchestration slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added `build-release-set` so one command produces both Windows and Linux releases for the same version
- multi-platform release orchestration now writes aggregate metadata to `releases/v<version>/release-set.json`
- revalidated the multi-platform release slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added tag-based GitHub Actions release pipeline in `.github/workflows/release.yml`
- CI now builds per-platform releases on `v*` tags, uploads workflow artifacts, and publishes a GitHub release from the aggregated outputs
- workflow dispatch is also available for manual release builds with an explicit version input
- locally validated only the code paths and workflow source shape; end-to-end publish still needs the first real GitHub Actions tag run
- revalidated the CI/release-pipeline slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`

### next

- add local web UI assets
- wire richer setup field persistence to real product config once those config contracts are defined
- add similar partial-report/error contracts for install and repair if maintenance execution needs the same operator visibility
- validate the first real GitHub Actions tag release and tighten any platform-specific pipeline gaps

## 2026-03-08

### done

- validated GitHub access through `gh` and inspected failed tag release run `22807368410`
- traced the release failure to `.gitignore` matching `internal/runtime/` via `runtime/`, which kept the whole runtime package out of git while local tests still passed
- narrowed the ignore rule to root-only `/runtime/` so `internal/runtime/*` is tracked correctly for CI and future commits
- validated the fixed pipeline with `workflow_dispatch` and found a Linux-only uninstall test bug where service deregistration failures were silently ignored
- aligned Linux service-host uninstall behavior with the existing uninstall partial-report contract by surfacing `disable-service` failures instead of swallowing them
- validated the first real tag build for `v0.1.1` through to the publish job and traced the remaining failure to `gh release` being invoked without an explicit repository in a non-checkout publish job
- validated the next real tag build for `v0.1.2` through to release asset upload and traced the remaining failure to duplicate per-platform support filenames like `SHA256SUMS.txt` and `release.json`
- updated release support asset naming so checksums and metadata are unique per platform and can coexist in one GitHub release upload
- completed the first successful end-to-end tag release with `v0.1.3`, including Linux and Windows build jobs, publish job, and a populated GitHub release with per-platform installer assets, metadata, and checksums
- disabled `actions/setup-go` caching in the release workflow so dependency-free module builds stop emitting misleading `go.sum` cache warnings
- added a supervisor-managed product-config store and generated runtime env file for operator-entered product values
- wired the optional `telegram-bot` setup field to persist `TELEGRAM_BOT_TOKEN`, regenerate `runtime/config/product.env`, and overlay the bot env in the in-memory service catalog without overwriting `services.json`
- extended the existing setup API/CLI contract with an optional `value` payload so setup state and real product config can be updated together
- added embedded local web UI assets served directly by the Go supervisor without adding frontend dependencies or a separate asset build pipeline
- shipped the first browser-facing control center shell for setup/status, license actions, service start-stop-restart, and package import/apply flows on top of the existing control API
- revalidated the local web UI slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- extended `install` and `repair` with partial-report/error contracts so browser and API operators can see what completed before a maintenance failure
- expanded the embedded control center with richer status details for directories, active package context, service probes/components, and last-update messaging
- added browser-facing maintenance flows for `install-package`, `repair`, and `uninstall`, including partial-failure report rendering
- revalidated the maintenance-reporting and richer-UI slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added a separate `product-config` contract in status and control API instead of overloading setup-state semantics
- extended generated `product.env` and in-memory runtime config with network-aware defaults for `VITE_API_BASE_URL`, `API_CORS_ALLOWED_ORIGINS`, and `DEVICE_SERVICE_CORS_ALLOWED_ORIGINS`
- added application-panel forms for preferred host selection and bot token management, while keeping setup-state focused on completion state
- changed new supervisor default listen address to `0.0.0.0:8787` for network-accessible local web control center on fresh configs
- revalidated the product-config and application-panel slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- polished the local web control center toward an application-panel UX with guided next-step messaging, less technical service cards, clearer section copy, and package-id handoff between update and maintenance flows
- reduced raw operational detail in the browser UI so the panel emphasizes application actions instead of observability-style diagnostics
- revalidated the application-panel UX polish slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- added an explicit Admin UI entry flow by deriving a network-aware Admin UI URL from the current product config and exposing it in the browser panel as a direct link
- surfaced the Admin UI URL in both application config summary and access guidance, without adding OS-level browser-launch side effects
- revalidated the Admin UI entry slice with `gofmt`, `go test ./...`, and `go build ./cmd/sg-supervisor`
- expanded maintenance result UX so install/repair/uninstall actions now show detailed path/service/hint lists from the existing report payloads instead of only a compact summary
- revalidated the maintenance-detail UX slice with `go test ./...` and `go build ./cmd/sg-supervisor`
- added separate owner-side `sg-release-panel` binary with its own file-based state, embedded local web UI, and CLI surface
- release panel now auto-generates stable license/package signing keypairs, stores owner recipe state, lists upstream repo versions through `gh`, and issues signed offline license files from direct input or activation-request payloads
- added local release workflow that downloads `school-gate`, `dahua-terminal-adapter`, and Node runtime artifacts, prepares a release workspace with a generated `config/supervisor.json`, cross-builds `sg-supervisor`, and produces local Windows/Linux installer artifacts plus aggregate release-set metadata
- validated the release-panel slice with `go test ./...` and `go build ./cmd/...`
- fixed `sg-release-panel` owner-job lifecycle so read-only `status` and other CLI/API calls no longer mark a live local-release job as interrupted
- introduced explicit owner lease handling for `serve` and `build-local-release`, with interrupted-job recovery only when a stale owner lease is detected
- changed local release policy to host-native builds only: Windows hosts build Windows installers and Linux hosts build Linux installers
- surfaced `hostPlatform` in release-panel status and updated the owner UI copy so it no longer promises unsupported cross-platform local builds
- revalidated the host-native release-panel slice with `go test ./internal/releasepanel ./internal/releasepanelhttp ./cmd/sg-release-panel`, `go build ./cmd/sg-release-panel`, and `go build ./cmd/sg-supervisor`
- trimmed Windows bundled Node runtime to a runtime-only payload by removing non-essential Node tooling files from the installer workspace
- added short Windows build-stage handling plus temporary `subst` mounting for WiX so MSI packaging no longer depends on long workspace paths
- added runtime-only filtering for install-tree materialization and WiX source generation so dev-only artifacts like `*.d.ts`, `*.map`, docs, and test content are excluded from the installer payload
- validated the Windows packaging fixes against a real owner workspace by rebuilding `v1.0.6-test` into a final `.msi` artifact at `.release-panel/releases/v1.0.6-test/windows/school-gate-installer-v1.0.6-test-windows-x64.msi`
