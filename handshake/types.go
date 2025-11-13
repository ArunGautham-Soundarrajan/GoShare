package handshake

type RequestPayload struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Ticket  string `json:"ticket,omitempty"`
}

type Response struct {
	Status string `json:"status"`
}

type FileData struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

type FileInfoPayload struct {
	Type     string `json:"type"`
	FileName string `json:"filename"`
	Size     int64  `json:"size"`
}
