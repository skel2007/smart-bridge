package yandex

import "encoding/json"

// HEAD /v1.0/ has no JSON request or response body.

type RequestIDResponse struct {
	RequestID string `json:"request_id"`
}

type DevicesResponse struct {
	RequestID string         `json:"request_id"`
	Payload   DevicesPayload `json:"payload"`
}

type DevicesPayload struct {
	UserID  string              `json:"user_id"`
	Devices []DeviceDescription `json:"devices"`
}

type DeviceDescription struct {
	ID           string                  `json:"id"`
	Name         string                  `json:"name"`
	StatusInfo   StatusInfo              `json:"status_info"`
	Description  string                  `json:"description,omitempty"`
	Room         string                  `json:"room,omitempty"`
	Type         string                  `json:"type"`
	CustomData   map[string]any          `json:"custom_data,omitempty"`
	Capabilities []CapabilityDescription `json:"capabilities,omitempty"`
	DeviceInfo   *DeviceInfo             `json:"device_info,omitempty"`
}

type StatusInfo struct {
	Reportable bool `json:"reportable"`
}

type DeviceInfo struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	HWVersion    string `json:"hw_version,omitempty"`
	SWVersion    string `json:"sw_version,omitempty"`
}

type CapabilityDescription struct {
	Type        string `json:"type"`
	Retrievable bool   `json:"retrievable"`
	Reportable  bool   `json:"reportable"`
	Parameters  any    `json:"parameters,omitempty"`
}

type RangeParameters struct {
	Instance     string      `json:"instance"`
	Unit         string      `json:"unit,omitempty"`
	RandomAccess *bool       `json:"random_access,omitempty"`
	Range        *ValueRange `json:"range,omitempty"`
}

type ValueRange struct {
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
	Precision float64 `json:"precision"`
}

type ColorSettingParameters struct {
	ColorModel  string             `json:"color_model,omitempty"`
	Temperature *TemperatureKRange `json:"temperature_k,omitempty"`
}

type TemperatureKRange struct {
	Min int `json:"min"`
	Max int `json:"max"`
}

type DevicesQueryRequest struct {
	Devices []DeviceRequest `json:"devices"`
}

type DeviceRequest struct {
	ID         string         `json:"id"`
	CustomData map[string]any `json:"custom_data,omitempty"`
}

type DevicesQueryResponse struct {
	RequestID string              `json:"request_id"`
	Payload   DevicesQueryPayload `json:"payload"`
}

type DevicesQueryPayload struct {
	Devices []DeviceState `json:"devices"`
}

type DeviceState struct {
	ID           string            `json:"id"`
	Capabilities []CapabilityState `json:"capabilities,omitempty"`
	ErrorCode    string            `json:"error_code,omitempty"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

type CapabilityState struct {
	Type  string               `json:"type"`
	State CapabilityStateValue `json:"state"`
}

type CapabilityStateValue struct {
	Instance string `json:"instance"`
	Value    any    `json:"value"`
}

type HSVValue struct {
	H int `json:"h"`
	S int `json:"s"`
	V int `json:"v"`
}

type DevicesActionRequest struct {
	Payload DevicesActionPayload `json:"payload"`
}

type DevicesActionPayload struct {
	Devices []DeviceAction `json:"devices"`
}

type DeviceAction struct {
	ID           string             `json:"id"`
	CustomData   map[string]any     `json:"custom_data,omitempty"`
	Capabilities []CapabilityAction `json:"capabilities"`
}

type CapabilityAction struct {
	Type  string                `json:"type"`
	State CapabilityActionState `json:"state"`
}

type CapabilityActionState struct {
	Instance string          `json:"instance"`
	Value    json.RawMessage `json:"value"`
	Relative *bool           `json:"relative,omitempty"`
}

type DevicesActionResponse struct {
	RequestID string               `json:"request_id"`
	Payload   DevicesActionResults `json:"payload"`
}

type DevicesActionResults struct {
	Devices []DeviceActionResult `json:"devices"`
}

type DeviceActionResult struct {
	ID           string                   `json:"id"`
	Capabilities []CapabilityActionResult `json:"capabilities,omitempty"`
	ActionResult *ActionResult            `json:"action_result,omitempty"`
}

type CapabilityActionResult struct {
	Type  string                      `json:"type"`
	State CapabilityActionResultState `json:"state"`
}

type CapabilityActionResultState struct {
	Instance     string       `json:"instance"`
	ActionResult ActionResult `json:"action_result"`
}

type ActionResult struct {
	Status       string `json:"status"`
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
}
