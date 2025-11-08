package handshake

type RequestPayload struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Ticket  string `json:"ticket,omitempty"`
}

type Response struct {
	Status string `json:"status"`
}
