package cobalt

type TunnelResponse struct {
	Status   string `json:"status"`
	Url      string `json:"url"`
	FileName string `json:"filename"`
}

type PickerObject struct {
	Type  string `json:"type"`
	Url   string `json:"url"`
	Thumb string `json:"thumb"`
}

type PickerResponse struct {
	Status        string         `json:"status"`
	Audio         string         `json:"audio"`
	AudioFilename string         `json:"audioFilename"`
	Picker        []PickerObject `json:"picker"`
}

type ErrorContext struct {
	Service string `json:"service"`
	Limit   int    `json:"limit"`
}

type ErrorObject struct {
	Code    int          `json:"code"`
	Context ErrorContext `json:"context"`
}

type ErrorResponse struct {
	Status string      `json:"status"`
	Error  ErrorObject `json:"error"`
}
