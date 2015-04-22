package protocol

import (
	"encoding/json"
	"sort"
	"strings"
	"time"
)

type ServiceRegistry struct {
	Version     string       `json:"version"`
	Publication time.Time    `json:"publication"`
	Description string       `json:"description,omitempty"`
	Services    ServicesList `json:"services"`
}

type ServicesList []Service

type Service [2]Values

type Values []string

func (s Service) Entries() []string {
	return s[0]
}

func (s Service) URIs() []string {
	return s[1]
}

func (s *Service) UnmarshalJSON(b []byte) error {
	sv := [2]Values(Service{})

	if err := json.Unmarshal(b, &sv); err != nil {
		return err
	}

	sort.Sort(sv[1])
	*s = sv

	return nil
}

func (v Values) Len() int {
	return len(v)
}

func (v Values) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v Values) Less(i, j int) bool {
	return strings.Split(v[i], ":")[0] == "https"
}
