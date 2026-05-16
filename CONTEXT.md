# smart-bridge

This context describes the smart home concepts used by smart-bridge.

## Language

**Device**:
A vendor-neutral representation of a smart-home device.
_Avoid_: Tuya Device, smart device

**Device Type**:
A vendor-neutral smart-home device kind such as `light`, `socket`, or `switch`.
_Avoid_: Tuya category

**Capability**:
A vendor-neutral controllable feature of a **Device**, such as power, brightness, color temperature, or operation mode.
_Avoid_: Tuya function, Tuya status, Yandex API capability

**Capability Type**:
A vendor-neutral class of control, such as `on_off`, `range`, `color`, or `mode`.
_Avoid_: Tuya type, Yandex API capability type

**Capability Instance**:
The specific thing controlled by a **Capability**, such as `power`, `brightness`, `color_temperature`, `color_temperature_level`, `color`, or `work_mode`.
_Avoid_: Tuya code, Yandex API instance

**Property**:
A vendor-neutral read-only characteristic reported by a **Device**, such as temperature, humidity, or battery level.
_Avoid_: Tuya status, Yandex API property

## Relationships

- A **Device** is not tied to Tuya-specific metadata or Yandex-specific API fields.
- A **Device** is a summary loaded from upstream device listing APIs.
- **Capabilities** and **Properties** are loaded separately for a known **Device** when additional upstream reads are needed.
- A **Device Type** is mapped from upstream platform categories and later mapped to downstream platform device types.
- A **Capability** is mapped from upstream platform functions and later mapped to downstream platform capabilities.
- A **Capability** has a **Capability Type** and a **Capability Instance**.
- Known **Range Capabilities** use vendor-neutral domain scales, not upstream platform scales. For example, `brightness` and `color_temperature_level` are represented as `0..100`.
- Unknown upstream functions are not **Capabilities** until smart-bridge understands their meaning.
- A **Property** is mapped from upstream platform statuses and later mapped to downstream platform properties.
- A value that can be changed by a command is a **Capability** state, not a **Property**.

## Example dialogue

> **Dev:** "When we list **Devices**, are they Tuya-specific?"
> **Domain expert:** "No — Tuya is only the first upstream platform; **Device** is the vendor-neutral domain term."

## Flagged ambiguities

- "Tuya device" was used for objects returned by Tuya Cloud — resolved: use **Device** for the domain object and keep Tuya-specific names inside integration code.
- "category" was used for the domain device kind — resolved: use **Device Type** in the domain model and keep upstream category codes inside integration code.
- "capability" can mean a specific downstream API object — resolved: use **Capability** for the vendor-neutral domain concept and keep platform-specific capability names inside adapter code.
- "status" can mean either a controllable state or read-only telemetry — resolved: use **Capability** for controllable features and **Property** for read-only reported characteristics.
- "is the light on?" is the state of a controllable **Capability**, not a **Property**.
