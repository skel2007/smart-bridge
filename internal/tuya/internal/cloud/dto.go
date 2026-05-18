package cloud

import (
	"encoding/json"
)

type tokenResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireTime   int64  `json:"expire_time"`
}

type Device struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	CustomName string `json:"customName"`
	Category   string `json:"category"`
	IsOnline   bool   `json:"isOnline"`
}

type DeviceSpecifications struct {
	Category  string         `json:"category"`
	Functions []FunctionSpec `json:"functions"`
	Status    []StatusSpec   `json:"status"`
}

type FunctionSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type StatusSpec struct {
	Code   string          `json:"code"`
	Type   string          `json:"type"`
	Values json.RawMessage `json:"values"`
}

type DeviceStatus struct {
	Code  string          `json:"code"`
	Value json.RawMessage `json:"value"`
}

type Command struct {
	Code  string `json:"code"`
	Value any    `json:"value"`
}

type commandsRequest struct {
	Commands []Command `json:"commands"`
}
