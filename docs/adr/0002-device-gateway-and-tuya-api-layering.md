# ADR 0002: Device gateway and Tuya API layering

## Status

Accepted.

## Context

smart-bridge currently has one Tuya client that handles domain-level device operations, Tuya Cloud endpoints, authentication, request signing, response decoding, and Tuya-specific mapping.

The CLI is the only application layer today, but a future Yandex Smart Home API layer is expected. smart-bridge may also support upstream platforms beyond Tuya. Those callers should not depend on Tuya-specific DTOs, signing, endpoints, or specification caching details.

Tuya command mapping needs upstream specifications. For example, domain commands such as `power=on` and `brightness=50` must be converted to the correct Tuya function codes and vendor ranges before they can be sent. Re-reading those specifications for every command request is acceptable for the current short-lived CLI, but a future long-running service will likely need caching.

## Decision

Introduce a vendor-neutral `DeviceGateway` interface in the domain package. It represents an upstream source and controller of devices using domain types only:

- list **Devices**;
- list **Capabilities** for a known **Device**;
- send **Capability Commands** to a known **Device**.

Tuya will implement this interface with a high-level `tuya.Client`. This client is responsible for domain mapping and is the future home for Tuya specification caching.

Inside the Tuya adapter, separate low-level Tuya Cloud endpoint calls into an unexported concrete `api` type. The `api` type returns Tuya DTOs, owns Tuya authentication/token state, and uses the existing transport/signing/response decoding code. It does not return domain types and does not expose an interface until there is a second implementation or a concrete testing need.

Keep `tuya.NewClient(credentials, options...)` as the construction entry point. Options may configure the internal `api` implementation, but callers should not construct or depend on `api` directly.

## Consequences

CLI and the future Yandex Smart Home API layer can depend on `devices.DeviceGateway` instead of `tuya.Client` when they need a vendor-neutral device source.

The domain model stays free of Tuya-specific metadata. Tuya specifications and any future cache remain inside the Tuya adapter, consistent with ADR 0001.

The low-level Tuya API layer remains a concrete implementation detail. Existing `httptest`-based tests continue to exercise real request paths, URLs, request bodies, signing headers, and response envelopes.

The split adds one internal layer, but it localizes future changes: retries, token refresh behavior, and Tuya endpoint DTO changes belong near `api`, while capability mapping and specification caching belong near `tuya.Client`.

The current `tuya.Client` is sufficient for short-lived CLI command execution. A future long-running HTTP service must explicitly address concurrent access to token state and any specification cache before sharing a client instance across requests.
