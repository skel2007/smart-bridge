# ADR 0003: Yandex Smart Home API layer

## Status

Accepted.

## Context

The CLI can discover Tuya devices, inspect capabilities, and send basic commands. smart-bridge also exposes the same known devices through the Yandex Smart Home API.

smart-bridge is a **Personal Bridge**: one running instance represents one configured smart-home environment for one user. It is not a public multi-tenant service.

Yandex has its own device types, capability types, request/response envelopes, `request_id` handling, action results, and error codes. These concepts stay at the Yandex boundary and must not leak into the domain model or **Upstream Platform** adapters such as Tuya.

The domain model uses normalized capability state. This makes `brightness` straightforward, but creates explicit mapping choices for Yandex-only shapes: `color_temperature_level` is `0..100` while Yandex `temperature_k` expects Kelvin, and raw `work_mode` is not a stable user-facing Yandex mode or scene.

## Decision

Use package `internal/yandex` for the Yandex Smart Home API layer.
Use `internal/server` for HTTP service wiring: config loading and validation, Tuya gateway construction, Yandex handler construction, and route mounting.
The `cmd/smart-bridge-server` entrypoint stays thin and calls the server package.
Keep integration packages flat for now: `internal/tuya` and `internal/yandex`; their roles are described by **Upstream Platform** and **Downstream Platform**, not encoded in nested paths.

Use the Yandex REST protocol, not JSON-RPC. JSON-RPC is only supported via Yandex Cloud Functions, while smart-bridge is expected to run as a regular Go HTTP service.

The router covers the REST surface:

- `HEAD /v1.0/`
- `POST /v1.0/user/unlink`
- `GET /v1.0/user/devices`
- `POST /v1.0/user/devices/query`
- `POST /v1.0/user/devices/action`

All Yandex handlers require `X-Request-Id` and copy it into response `request_id`; missing `X-Request-Id` is HTTP 400.

`internal/yandex` depends on `devices.DeviceGateway`, not on `tuya.Gateway`. It owns Yandex DTOs and Yandex-specific mapping rules. DTOs stay close to the JSON boundary shape and are not a second domain model. Inside package `internal/yandex`, do not repeat the `Yandex` prefix in DTO names. Use separate capability DTOs for the different Yandex shapes: `CapabilityDescription`, `CapabilityState`, `CapabilityAction`, and `CapabilityActionResult`. Response DTO values may use `any`; request DTO action values stay as `json.RawMessage` until mapping interprets them by capability `type` and `instance`. Support `custom_data` in DTOs for protocol compatibility, but current mapping ignores it.

Expose these Yandex mappings:

- `light` -> `devices.types.light`
- `socket` -> `devices.types.socket`
- `switch` -> `devices.types.switch`
- `other` -> `devices.types.other`
- `power` -> `devices.capabilities.on_off`
- `brightness` -> `devices.capabilities.range`, `instance=brightness`, `unit=unit.percent`
- `color` -> `devices.capabilities.color_setting` with HSV
- `color_temperature_level` -> `devices.capabilities.color_setting` with approximate `temperature_k`

Map `color_temperature_level` to Yandex Kelvin with a compatibility range of `2700..6500K`:

- domain `0` -> `2700K`
- domain `100` -> `6500K`
- query response: `kelvin = 2700 + level/100*(6500-2700)`
- action request: `level = (kelvin-2700)/(6500-2700)*100`

This is a downstream compatibility mapping, not a claim that the upstream device reports real Kelvin values. If a device has both `color` and `color_temperature_level`, expose one Yandex `devices.capabilities.color_setting` capability with both HSV and `temperature_k` parameters. The Kelvin range may be adjusted later after observing real device behavior.

Do not expose Yandex properties yet. The domain model has the **Property** term, but no property model or upstream property mapping yet. Also do not expose raw `work_mode`; it needs an explicit translation table before it can become Yandex scenes or modes.

