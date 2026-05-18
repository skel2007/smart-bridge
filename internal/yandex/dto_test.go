package yandex

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequestIDResponseJSON(t *testing.T) {
	response := RequestIDResponse{RequestID: "request-1"}

	data, err := json.Marshal(response)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"request_id": "request-1"
	}`, string(data))
}

func TestDevicesResponseJSON(t *testing.T) {
	randomAccess := true
	response := DevicesResponse{
		RequestID: "request-1",
		Payload: DevicesPayload{
			UserID: "bridge-user",
			Devices: []DeviceDescription{
				{
					ID:          "light-1",
					Name:        "Desk light",
					StatusInfo:  StatusInfo{Reportable: false},
					Description: "Main desk light",
					Room:        "Office",
					Type:        "devices.types.light",
					CustomData: map[string]any{
						"upstream_id": "tuya-1",
					},
					Capabilities: []CapabilityDescription{
						{
							Type:        "devices.capabilities.on_off",
							Retrievable: true,
							Reportable:  false,
						},
						{
							Type:        "devices.capabilities.range",
							Retrievable: true,
							Reportable:  false,
							Parameters: RangeParameters{
								Instance:     "brightness",
								Unit:         "unit.percent",
								RandomAccess: &randomAccess,
								Range:        &ValueRange{Min: 0, Max: 100, Precision: 1},
							},
						},
						{
							Type:        "devices.capabilities.color_setting",
							Retrievable: true,
							Reportable:  false,
							Parameters: ColorSettingParameters{
								ColorModel:  "hsv",
								Temperature: &TemperatureKRange{Min: 2700, Max: 6500},
							},
						},
					},
					DeviceInfo: &DeviceInfo{
						Manufacturer: "Tuya",
						Model:        "unknown",
					},
				},
			},
		},
	}

	data, err := json.Marshal(response)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"request_id": "request-1",
		"payload": {
			"user_id": "bridge-user",
			"devices": [
				{
					"id": "light-1",
					"name": "Desk light",
					"status_info": {
						"reportable": false
					},
					"description": "Main desk light",
					"room": "Office",
					"type": "devices.types.light",
					"custom_data": {
						"upstream_id": "tuya-1"
					},
					"capabilities": [
						{
							"type": "devices.capabilities.on_off",
							"retrievable": true,
							"reportable": false
						},
						{
							"type": "devices.capabilities.range",
							"retrievable": true,
							"reportable": false,
							"parameters": {
								"instance": "brightness",
								"unit": "unit.percent",
								"random_access": true,
								"range": {
									"min": 0,
									"max": 100,
									"precision": 1
								}
							}
						},
						{
							"type": "devices.capabilities.color_setting",
							"retrievable": true,
							"reportable": false,
							"parameters": {
								"color_model": "hsv",
								"temperature_k": {
									"min": 2700,
									"max": 6500
								}
							}
						}
					],
					"device_info": {
						"manufacturer": "Tuya",
						"model": "unknown"
					}
				}
			]
		}
	}`, string(data))
}

func TestDevicesQueryRequestJSON(t *testing.T) {
	data := []byte(`{
		"devices": [
			{
				"id": "light-1",
				"custom_data": {
					"upstream_id": "tuya-1"
				}
			}
		]
	}`)

	var request DevicesQueryRequest
	err := json.Unmarshal(data, &request)

	require.NoError(t, err)
	require.Len(t, request.Devices, 1)
	require.Equal(t, "light-1", request.Devices[0].ID)
	require.Equal(t, "tuya-1", request.Devices[0].CustomData["upstream_id"])
}

