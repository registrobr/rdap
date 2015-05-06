package protocol

type IPAddresses struct {
	V4 []string `json:"v4,omitempty"`
	V6 []string `json:"v6,omitempty"`
}

type Nameserver struct {
	ObjectClassName string       `json:"objectClassName"`
	LDHName         string       `json:"ldhName,omitempty"`
	UnicodeName     string       `json:"unicodeName,omitempty"`
	IPAddresses     *IPAddresses `json:"ipAddresses"`
	Status          []Status     `json:"status,omitempty"`
	Remarks         []Remark     `json:"remarks,omitempty"`
}