Set capability `retrievable=true` and `reportable=false`. Set device-level `status_info.reportable=false`. `reportable=true` requires a state notification mechanism, such as upstream push events or polling with stored last-known state and Yandex notification delivery; a current `Device.Online` value is not enough.

Do not hide offline devices from `/devices` if the upstream device list still includes them and their capability description can be loaded. If a single device's capability description cannot be loaded for `/devices`, omit that device from the discovery response and continue returning other devices; Yandex Smart Home leaves unavailable devices unchanged and only the user can delete them. If current state cannot be read, `/query` returns a device-level error for that device and continues returning other devices. If `/action` mapping or upstream sending fails for a device, mark all actions for that device as failed. Do not attempt partial success inside one device.

Use this minimal error policy:

- missing or invalid `Authorization` -> HTTP 401
- missing `X-Request-Id`, invalid JSON, or malformed request -> HTTP 400
- unknown route -> HTTP 404
- unexpected internal bug -> HTTP 500
- upstream read/send failure -> Yandex `DEVICE_UNREACHABLE`
- requested device not found -> Yandex `DEVICE_NOT_FOUND`
- invalid action value -> Yandex `INVALID_VALUE`
- unsupported capability/action -> Yandex `NOT_SUPPORTED_IN_CURRENT_MODE`

The Yandex layer is single-user. It uses flat runtime and platform config:

```yaml
http:
  port: 0
tuya:
  # ...
yandex:
  user_id: your-stable-user-id
  bearer_token: ...
  path_prefix: /api/yandex
```

`http.port` is local-only; `0` means an OS-assigned port.
Domain name, TLS, DNS, and reverse proxy or load balancer routing are outside smart-bridge.

`yandex.user_id` is the bridge-side user ID returned to Yandex for this **Personal Bridge**, not a Yandex account ID.
`yandex.path_prefix` is the local mount path before the Yandex REST protocol paths;
the Yandex handler still owns `/v1.0/...` internally.

The shared config loader may read all sections, but validation stays caller-specific.
The HTTP server validates `http`, `tuya`, and `yandex` before listening.
The Yandex layer uses the preconfigured bearer token for protocol requests. OAuth endpoint compatibility was added later by ADR 0006.
`POST /v1.0/user/unlink` acknowledges unlink notifications without deleting local config or upstream credentials.

`GET /v1.0/user/devices` may call `ListCapabilities` per device to describe Yandex capabilities accurately. Repeated upstream specification reads are handled by the opt-in Tuya specification cache described in ADR 0001 and ADR 0002.

## Open Questions

Yandex request-level timeouts are intentionally undecided.
Choose an outer timeout only after aligning it with Tuya per-call timeouts, retries, and observed live latency.

## Implementation History

The initial Yandex code slice was split into reviewable commits:

1. DTOs only: add `internal/yandex/doc.go`, `internal/yandex/dto.go`, and `internal/yandex/dto_test.go`. Cover all REST endpoints and include minimal JSON shape tests for representative responses and requests. Do not include domain mapping, HTTP handlers, config changes, auth middleware, or a server entrypoint.
2. Pure mapping.
3. HTTP handlers with a fake `devices.DeviceGateway`.

Before exposing the endpoint outside local development, smart-bridge must be deployed behind HTTPS and configured with a strong bearer token.

## References

- [Yandex Smart Home operating protocol](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/resources)
- [Information about user devices](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/get-devices)
- [Information about the states of user devices](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/post-devices-query)
- [Change device state](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/post-action)
- [Notification of unlinked accounts](https://yandex.ru/dev/dialogs/smart-home/doc/en/reference/unlink)
- [About capabilities](https://yandex.ru/dev/dialogs/smart-home/doc/en/concepts/capability-types)
- [Mode](https://yandex.ru/dev/dialogs/smart-home/doc/en/concepts/mode)
- [Color_setting](https://yandex.ru/dev/dialogs/smart-home/doc/en/concepts/color_setting)
