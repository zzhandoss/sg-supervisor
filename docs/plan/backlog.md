# Backlog

## Current

- Validate the Linux local-release path on a real Linux host or VM now that the owner panel build policy is host-native.
- Validate the fresh delivery-based Windows local-release flow end-to-end with the new `bootstrap installer + local payload bundle` contract, including MSI auto-apply from `payload\`, and record timing baselines.

## Next

- Decide whether Linux install should mirror Windows by auto-applying the local payload bundle from the extracted delivery package without a manual panel step.
- Extend the product-config contract with more owner-approved operator-safe application fields if the installed application panel scope grows.
- Add richer release-panel reporting for download/build prerequisites such as missing `gh`, missing `go`, or missing WiX.

## Later

- Add optional GitHub publish mode on top of the local release-panel workflow if product policy returns to repository-driven releases.
- Add stricter platform-specific fingerprint providers where justified.
- Add rollback support for update failures if product policy changes.
