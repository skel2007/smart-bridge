package devices

type Device struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Online   bool   `json:"online"`
}
