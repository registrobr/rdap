package protocol

type Remark struct {
	Type        string   `json:"type,omitempty"`
	Description []string `json:"description"`
}
