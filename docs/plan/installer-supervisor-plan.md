# Installer/Supervisor Foundation

## Purpose

This workspace defines the initial implementation foundation for the `School Gate` installer and native supervisor.

The supervisor is responsible for:

- install/update/uninstall orchestration
- offline activation and license validation
- one-service process supervision model
- local web control center
- artifact compatibility checks for core and adapter bundles

## Current implementation scope

Implemented in this repository:

- Go supervisor module scaffold
- directory layout bootstrap
- machine-readable manifest schema
- activation request and signed license schema
- local fingerprint generation foundation
- local web control API for status, activation request generation, and license import
- service catalog with packaged core/adapter command groups and runtime gating
- imported package manifest storage and active-package selection
- zip bundle import with staged extraction and install-directory replacement foundation
- graceful update sequencing that stops running managed services before apply and restarts them afterwards
- health/readiness policy for managed services with explicit `ready`, `not_ready`, and `unknown` states
- package signature verification for manifest and bundle imports using detached `manifest.sig`
- OS service-host artifact generation for `systemd` and Windows Service installation scripts
- install/repair/uninstall maintenance flows wired through supervisor CLI and control API
- persisted setup-state storage for required vs optional setup fields, surfaced through status and control API
- deterministic update rollback planning and restore flow when post-apply service restart fails
- execution of generated service-host artifacts during install, repair, and uninstall flows
- persistent update-operation reporting for apply success, rollback success, and rollback failure states
- structured partial uninstall reporting for deregistration failures and incomplete maintenance outcomes
- internal platform install manifests for Windows and Linux packaging workflows
- platform staging package assembly under `build/<platform>`
- Linux controlled distribution package generation and Windows WiX/MSI pipeline generation
- versioned release packaging with checksums and release metadata
- multi-platform release orchestration for one-version Windows+Linux outputs
- GitHub Actions release pipeline for tag-driven build and publish
- embedded local web control center assets served by the supervisor itself
- first browser-facing control center pages for setup/status, service control, and package actions
- browser-facing maintenance pages for install, repair, and uninstall flows
- partial-report handling for install, repair, and uninstall failures in the control API and local web UI
- separate product-config API/UX for operator-safe application settings such as preferred host and bot token
- network-aware derived application config for `VITE_API_BASE_URL` and default CORS origin lists
- separate owner-only `sg-release-panel` binary with file-based state, embedded UI, local installer release workflow, and offline license issuance
- project tracking documents

Not implemented yet:

- Windows MSI release orchestration beyond local WiX invocation
- trust-chain hardening beyond local detached-signature verification
- binary extraction/replacement for updates
- richer operator-facing rollback/recovery reporting around maintenance operations
- stronger upstream artifact validation and owner guidance around malformed prebuilt bundles

## Directory model

Supervisor root uses the following managed directories:

- `install/`
- `config/`
- `data/`
- `logs/`
- `licenses/`
- `backups/`
- `runtime/`
- `updates/`

## Key decisions

- `local web` is the default setup and control experience.
- First-run blocks only on license activation.
- First admin creation stays inside `school-gate` application bootstrap.
- Adapter compatibility is enforced with machine-readable manifest metadata.
- Adapter updates may be delivered through installer-only updates without replacing the core bundle.
- Invalid license blocks core runtime but does not need to block adapter startup.

## Implementation sequence

1. Foundation: docs, config, layout, manifest, licensing, control API
2. Process orchestration and health model
3. Update/import flows and compatibility validation
4. OS service integration
5. Packaging, uninstall, repair, and release automation
