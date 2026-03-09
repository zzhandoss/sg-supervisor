# Backlog

## Current

- Validate the Linux local-release path on a real Linux host or VM now that the owner panel build policy is host-native.
- Decide whether the first-start experience after successful bootstrap should stay entirely manual or whether Control Panel should offer a guided "Start application" action that follows the completed bootstrap state.

## Next

- Extend the bootstrap UX in Control Panel so the operator can clearly see "delivery extracted -> bootstrap succeeded -> activate license -> manually start services" as the happy path.
- Extend the product-config contract with more owner-approved operator-safe application fields if the installed application panel scope grows.
- Add richer release-panel reporting for download/build prerequisites such as missing `gh`, missing `go`, or missing Node toolchain pieces on the client host.

## Later

- Add optional GitHub publish mode on top of the local release-panel workflow if product policy returns to repository-driven releases.
- Add stricter platform-specific fingerprint providers where justified.
- Add rollback support for update failures if product policy changes.
