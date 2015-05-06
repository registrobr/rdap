package protocol

type DomainResponse struct {
	ObjectClassName string       `json:"objectClassName"`
	LDHName         string       `json:"ldhName,omitempty"`
	UnicodeName     string       `json:"unicodeName,omitempty"`
	Nameservers     []Nameserver `json:"nameservers"`
	SecureDNS       SecureDNS    `json:"secureDNS,omitempty"`
	Links           []Link       `json:"links,omitempty"`
	Entities        []Entity     `json:"entities,omitempty"`
	Events          []Event      `json:"events,omitempty"`
	Status          []Status     `json:"status,omitempty"`
	PublicIDs       []PublicID   `json:"publicIds,omitempty"`
	Remarks         []Remark     `json:"remarks,omitempty"`
}
