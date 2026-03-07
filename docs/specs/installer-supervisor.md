# Installer And Supervisor Implementation Brief

## Purpose

This document is a handoff brief for the next agent that will design and implement the product installer and runtime supervisor for `School Gate` on Windows and Linux.

The target users are non-technical operators. Installation, activation, update, and uninstall flows must be safe and simple.

This brief is not a code-level design yet. It defines the product constraints, the recommended architecture, and the implementation direction that must be followed unless explicitly revised by the owner.

## Product Context

Current product is a multi-service Node.js system with at least these runtime parts:

- `api`
- `device-service`
- `worker`
- `bot`
- `admin-ui`

The existing monorepo is the application core. It should continue to produce release artifacts, but end users must not directly run multiple services or manage Node.js manually.

## Primary Goals

The installer/supervisor solution must solve all of these:

- simple installation on Windows and Linux
- no requirement for users to install Node.js manually
- no requirement for users to edit `.env` manually
- simple first-run setup for secrets and required configuration
- offline licensing support
- support for two license modes:
  - hardware-bound
  - free / unbound
- one-click or simple guided start/stop/status
- update flow from an offline package
- uninstall flow
- support for bundling at least one external adapter that is maintained in a separate project and released separately

## Recommended High-Level Architecture

Recommended architecture is:

1. Keep the current monorepo as the product core.
2. Build a separate native `supervisor` project.
3. Build the installer around the supervisor, not around raw Node processes.

Recommended separation:

- `school-gate-core`
  - current monorepo
  - produces application artifacts
- `school-gate-supervisor`
  - separate native project, preferably Go
  - performs license checks, config bootstrap, process supervision, service install/uninstall, update, and repair
- optional `product assembly` pipeline
  - combines core artifacts + supervisor + external adapter artifacts into final customer package

## Why Supervisor Must Be Native

Do not build the supervisor as another Node service inside the current monorepo.

Reasons:

- users must not depend on a system-installed Node runtime
- installer and runtime control should remain available even if app services fail
- license verification is stronger when it happens in a native entry point
- packaging and service registration are much simpler with a native binary
- Windows and Linux service integration is cleaner

Recommended language: `Go`.

Reasons:

- easy single-binary distribution
- good Windows and Linux support
- easy process supervision
- easy file IO, crypto, signatures, and service integration

Rust is also acceptable, but Go is the preferred recommendation for implementation speed and operational simplicity.

## Runtime Packaging Strategy

Do not require global Node.js installation on customer machines.

Instead, ship a bundled runtime inside the product package:

- Windows: bundled `node.exe`
- Linux: bundled `node`

The supervisor must launch the internal services by using the bundled runtime from the installation directory.

Users must never need to:

- install Node.js
- set `PATH`
- run `pnpm`
- open terminal to manage the product

## External Service Model

The operating system should know only one product service:

- Windows: one Windows Service, for example `SchoolGateSupervisor`
- Linux: one `systemd` service, for example `school-gate-supervisor.service`

The supervisor then starts and monitors the internal child processes:

- `api`
- `device-service`
- `worker`
- `bot`
- `admin-ui`
- external adapter processes if they are bundled into the product package

Do not use these as the primary orchestration model:

- `pm2`
- Windows Task Scheduler
- `cron`
- separate OS services for every internal component

Rationale:

- they complicate support
- they leak implementation details to operators
- they make licensing and updates harder
- they increase cross-platform differences

## Installer Responsibilities

The installer must support:

- fresh install
- first-run setup
- license activation
- service registration
- start / stop / restart
- repair
- update from package
- uninstall

The installer may be a separate binary or a mode of the supervisor binary. Either is acceptable if the resulting UX is simple.

## First-Run Setup Flow

Target first-run flow:

1. Product is installed.
2. Supervisor starts in setup-required mode if no valid configuration exists.
3. User sees a guided setup UI or local web wizard.
4. Setup collects required operator inputs.
5. Setup generates internal secrets automatically.
6. Setup stores config.
7. Setup requests or imports license.
8. Setup starts the service.
9. User opens Admin UI.

