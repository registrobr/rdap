package protocol

import "fmt"

type Error struct {
	Notices     []Notice `json:"notices,omitempty"`
	Lang        string   `json:"lang,omitempty"`
	ErrorCode   int      `json:"errorCode,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description []string `json:"description,omitempty"`
	Conformance
}

func (e Error) Error() string {
	return fmt.Sprintf("%d", e.ErrorCode)
}
