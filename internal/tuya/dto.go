package tuya

import (
	"encoding/json"
)

type tuyaTokenResult struct {
	AccessToken string `json:"access_token"`
}

type tuyaDevice struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomName string `json:"customName"`
	Category   string `json:"category"`
	IsOnline   bool   `json:"isOnline"`
}

type tuyaDeviceSpecifications struct {
	Category  string             `json:"category"`
	Functions []tuyaFunctionSpec `json:"functions"`
	Status    []tuyaStatusSpec   `json:"status"`
}

type tuyaFunctionSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type tuyaStatusSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type tuyaDeviceStatus struct {
	Code  string          `json:"code"`
	Value json.RawMessage `json:"value"`
}

type tuyaCommand struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}