### Setup Inputs

Split inputs into two groups.

Auto-generated, not entered by user:

- JWT secrets
- cookie secrets
- internal service keys
- encryption keys
- any random installation-specific secret

Entered by user:

- first admin credentials
- Telegram bot token, if this deployment uses it
- ports only if really needed
- license file or activation response file

Do not make raw `.env` editing the default operator flow.

Advanced mode may expose `.env`-like editing later, but default UX must be form-driven.

## Configuration Storage

Recommended model:

- store config in a product-owned config directory
- store secrets in the same protected config area
- keep runtime data in a separate data directory
- keep logs in a separate logs directory

The supervisor should translate config into process environment variables internally.

The end user should think in terms of:

- settings
- activation
- service status

not in terms of `.env` files.

## Offline Licensing Requirements

The product must support offline licensing.

Two license modes are required:

1. `bound`
  - tied to machine fingerprint
  - must fail on different hardware
2. `free`
  - no hardware binding
  - same build should be installable on multiple machines

Do not implement separate builds for these two modes unless there is a later business requirement. Prefer one product build and two license modes.

### License Format

Recommended model:

- signed offline license file
- supervisor contains public key
- license file is signed with vendor private key

License should contain at least:

- `licenseId`
- `customer`
- `mode` = `bound` or `free`
- `edition`
- `features`
- `expiresAt` or perpetual flag
- hardware fingerprint for `bound` mode
- signature

### Hardware Binding

Do not bind only to MAC address.

MAC-only binding is not stable enough.

Use a hardware fingerprint composed from multiple signals, for example:

- Windows `MachineGuid`
- Linux `/etc/machine-id`
- board UUID or product UUID if available
- disk identifier if stable enough
- one physical MAC as an optional factor

Recommended policy:

- fingerprint is derived from multiple machine properties
- `bound` license validates against this fingerprint
- minor hardware drift policy may be considered later, but initial implementation can require exact match

### Offline Activation Flow

Recommended flow:

1. Product generates `activation-request.json`.
2. User sends it to vendor manually.
3. Vendor returns signed `license.dat`.
4. User imports `license.dat`.
5. Supervisor validates signature and activation rules locally.

This flow must work without internet access on the customer machine.

## Anti-Piracy Expectations

This product is based on Node.js services, so the protection is deterrence, not perfect DRM.

Recommended enforcement point:

- supervisor validates license before starting internal services

Do not rely on application-level checks inside only TypeScript/Node services as the primary protection mechanism.

If stronger protection is needed later, the supervisor may become the sole component that provides runtime secrets required for child services to start.

That is not required for initial implementation, but the architecture should not block this later.

## Updates

Updates must be supported through the installer/supervisor.

Internet must not be required for the update process itself. Offline package update is enough for the first version.

Recommended update flow:

1. user obtains update package
2. supervisor validates package version and signature if implemented
3. supervisor stops managed services
4. supervisor backs up config and data metadata
5. supervisor replaces runtime/application artifacts
6. supervisor runs required migrations
7. supervisor starts services
8. supervisor runs health checks
9. on failure, supervisor can report error and optionally roll back

### Update Package Support

The next agent should design a versioned update package format that can contain:

- supervisor update if needed
- core runtime bundle
- bundled adapter artifacts
- manifest with compatibility information

## Uninstall

Yes, uninstall is required.

Uninstall flow must support at least these modes:

- uninstall application but keep data and config
- uninstall application and remove service registration
- full uninstall including data, logs, config, and license

This is important for:

- reinstall
- support scenarios
- hardware replacement
- customer offboarding

Windows should expose normal uninstall behavior in installed applications.

Linux should expose a clear uninstall command or package removal flow, depending on packaging strategy.

## External Adapter Requirement

There is at least one important external adapter that lives in a separate project with its own release lifecycle.

This external adapter must not be assumed to be part of the current monorepo.

