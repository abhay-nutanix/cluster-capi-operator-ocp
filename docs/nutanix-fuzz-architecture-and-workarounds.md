## Nutanix MAPI ↔ CAPI fuzz tests: architectural differences and workarounds

### Overview

This document explains why Nutanix round‑trip fuzz tests between Machine API (MAPI) and Cluster API (CAPI) can fail despite functionally correct conversions. The failures stem from fundamental architectural differences between the APIs and how our harness validates round‑trips.

It also provides practical workarounds to keep tests meaningful while acknowledging lossy areas, and offers guidance for safely re‑enabling or tuning fuzz suites.

### TL;DR

- **Provider‑level metadata does not exist in CAPI** → MAPI providerSpec metadata is lost on MAPI→CAPI and cannot round‑trip.
- **Empty slice vs nil serialization differs** → JSON `[]` vs `null` breaks strict round‑trip comparisons if not normalized.
- **Cluster labels are handled differently** → CAPI injects/relies on cluster labels that may not perfectly match MAPI after a round‑trip.

These behaviors are architectural, not Nutanix conversion bugs.

---

## 1) Provider spec metadata cannot be preserved

### What happens

- MAPI allows `ObjectMeta` on the provider spec (e.g., `mapiv1.NutanixMachineProviderConfig.ObjectMeta`).
- CAPI Nutanix provider types do not support provider‑spec‑level metadata; metadata lives on the parent objects instead.
- During MAPI→CAPI, provider‑level labels/annotations have no equivalent target and are dropped.

### Impact on fuzz tests

- If fuzzers populate providerSpec metadata, strict round‑trip comparisons fail because the metadata disappears on the way to CAPI and cannot be reconstructed on return to MAPI.

### Workarounds

- In fuzzers for MAPI→CAPI tests, explicitly clear providerSpec metadata:

```go
providerSpec.ObjectMeta = metav1.ObjectMeta{}
```

- Treat providerSpec metadata as intentionally lossy in tests. If needed, adjust the comparison to ignore providerSpec metadata fields when asserting equality.

---

## 2) JSON `nil` vs empty slices (`null` vs `[]`) break strict equality

### What happens

- Some collections in MAPI/CAPI specs (e.g., `Categories`, `GPUs`, `DataDisks`) may be set to either `nil` or empty slices.
- JSON serialization differs: `nil` becomes `null`, empty slice becomes `[]`.
- Our harness performs JSON‑based structural comparisons; `null` and `[]` are not equal.

### Impact on fuzz tests

- Fuzzers that randomly produce empty slices cause flapping comparisons when the conversion code normalizes to `nil` (or vice‑versa), resulting in mismatches.

### Workarounds

- Normalize collections in fuzzers to a single convention (recommend `nil` for stability):

```go
providerSpec.Categories = nil
providerSpec.GPUs = nil
providerSpec.DataDisks = nil
```

- Optionally, normalize in conversion functions (when safe) to produce consistent `nil`/empty behavior for round‑trip stability.

---

## 3) Cluster label handling diverges between MAPI and CAPI

### What happens

- CAPI relies on and injects cluster identity via labels (e.g., `cluster.x-k8s.io/cluster-name`).
- During conversions, label/annotation merging and defaulting behavior differ across MAPI and CAPI. Some labels can be added in one direction and not fully reconstructable in the reverse direction.

### Impact on fuzz tests

- Strict equality on object metadata may fail if cluster labels present in CAPI do not map 1:1 back to MAPI or if the merger logic is not symmetric.
- MachineSet template references also gain a hash suffix by design. Our harness already tolerates the suffix and resets the name before final equality.

### Workarounds

- Seed cluster labels consistently in the fuzzer (the shared harness already forces cluster labels on CAPI objects).
- Limit comparisons to meaningful invariants for round‑trip (e.g., ignore known ephemeral metadata or compare after normalization). The harness already isolates providerSpec Raw JSON and handles the template name hash.

---

## Additional Nutanix‑specific notes

- `CredentialsSecret`: In MAPI it’s machine‑level; in CAPI it’s cluster‑level (NutanixCluster). Do not attempt strict round‑trip for this field; tests should either ignore it or set it to `nil` in MAPI fuzzers.
- Storage identifiers: Prefer UUID identifiers in fuzzers to avoid lossy conversions where Names are unsupported.
- GPU DeviceID: Type differences (CAPI int64 ↔ MAPI int32) are handled in conversions; ensure fuzzers set valid ranges and avoid unnecessary diffs by not fuzzing boundary/overflow values.

---

## Practical testing guidance

### Keep fuzzers realistic and stable

- Start from a minimal valid provider spec and selectively add variability.
- Avoid generating providerSpec metadata and credentials at the machine level.
- Normalize optional collections to `nil` (not `[]`).
- Prefer UUID‑typed resource identifiers for disks, images, subnets, containers.

### If you need strict round‑trip equality

- Use comparison transforms to ignore known lossy fields (providerSpec metadata, credentials, ephemeral labels, name hash suffixes already handled in the harness).
- Normalize object meta prior to equality (e.g., clear finalizers, ensure maps are `nil` when empty). The harness already does many of these.

### Re‑enabling Nutanix fuzz tests without flakiness

1) Ensure the Nutanix MAPI providerSpec fuzzer:

```go
providerSpec.ObjectMeta = metav1.ObjectMeta{}
providerSpec.CredentialsSecret = nil
providerSpec.Categories = nil
providerSpec.GPUs = nil
providerSpec.DataDisks = nil
```

2) Ensure required identifiers use UUID where supported and set valid sizes for required fields (memory, system disk, vCPU).

3) Keep optional fields like `Project` and `Image` valid but simple (single identifier by UUID or Name, not both).

4) Avoid fuzzing fields that are architecturally lossy from the outset (e.g., providerSpec metadata), or relax comparisons for those fields.

---

## FAQ

- **Q: Can we make providerSpec metadata round‑trip perfectly?**
  - **A:** No. CAPI does not have providerSpec‑level metadata. This is an architectural difference.

- **Q: Can we force JSON `[]` to equal `null` in comparisons?**
  - **A:** We should instead normalize inputs/outputs to a single convention (prefer `nil`) to keep behavior deterministic and unambiguous.

- **Q: Why does CAPI add a hash to the MachineSet infra template name?**
  - **A:** This is intentional to reflect template content changes. The harness already tolerates this by resetting the name before equality.

- **Q: Are these issues Nutanix‑specific?**
  - **A:** No. They are systemic differences between MAPI and CAPI. Nutanix just exposes them clearly in round‑trip fuzz tests.

---

## References

- Nutanix conversions: `pkg/conversion/{mapi2capi,capi2mapi}/nutanix.go`
- Fuzz harness utilities: `pkg/conversion/test/fuzz/fuzz.go`
- Nutanix fuzzer: `pkg/conversion/mapi2capi/nutanix_fuzz_test.go`


