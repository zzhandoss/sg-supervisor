# Backlog

## Current

- Validate the Linux local-release path on a real Linux host or VM now that the owner panel build policy is host-native.
- Finish the Windows local-release path end-to-end and confirm the final `.msi` artifact plus owner-facing output paths.

## Next

- Harden the local release workflow against malformed upstream `school-gate` prebuilt bundles and surface clearer owner-facing remediation guidance.
- Extend the product-config contract with more owner-approved operator-safe application fields if the installed application panel scope grows.
- Add richer release-panel reporting for download/build prerequisites such as missing `gh`, missing `go`, or missing WiX.

## Later

- Add optional GitHub publish mode on top of the local release-panel workflow if product policy returns to repository-driven releases.
- Add stricter platform-specific fingerprint providers where justified.
- Add rollback support for update failures if product policy changes.
