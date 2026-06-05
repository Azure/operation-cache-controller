## Why

`operation-cache-controller.azure.github.com/cache-mode` is written to cached `Operation`s, but the baseline specs do not say whether cluster operators may rely on it. This change clarifies that observable annotation contract before future cache changes drift around an undocumented behavior.

## What Changes

- Document the cache-mode annotation on `Operation`s created for a `Cache`.
- Specify the expected value and when the annotation is applied.
- Add focused coverage that cached `Operation`s carry the annotation.
- Do not change cache acquisition, cache key derivation, or pool sizing behavior.

## Capabilities

### New Capabilities

<!-- None. -->

### Modified Capabilities

- `cache-pool`: Add the externally observable cache-mode annotation to the cached `Operation` creation contract.

## Impact

- Affected code: `internal/handler/cache.go`, `internal/utils/controller/const.go`, and related cache tests.
- Affected APIs: no CRD schema changes; this documents an annotation already emitted on child `Operation`s.
- Affected systems: dashboards, scripts, or GitOps checks can treat the annotation as an intentional signal after this lands.
- Risk: low; the annotation already exists, so the main risk is documenting a behavior that maintainers later decide should remain internal.
