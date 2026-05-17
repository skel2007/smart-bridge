# ADR 0001: Domain capabilities and vendor specifications

## Status

Accepted.

## Context

smart-bridge discovers smart-home devices from upstream platforms such as Tuya Cloud and will later expose or control them through downstream integrations.

Upstream platforms use their own device function codes, value ranges, and command payloads. For example, Tuya lights can represent brightness as an integer range such as `10..1000`, while the smart-bridge domain model wants a vendor-neutral `brightness` capability.

The same mapping problem exists in both directions:

- Read path: upstream status values must become domain capability state.
- Write path: future domain commands must become upstream command payloads.

## Decision

Domain capabilities use vendor-neutral names and scales. Known range capabilities such as `brightness` and `color_temperature_level` are represented as `0..100` in the domain model.

Read-side range values are rounded to the domain capability precision after adapter normalization, so floating-point artifacts from upstream ranges do not leak into domain output.

Domain capability objects do not store upstream command metadata such as Tuya function code, value range, scale, or step. That metadata belongs to the upstream adapter.

Domain capability commands are also vendor-neutral. A command targets one known capability instance and carries typed desired state, but it does not carry upstream command metadata or raw platform payloads.

Upstream adapters are responsible for both read and write conversion:

- on read, map upstream specifications and status into normalized domain capabilities;
- on write, use upstream specifications to convert normalized domain commands back into platform-specific commands.

Adapters may cache upstream specifications. The cache belongs in the adapter or application service layer, not in the domain capability model.

## Consequences

The domain model stays vendor-neutral and can be reused by CLI, future HTTP API, and downstream integrations.

Write operations must not try to reverse-map from a domain capability alone. They need access to the relevant upstream specification or a cached adapter-specific command descriptor.

Command validation can check domain invariants such as known capability instances, payload shape, and normalized value ranges. Adapter-specific validation still needs the upstream specification, for example to check allowed modes or convert normalized ranges to device ranges.

For a short-lived CLI process, reading specifications during a command is acceptable. For a future long-running HTTP service, caching specifications in the adapter/service layer should avoid repeated upstream API calls.

If an upstream device specification changes, cached specifications must eventually refresh. The initial implementation can use process-lifetime caching or no caching; a future HTTP service should choose an explicit TTL or invalidation strategy.
