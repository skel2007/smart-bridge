# ADR 0002: Device gateway and Tuya API layering

## Status

Accepted.

## Context

smart-bridge currently has one Tuya gateway that handles domain-level device operations, Tuya Cloud endpoints, authentication, request signing, response decoding, and Tuya-specific mapping.

The CLI is the only application layer today, but a future Yandex Smart Home API layer is expected. smart-bridge may also support upstream platforms beyond Tuya. Those callers should not depend on Tuya-specific DTOs, signing, endpoints, or specification caching details.

Tuya command mapping needs upstream specifications. For example, domain commands such as `power=on` and `brightness=50` must be converted to the correct Tuya function codes and vendor ranges before they can be sent. Re-reading those specifications for every command request is acceptable for the current short-lived CLI, but a future long-running service will likely need caching.

## Decision

Introduce a vendor-neutral `DeviceGateway` interface in the domain package. It represents an upstream source and controller of devices using domain types only:

- list **Devices**;
- list **Capabilities** for a known **Device**;
- send **Capability Commands** to a known **Device**.

Tuya will implement this interface with a high-level `tuya.Gateway`. This gateway is responsible for domain mapping and is the future home for Tuya specification caching.

Inside the Tuya adapter, separate low-level Tuya Cloud endpoint calls into an unexported concrete `api` type. The `api` type returns Tuya DTOs, owns Tuya authentication/token state, and uses the existing transport/signing/response decoding code. It does not return domain types.

Use an unexported `tuyaAPI` interface between `tuya.Gateway` and `api`. This keeps gateway tests focused on domain mapping and error propagation without exercising HTTP transport details.

Keep `tuya.NewGateway(credentials)` as the construction entry point. Callers should not construct or depend on `api` directly.

## Consequences

CLI and the future Yandex Smart Home API layer can depend on `devices.DeviceGateway` instead of `tuya.Gateway` when they need a vendor-neutral device source.

The domain model stays free of Tuya-specific metadata. Tuya specifications and any future cache remain inside the Tuya adapter, consistent with ADR 0001.

The low-level Tuya API layer remains an implementation detail. `httptest`-based tests exercise real request paths, URLs, request bodies, signing headers, and response envelopes at the `api`/transport level. Gateway tests use a fake `tuyaAPI`.

The split adds one internal layer, but it localizes future changes: retries, token refresh behavior, and Tuya endpoint DTO changes belong near `api`, while capability mapping and specification caching belong near `tuya.Gateway`.

The current `tuya.Gateway` is sufficient for short-lived CLI command execution. A future long-running HTTP service must explicitly address concurrent access to token state and any specification cache before sharing a gateway instance across requests.
