package bootstrap

import (
	"strings"
	"time"
)

// ServiceRegistry reflects the structure of a RDAP Bootstrap Service
// Registry.
//
// See http://tools.ietf.org/html/rfc7484#section-10.2
type ServiceRegistry struct {
	Version     string    `json:"version"`
	Publication time.Time `json:"publication"`
	Description string    `json:"description,omitempty"`
	Services    []Service `json:"services"`
}

// Service is an array composed by two items. The first one is a list of
// entries and the second one is a list of URIs.
type Service [2][]string

// Entries is a helper that returns the list of entries of a service
func (s Service) Entries() []string {
	return s[0]
}

// URIs is a helper that returns the list of URIs of a service
func (s Service) URIs() []string {
	return s[1]
}

type PrioritizeHTTPS []string

func (v PrioritizeHTTPS) Len() int {
	return len(v)
}

func (v PrioritizeHTTPS) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v PrioritizeHTTPS) Less(i, j int) bool {
	return strings.Split(v[i], ":")[0] == "https"
}
