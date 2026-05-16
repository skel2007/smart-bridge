# smart-bridge

This context describes the smart home concepts used by smart-bridge.

## Language

**Device**:
A vendor-neutral representation of a smart-home device.
_Avoid_: Tuya Device, smart device

## Relationships

- A **Device** is not tied to Tuya-specific metadata or Yandex-specific API fields.

## Example dialogue

> **Dev:** "When we list **Devices**, are they Tuya-specific?"
> **Domain expert:** "No — Tuya is only the first upstream platform; **Device** is the vendor-neutral domain term."

## Flagged ambiguities

- "Tuya device" was used for objects returned by Tuya Cloud — resolved: use **Device** for the domain object and keep Tuya-specific names inside integration code.
