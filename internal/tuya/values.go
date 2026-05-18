package tuya

import "encoding/json"

func decodeTuyaValues(raw json.RawMessage, out any) bool {
	if decodeRawJSON(raw, out) {
		return true
	}

	// Tuya returns specification "values" both as JSON objects and as JSON-encoded strings.
	var text string
	if !decodeRawJSON(raw, &text) || text == "" {
		return false
	}

	return json.Unmarshal([]byte(text), out) == nil
}

func decodeRawJSON(raw json.RawMessage, out any) bool {
	if len(raw) == 0 {
		return false
	}

	return json.Unmarshal(raw, out) == nil
}

type tuyaIntegerValues struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Scale float64 `json:"scale"`
	Step  float64 `json:"step"`
}

func (values tuyaIntegerValues) validRange() bool {
	return values.Max > values.Min
}

type tuyaEnumValues struct {
	Range []string `json:"range"`
}

type tuyaHSVValue struct {
	Hue        float64 `json:"h"`
	Saturation float64 `json:"s"`
	Value      float64 `json:"v"`
}
