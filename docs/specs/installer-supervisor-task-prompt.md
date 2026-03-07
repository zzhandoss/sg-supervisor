# Task Prompt For Installer/Supervisor Agent

## Task

You are implementing the deployment foundation for `School Gate` on Windows and Linux.

Before doing any coding, read:

- `docs/specs/installer-supervisor.md`

That document is the source of truth for the expected installer/supervisor architecture.

## Goal

Design and implement the installer/supervisor foundation for a multi-service product that must be usable by non-technical operators.

The solution must support:

- Windows and Linux
- bundled Node runtime
- one OS-level service only
- simple install/setup/start/stop flow
- offline licensing
- two license modes:
  - `bound`
  - `free`
- update from offline package
- uninstall
- external adapter packaging as a separate artifact

## Important Constraints

- Do not redesign the product architecture from scratch.
- Use `docs/specs/installer-supervisor.md` as the approved direction unless the owner explicitly changes it.
- Do not assume the external adapter is part of the current monorepo.
- Do not assume `pm2` is the primary runtime orchestration model.
- Do not assume end users can edit `.env` manually.
- Do not assume Node.js is preinstalled on customer machines.

## Required Working Assumption

The current application monorepo remains the product core.

Recommended target architecture already chosen:

- separate native supervisor project
- bundled Node runtime inside the product package
- one OS-level service
- supervisor manages child processes
- signed offline license file
- same product build supports both `bound` and `free` license modes

## What You Must Do First

Before implementation, ask the owner the unresolved questions listed in:

- `docs/specs/installer-supervisor.md`

At minimum, you must explicitly ask for:

1. external adapter project details
2. how the adapter is released today
3. adapter runtime requirements
4. adapter startup model
5. adapter compatibility expectations with the core product
6. Windows packaging preference
7. Linux distro support scope
8. whether setup UI should be native, local web, or hybrid

Do not guess these.

## Expected Delivery Sequence

Work in this order:

1. Confirm owner answers for unresolved packaging/licensing/adapter questions.
2. Propose concrete repository and artifact layout.
3. Propose install directory, config directory, data directory, log directory, and license directory layout for Windows and Linux.
4. Propose setup flow.
5. Propose activation flow.
6. Propose update flow.
7. Propose uninstall flow.
8. Only after that, begin implementation.

## Expected Deliverables

You should produce, in order:

1. architecture note for supervisor/installer structure
2. artifact and release model
3. config/data/license/log directory layout
4. service management design for Windows and Linux
5. licensing design with `bound` and `free` modes
6. external adapter packaging strategy
7. implementation plan split into small phases
8. then code

## Adapter Warning

There is an important external adapter in another project with its own releases.

You must request its details from the owner before implementing adapter packaging or adapter supervision.

Treat adapter integration as:

- separate project
- separate release lifecycle
- separate artifact

Do not hardcode assumptions.

## Definition Of Success

The result should make it realistic for a non-technical operator to:

- install the product
- activate it
- start it
- update it
- uninstall it

without manually installing Node.js or managing multiple services.
