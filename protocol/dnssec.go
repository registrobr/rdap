package protocol

import "time"

type DS struct {
	KeyTag      int       `json:"keyTag"`
	Algorithm   int       `json:"algorithm"`
	Digest      string    `json:"digest"`
	DigestType  int       `json:"digestType"`
	Events      []Event   `json:"events,omitempty"`
	DSStatus    string    `json:"nicbr_status,omitempty"`
	LastCheckAt time.Time `json:"nicbr_last_check,omitempty"`
	LastOKAt    time.Time `json:"nicbr_last_ok,omitempty"`
}

type SecureDNS struct {
	ZoneSigned        bool `json:"zoneSigned"`
	DelegationsSigned bool `json:"delegationsSigned"`
	DSData            []DS `json:"dsData,omitempty"`
}
