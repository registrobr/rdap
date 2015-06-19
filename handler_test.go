package rdap

import (
	"fmt"

	"reflect"
	"testing"

	"github.com/registrobr/rdap/protocol"
)

func TestHandlerDomain(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		bootstrapEntry string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "Domain handler should not be executed due to an invalid identifier in input",
			identifier:    "invalid&invalid",
			expectedError: ErrInvalidQuery,
		},
		{
			description: "Domain handler should return a valid RDAP response",
			identifier:  "example.br",
			expectedObject: protocol.DomainResponse{
				ObjectClassName: "domain",
				LDHName:         "example.br",
			},
		},
		{
			description:    "Domain handler should return a valid RDAP response (bootstrapped)",
			identifier:     "example.br",
			bootstrapEntry: "br",
			expectedObject: protocol.DomainResponse{
				ObjectClassName: "domain",
				LDHName:         "example.br",
			},
		},
	}

	for i, test := range tests {
		ts, bs := createTestServers(test.expectedObject, test.bootstrapEntry)

		h := Handler{
			URIs: []string{ts.URL},
		}

		if len(test.bootstrapEntry) > 0 {
			h.Bootstrap = &Bootstrap{Bootstrap: bs.URL + "/%v"}
		}

		object, err := h.Domain(test.identifier)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if object != nil && !reflect.DeepEqual(test.expectedObject, *object) {
				for _, l := range diff(test.expectedObject, *object) {
					t.Log(l)
				}

				t.Fatalf("[%d] “%s”", i, test.description)
			}
		}
	}
}

func TestHandlerASN(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		bootstrapEntry string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "ASN handler should not be executed due to an invalid identifier in input",
			identifier:    "invalid&invalid",
			expectedError: ErrInvalidQuery,
		},
		{
			description: "ASN handler should return a valid RDAP response",
			identifier:  "1",
			expectedObject: protocol.ASResponse{
				ObjectClassName: "as",
				StartAutnum:     1,
				EndAutnum:       16,
			},
		},
		{
			description:    "ASN handler should return a valid RDAP response (bootstrapped)",
			identifier:     "1",
			bootstrapEntry: "1-16",
			expectedObject: protocol.ASResponse{
				ObjectClassName: "as",
				StartAutnum:     1,
				EndAutnum:       16,
			},
		},
	}

	for i, test := range tests {
		ts, bs := createTestServers(test.expectedObject, test.bootstrapEntry)

		h := Handler{
			URIs: []string{ts.URL},
		}

		if len(test.bootstrapEntry) > 0 {
			h.Bootstrap = &Bootstrap{Bootstrap: bs.URL + "/%v"}
		}

		object, err := h.ASN(test.identifier)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if object != nil && !reflect.DeepEqual(test.expectedObject, *object) {
				for _, l := range diff(test.expectedObject, *object) {
					t.Log(l)
				}

				t.Fatalf("[%d] “%s”", i, test.description)
			}
		}
	}
}

func TestHandlerIP(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		bootstrapEntry string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "IP handler should not be executed due to an invalid identifier in input",
			identifier:    "invalid",
			expectedError: ErrInvalidQuery,
		},
		{
			description: "IP handler should return a valid RDAP response",
			identifier:  "192.168.0.1",
			expectedObject: protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				StartAddress:    "192.168.0.0",
				EndAddress:      "192.168.0.255",
			},
		},
		{
			description:    "IP handler should return a valid RDAP response (bootstrapped)",
			identifier:     "192.168.0.1",
			bootstrapEntry: "192.168.0.0/24",
			expectedObject: protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				StartAddress:    "192.168.0.0",
				EndAddress:      "192.168.0.255",
			},
		},
	}

	for i, test := range tests {
		ts, bs := createTestServers(test.expectedObject, test.bootstrapEntry)

		h := Handler{
			URIs: []string{ts.URL},
		}

		if len(test.bootstrapEntry) > 0 {
			h.Bootstrap = &Bootstrap{Bootstrap: bs.URL + "/%v"}
		}

		object, err := h.IP(test.identifier)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if object != nil && !reflect.DeepEqual(test.expectedObject, *object) {
				for _, l := range diff(test.expectedObject, *object) {
					t.Log(l)
				}

				t.Fatalf("[%d] “%s”", i, test.description)
			}
		}
	}
}