func TestDevicesQueryResponseJSON(t *testing.T) {
	response := DevicesQueryResponse{
		RequestID: "request-1",
		Payload: DevicesQueryPayload{
			Devices: []DeviceState{
				{
					ID: "light-1",
					Capabilities: []CapabilityState{
						{
							Type: "devices.capabilities.on_off",
							State: CapabilityStateValue{
								Instance: "on",
								Value:    true,
							},
						},
						{
							Type: "devices.capabilities.color_setting",
							State: CapabilityStateValue{
								Instance: "hsv",
								Value:    HSVValue{H: 255, S: 100, V: 50},
							},
						},
					},
				},
				{
					ID:           "missing-device",
					ErrorCode:    "DEVICE_NOT_FOUND",
					ErrorMessage: "device not found",
				},
			},
		},
	}

	data, err := json.Marshal(response)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"request_id": "request-1",
		"payload": {
			"devices": [
				{
					"id": "light-1",
					"capabilities": [
						{
							"type": "devices.capabilities.on_off",
							"state": {
								"instance": "on",
								"value": true
							}
						},
						{
							"type": "devices.capabilities.color_setting",
							"state": {
								"instance": "hsv",
								"value": {
									"h": 255,
									"s": 100,
									"v": 50
								}
							}
						}
					]
				},
				{
					"id": "missing-device",
					"error_code": "DEVICE_NOT_FOUND",
					"error_message": "device not found"
				}
			]
		}
	}`, string(data))
}

func TestDevicesActionRequestKeepsCapabilityValuesRaw(t *testing.T) {
	data := []byte(`{
		"payload": {
			"devices": [
				{
					"id": "light-1",
					"custom_data": {
						"upstream_id": "tuya-1"
					},
					"capabilities": [
						{
							"type": "devices.capabilities.range",
							"state": {
								"instance": "brightness",
								"value": 42,
								"relative": false
							}
						},
						{
							"type": "devices.capabilities.color_setting",
							"state": {
								"instance": "hsv",
								"value": {
									"h": 120,
									"s": 80,
									"v": 90
								}
							}
						}
					]
				}
			]
		}
	}`)

	var request DevicesActionRequest
	err := json.Unmarshal(data, &request)

	require.NoError(t, err)
	require.Len(t, request.Payload.Devices, 1)
	device := request.Payload.Devices[0]
	require.Equal(t, "light-1", device.ID)
	require.Equal(t, "tuya-1", device.CustomData["upstream_id"])
	require.Len(t, device.Capabilities, 2)
	require.JSONEq(t, `42`, string(device.Capabilities[0].State.Value))
	require.NotNil(t, device.Capabilities[0].State.Relative)
	require.False(t, *device.Capabilities[0].State.Relative)
	require.JSONEq(t, `{"h":120,"s":80,"v":90}`, string(device.Capabilities[1].State.Value))
}

func TestDevicesActionResponseJSON(t *testing.T) {
	response := DevicesActionResponse{
		RequestID: "request-1",
		Payload: DevicesActionResults{
			Devices: []DeviceActionResult{
				{
					ID: "light-1",
					Capabilities: []CapabilityActionResult{
						{
							Type: "devices.capabilities.on_off",
							State: CapabilityActionResultState{
								Instance: "on",
								ActionResult: ActionResult{
									Status: "DONE",
								},
							},
						},
					},
				},
				{
					ID: "light-2",
					ActionResult: &ActionResult{
						Status:       "ERROR",
						ErrorCode:    "DEVICE_UNREACHABLE",
						ErrorMessage: "upstream is unavailable",
					},
				},
			},
		},
	}

	data, err := json.Marshal(response)

	require.NoError(t, err)
	require.JSONEq(t, `{
		"request_id": "request-1",
		"payload": {
			"devices": [
				{
					"id": "light-1",
					"capabilities": [
						{
							"type": "devices.capabilities.on_off",
							"state": {
								"instance": "on",
								"action_result": {
									"status": "DONE"
								}
							}
						}
					]
				},
				{
					"id": "light-2",
					"action_result": {
						"status": "ERROR",
						"error_code": "DEVICE_UNREACHABLE",
						"error_message": "upstream is unavailable"
					}
				}
			]
		}
	}`, string(data))
}
