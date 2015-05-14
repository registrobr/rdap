package client

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/registrobr/rdap-client/protocol"
)

func TestFetch(t *testing.T) {
	tests := []struct {
		description   string
		uri           string
		expectedBody  string
		expectedError error
	}{
		{
			description:   "it should return an error due to an invalid URI",
			uri:           "%gh&%ij",
			expectedError: fmt.Errorf("parse %%gh&%%ij: invalid URL escape \"%%gh\""),
		},
	}

	for _, test := range tests {
		c := NewClient(nil, nil)
		body := ""
		r, err := c.fetch(test.uri)

		if err == nil {
			content, _ := ioutil.ReadAll(r)
			body = string(content)
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error “%s”, got “%s”", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedBody, body) {
				t.Fatalf("“%s”: expected “%v”, got “%v”", test.description, test.expectedBody, body)
			}
		}
	}
}

func TestQuery(t *testing.T) {
	tests := []struct {
		description    string
		kind           kind
		identifier     interface{}
		uris           []string
		responseBody   string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "it should return an error due to an invalid uri",
			kind:          dns,
			identifier:    "example.br",
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("no data available for example.br"),
		},
		{
			description:   "it should return an error due to invalid json in rdap response",
			kind:          dns,
			identifier:    "example.br",
			responseBody:  "invalid",
			expectedError: fmt.Errorf("no data available for example.br"),
		},
		{
			description:    "it should return a valid domain object",
			kind:           dns,
			identifier:     "example.br",
			responseBody:   "{\"objectClassName\": \"domain\"}",
			expectedObject: map[string]interface{}{"objectClassName": "domain"},
		},
	}

	for _, test := range tests {
		var object interface{}

		c := NewClient(test.uris, nil)

		if len(test.responseBody) > 0 {
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(test.responseBody))
				}),
			)

			c.uris = []string{ts.URL}
		}

		err := c.query(test.kind, test.identifier, &object)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error “%s”, got “%s”", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedObject, object) {
				t.Fatalf("“%s”: expected “%v”, got “%v”", test.description, test.expectedObject, object)
			}
		}
	}
}

func TestQueriers(t *testing.T) {
	tests := []struct {
		description    string
		kind           kind
		identifier     interface{}
		uris           []string
		responseBody   string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:    "it should return the right object when matching a domain",
			kind:           dns,
			identifier:     "example.br",
			responseBody:   "{\"objectClassName\": \"domain\"}",
			expectedObject: &protocol.DomainResponse{ObjectClassName: "domain"},
		},
		{
			description:    "it should return the right uris when matching a domain",
			kind:           asn,
			identifier:     uint64(1),
			responseBody:   "{\"objectClassName\": \"as\"}",
			expectedObject: &protocol.ASResponse{ObjectClassName: "as"},
		},
		{
			description: "it should return the right uris when matching an ipv4 network",
			kind:        ip,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			responseBody:   "{\"objectClassName\": \"ipv4\"}",
			expectedObject: &protocol.IPNetwork{ObjectClassName: "ipv4"},
		},
		{
			description: "it should return the right uris when matching an ipv4 network",
			kind:        ip,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("2001:0200:1000::/48")
				return cidr
			}(),
			responseBody:   "{\"objectClassName\": \"ipv6\"}",
			expectedObject: &protocol.IPNetwork{ObjectClassName: "ipv6"},
		},
		{
			description:   "it should return an error when matching a domain due to an invalid uri",
			kind:          dns,
			identifier:    "example.br",
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("no data available for example.br"),
		},
		{
			description:   "it should return an error when matching an as number due to an invalid uri",
			kind:          asn,
			identifier:    uint64(1),
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("no data available for 1"),
		},
		{
			description: "it should return an error when matching an ip network due to an invalid uri",
			kind:        ip,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("no data available for 192.168.0.0/24"),
		},
	}

	for _, test := range tests {
		c := NewClient(nil, nil)

		if len(test.uris) == 0 {
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(test.responseBody))
				}),
			)
			c.uris = []string{ts.URL}
		} else {
			c.uris = test.uris
		}

		var (
			object interface{}
			err    error
		)

		switch test.kind {
		case dns:
			object, err = c.Domain(test.identifier.(string))
		case asn:
			object, err = c.ASN(test.identifier.(uint64))
		case ip:
			object, err = c.IPNetwork(test.identifier.(*net.IPNet))
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error “%s”, got “%s”", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedObject, object) {
				t.Fatalf("“%s”: expected “%v”, got “%v”", test.description, test.expectedObject, object)
			}
		}
	}
}
