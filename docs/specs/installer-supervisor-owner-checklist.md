# Owner Checklist For Installer/Supervisor Work

Use this checklist before handing the installer/supervisor task to the next agent.

The goal is to answer the product and deployment questions that the implementation agent must not guess.

## 1. External Adapter

Prepare answers for all of these:

- adapter project name
- repository location
- who owns and maintains it
- current release process
- current versioning scheme
- supported operating systems
- supported CPU architectures
- required runtime
- startup command
- shutdown behavior
- config format
- config file location expectations
- environment variable requirements
- log output model
- whether it runs as a long-lived process or on-demand
- whether one instance or multiple instances may run on one machine
- compatibility expectations with `device-service`
- whether adapter artifacts are already built for Windows and Linux
- whether adapter release artifacts are private or public

## 2. Packaging Scope

Decide and document:

- is Windows the primary target or equal priority with Linux
- which Linux distributions are in scope
- whether ARM is required or x64 only
- whether installer UX should be:
  - native desktop UI
  - local web setup wizard
  - hybrid
- whether Linux should use:
  - native packages
  - install script
  - another packaging model

## 3. Installation Behavior

Decide and document:

- default install directory on Windows
- default install directory on Linux
- config directory
- data directory
- logs directory
- license directory
- backup directory
- whether these locations must be configurable
- whether the product must run under a dedicated service account

## 4. Setup Flow

Be ready to define:

- exact first-run fields the user must fill
- which values must be auto-generated
- whether Telegram bot token is mandatory at initial install
- whether first admin creation is mandatory during setup
- whether setup may be resumed later if incomplete

## 5. Licensing

Decide and document:

- what editions exist
- what the difference is between `bound` and `free`
- whether licenses expire
- whether feature flags are needed in license payload
- how activation request files should be exchanged
- who signs licenses
- what hardware-binding strictness is acceptable
- what rehost policy should be used when hardware changes
- whether offline-only operation is mandatory after activation

## 6. Update Policy

Decide and document:

- whether updates must be fully offline
- whether in-app update is required in v1
- whether rollback is required in v1
- whether DB backup is mandatory before update
- whether supervisor self-update is required
- whether adapter updates are bundled with product updates or may be updated separately

## 7. Uninstall Policy

Decide and document:

- whether uninstall should keep data by default
- whether uninstall should keep config by default
- whether uninstall should keep license by default
- whether full wipe should be offered as a separate option
- whether audit/history retention must survive uninstall in some cases

## 8. Operations UX

Be ready to answer:

- what status information operators must see
- whether they need access to logs
- whether they need a "repair installation" action
- whether they need a "re-run setup" action
- whether they need a "generate activation request" action in UI
- whether they need manual import of `license.dat`

## 9. Security Expectations

Decide and document:

- whether bundled runtime and app files are allowed to be readable by local admins
- whether secrets should be encrypted at rest
- whether license verification failure should block all runtime processes
- whether external adapter processes should also be blocked by supervisor when license is invalid

## 10. Release And Delivery

Decide and document:

- whether final customer delivery is:
  - one installer package
  - one installer plus separate update packs
  - one installer plus separately distributed adapter packs
- whether GitHub Releases remain the source of truth for delivery artifacts
- whether supervisor should live in a separate repository
- whether a future packaging pipeline should assemble:
  - supervisor
  - bundled Node runtime
  - core app artifacts
  - adapter artifacts

## Minimum Answers The Next Agent Needs

At minimum, be ready to answer these before implementation starts:

1. What is the external adapter and how is it released?
2. Which OS targets are really supported in v1?
3. What packaging UX is expected on Windows and Linux?
4. What exact setup fields must the operator enter?
5. What is the intended offline license model?
6. What should update and uninstall do with data/config/license?

## Recommended Handoff Order

When you start the next agent, provide materials in this order:

1. `docs/specs/installer-supervisor.md`
2. `docs/specs/installer-supervisor-task-prompt.md`
3. your answers from this checklist

Without the checklist answers, the next agent should stop and ask instead of assuming.
