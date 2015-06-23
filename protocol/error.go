package protocol

type Error struct {
	Notices     []Notice `json:"notices,omitempty"`
	Lang        string   `json:"lang,omitempty"`
	ErrorCode   int
	Title       string
	Description []string
	Conformance
}
