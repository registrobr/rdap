package protocol

// https://tools.ietf.org/html/rfc7483#section-6

type Error struct {
	RDAPConformance []string `json:"rdapConformance,omitempty"`
	Notices         []Notice `json:"notices,omitempty"`
	Lang            string   `json:"lang,omitempty"`
	ErrorCode       int
	Title           string
	Description     []string
}
