package protocol

import "br/core/protocol"

type DS struct {
	KeyTag     int                   `json:"keyTag"`
	Algorithm  protocol.DSAlgorithm  `json:"algorithm"`
	Digest     string                `json:"digest"`
	DigestType protocol.DSDigestType `json:"digestType"`
	Events     []Event               `json:"events,omitempty"`
}

type SecureDNS struct {
	ZoneSigned        bool `json:"zoneSigned"`
	DelegationsSigned bool `json:"delegationsSigned"`
	DSData            []DS `json:"dsData"`
}
