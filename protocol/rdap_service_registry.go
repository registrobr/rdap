package protocol

import "time"

type RDAPServiceRegistry struct {
	Version     string       `json:"version"`
	Publication time.Time    `json:"publication"`
	Description string       `json:"description,omitempty"`
	Services    ServicesList `json:"services"`
}

type ServicesList []Service

type Service [2][]string