The installer/supervisor design must treat adapters as external versioned artifacts.

Recommended model:

- supervisor/product package includes one or more adapter bundles
- each adapter bundle has explicit version metadata
- supervisor checks compatibility before launch

The next agent must stop and request the following details from the owner before implementing adapter packaging:

- adapter repository/project identity
- how the adapter is currently released
- operating systems and architectures supported by the adapter
- how the adapter is started
- what runtime it requires
- what config it expects
- how compatibility with `device-service` should be defined

Do not guess adapter integration details.

## Release Model

Recommended release model:

### Core Product Repo

Current monorepo continues to produce:

- source release
- prebuilt application bundle
- manifest metadata if needed

### Supervisor Repo

Separate repo produces:

- `supervisor-windows-x64`
- `supervisor-linux-x64`
- installer-related binaries or service helpers

### Final Customer Package

Final installer package should assemble:

- supervisor binary
- bundled Node runtime
- core app artifacts
- selected adapter artifacts
- product manifest

This can be built by:

- a separate packaging pipeline
- or initially by the supervisor repo release pipeline

Either approach is acceptable.

## Compatibility Manifest

The next agent should define a product manifest similar to:

```json
{
  "productVersion": "1.1.0",
  "coreVersion": "1.1.0",
  "supervisorVersion": "0.1.0",
  "runtime": {
    "nodeVersion": "20.x"
  },
  "adapters": [
    {
      "key": "external-adapter-key",
      "version": "x.y.z",
      "required": true
    }
  ],
  "compatibility": {
    "coreApi": 1,
    "adapterApi": 1
  }
}
```

Exact schema is up to the next agent, but compatibility must be explicit and machine-readable.

## Windows Packaging Direction

Recommended direction:

- normal installer experience
- install directory under standard program location
- register one Windows Service for supervisor
- register uninstaller
- create shortcuts for:
  - Control Center
  - Admin UI
  - Uninstall

The next agent may choose the exact packaging tool, but it must result in a normal Windows operator experience.

## Linux Packaging Direction

Recommended direction:

- native package or controlled install script
- register one `systemd` service for supervisor
- standard config/data/log locations
- clear uninstall path

Linux support is required, but usability expectations should assume that Linux deployments are still somewhat more operationally sensitive than Windows.

The next agent must keep parity in core product behavior even if packaging UX differs by OS.

## Control Center

The product should expose a simple operator-facing control surface, either native or local web-based.

Minimum functions:

- setup status
- license status
- service status
- start / stop / restart
- update from package
- log/status view
- open Admin UI

This can be implemented in phases, but architecture should reserve a place for it from the start.

## Strong Recommendations

The next agent should follow these decisions by default:

- use a separate native supervisor project
- prefer Go
- bundle Node runtime with the product
- use one OS-level service only
- do not use `pm2` as the main runtime orchestrator
- do not require system-installed Node
- do not require manual `.env` editing as the main setup flow
- implement offline signed license files
- support both `bound` and `free` license modes in one product build
- treat external adapters as separate versioned artifacts

## Open Questions The Next Agent Must Resolve With Owner

Before implementation starts, the next agent must clarify these:

1. Which packaging tool should be used for Windows?
2. Which Linux distributions are in actual support scope?
3. Is the installer expected to provide a native desktop UI, a local web setup wizard, or both?
4. What is the exact required first-run configuration set?
5. What is the desired license file format and signature process?
6. How strict should hardware binding be?
7. What is the exact external adapter release model?
8. What compatibility contract exists or should exist between core and adapter releases?
9. Where should config, logs, and data live on Windows and Linux?
10. Is rollback on failed update required in the first implementation?

## Delivery Expectation For The Next Agent

The next agent should begin with:

1. architecture proposal for supervisor + installer
2. artifact/release model
3. config/data/license directory layout
4. setup/activation/update/uninstall flow diagrams
5. list of required owner clarifications, especially for the external adapter

The next agent should not jump directly into implementation before confirming those details.
