package bootstrap

import (
	"strings"
	"time"
)

// ServiceRegistry reflects the structure of a RDAP Bootstrap Service
// Registry.
//
// See http://tools.ietf.org/html/rfc7484#section-10.2
type serviceRegistry struct {
	Version     string    `json:"version"`
	Publication time.Time `json:"publication"`
	Description string    `json:"description,omitempty"`
	Services    []service `json:"services"`
}

// service is an array composed by two items. The first one is a list of
// entries and the second one is a list of URIs.
type service [2][]string

// entries is a helper that returns the list of entries of a service
func (s service) entries() []string {
	return s[0]
}

// uris is a helper that returns the list of URIs of a service
func (s service) uris() []string {
	return s[1]
}

type prioritizeHTTPS []string

func (v prioritizeHTTPS) Len() int {
	return len(v)
}

func (v prioritizeHTTPS) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v prioritizeHTTPS) Less(i, j int) bool {
	return strings.Split(v[i], ":")[0] == "https"
}