func TestHandlerIPNetwork(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		bootstrapEntry string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "IP handler should not be executed due to an invalid identifier in input",
			identifier:    "invalid",
			expectedError: ErrInvalidQuery,
		},
		{
			description: "IP handler should return a valid RDAP response",
			identifier:  "192.168.0.0/24",
			expectedObject: protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				StartAddress:    "192.168.0.0",
				EndAddress:      "192.168.0.255",
			},
		},
		{
			description:    "IP handler should return a valid RDAP response (bootstrapped)",
			identifier:     "192.168.0.0/24",
			bootstrapEntry: "192.168.0.0/16",
			expectedObject: protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				StartAddress:    "192.168.0.0",
				EndAddress:      "192.168.0.255",
			},
		},
	}

	for i, test := range tests {
		ts, bs := createTestServers(test.expectedObject, test.bootstrapEntry)

		h := Handler{
			URIs: []string{ts.URL},
		}

		if len(test.bootstrapEntry) > 0 {
			h.Bootstrap = &Bootstrap{Bootstrap: bs.URL + "/%v"}
		}

		object, err := h.IPNetwork(test.identifier)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if object != nil && !reflect.DeepEqual(test.expectedObject, *object) {
				for _, l := range diff(test.expectedObject, *object) {
					t.Log(l)
				}

				t.Fatalf("[%d] “%s”", i, test.description)
			}
		}
	}
}

func TestHandlerEntity(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		bootstrapEntry string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description: "Entity handler should return a valid RDAP response",
			identifier:  "someone",
			expectedObject: protocol.Entity{
				ObjectClassName: "entity",
				Handle:          "someone",
			},
		},
	}

	for i, test := range tests {
		ts, bs := createTestServers(test.expectedObject, test.bootstrapEntry)

		h := Handler{
			URIs: []string{ts.URL},
		}

		if len(test.bootstrapEntry) > 0 {
			h.Bootstrap = &Bootstrap{Bootstrap: bs.URL + "/%v"}
		}

		object, err := h.Entity(test.identifier)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if object != nil && !reflect.DeepEqual(test.expectedObject, *object) {
				for _, l := range diff(test.expectedObject, *object) {
					t.Log(l)
				}

				t.Fatalf("[%d] “%s”", i, test.description)
			}
		}
	}
}

func TestHandlerQuery(t *testing.T) {
	tests := []struct {
		description    string
		identifier     string
		expectedObject interface{}
	}{
		{
			description:    "Generic handler should return an object of type protocol.ASResponse",
			identifier:     "1",
			expectedObject: protocol.ASResponse{},
		},
		{
			description:    "Generic handler should return an object of type protocol.IPNetwork",
			identifier:     "192.168.0.1",
			expectedObject: protocol.IPNetwork{},
		},
		{
			description:    "Generic handler should return an object of type protocol.IPNetwork",
			identifier:     "192.168.0.0/24",
			expectedObject: protocol.IPNetwork{},
		},
		{
			description:    "Generic handler should return an object of type protocol.DomainResponse",
			identifier:     "example.br",
			expectedObject: protocol.DomainResponse{},
		},
		{
			description:    "Generic handler should return an object of type protocol.Entity",
			identifier:     "someone",
			expectedObject: protocol.Entity{},
		},
	}

	for i, test := range tests {
		ts, _ := createTestServers(test.expectedObject, "")

		h := Handler{
			URIs: []string{ts.URL},
		}

		object, err := h.Query(test.identifier)

		if err != nil && err != ErrInvalidQuery {
			t.Fatal("unexpected error:", err)
		}

		expectedObjType := objType(test.expectedObject)
		objType := objType(object)

		if expectedObjType != objType {
			t.Fatalf("[%d] “%s” expected type “%s”, got “%s” ", i, test.description, expectedObjType, objType)
		}
	}
}
