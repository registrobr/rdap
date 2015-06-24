package protocol

type PublicID struct {
	Type       string `json:"type"`
	Identifier string `json:"identifier"`
}

type CustomerSupportService struct {
	Email   string `json:"email,omitempty"`
	Website string `json:"website,omitempty"`
	Phone   string `json:"phone,omitempty"`
}

type Entity struct {
	ObjectClassName        string                 `json:"objectClassName"`
	Handle                 string                 `json:"handle"`
	VCardArray             []interface{}          `json:"vcardArray,omitempty"`
	Roles                  []string               `json:"roles,omitempty"`
	PublicIds              []PublicID             `json:"publicIds,omitempty"`
	Responsible            string                 `json:"nicbr_responsible,omitempty"`
	CustomerSupportService CustomerSupportService `json:"nicbr_customer_support_service,omitempty"`
	Entities               []Entity               `json:"entities,omitempty"`
	Events                 []Event                `json:"events,omitempty"`
	Links                  []Link                 `json:"links,omitempty"`
	Remarks                []Remark               `json:"remarks,omitempty"`
	Conformance
}
