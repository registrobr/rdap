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

func TestHandleHTTPStatusCode(t *testing.T) {
	tests := []struct {
		description string
		expectedErr error
		kind        kind
		err         protocol.Error
		header      map[string]string
	}{
		{
			description: "it should return a nil error",
			expectedErr: nil,
			kind:        dns,
			err: protocol.Error{
				ErrorCode: http.StatusOK,
			},
		},
		{
			description: "it should got a not found error",
			expectedErr: fmt.Errorf("%s not found.", dns),
			kind:        dns,
			err: protocol.Error{
				ErrorCode: http.StatusNotFound,
			},
		},
		{
			description: "it should got an unexpected response error",
			expectedErr: fmt.Errorf("unexpected response: %d %s",
				http.StatusForbidden, http.StatusText(http.StatusForbidden)),
			kind: dns,
			err: protocol.Error{
				ErrorCode: http.StatusForbidden,
			},
			header: map[string]string{"Content-Type": "application/text"},
		},
	}

	for i, test := range tests {
		t.Logf("Test case number %d", i)
		t.Logf("Test case description: %s", test.description)

		response := &http.Response{
			StatusCode: test.err.ErrorCode,
			Header:     http.Header{},
		}

		if len(test.header) > 0 {
			for k, v := range test.header {
				response.Header.Set(k, v)
			}
		}

		var c Client
		err := c.handleHTTPStatusCode(test.kind, response)
		if test.expectedErr == nil {
			if err == nil {
				// nothig to do
				continue
			}

			t.Fatalf("Expecting '%v', got '%s'",
				test.expectedErr,
				err.Error())
		}

		if err.Error() != test.expectedErr.Error() {
			t.Fatalf("Expecting '%s', got '%s'",
				test.expectedErr,
				err.Error())
		}
	}
}

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
			content, _ := ioutil.ReadAll(r.Body)
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
		status         int
		responseBody   string
		expectedObject interface{}
		expectedError  error
	}{
		{
			description:   "it should return an error due to an invalid uri",
			kind:          dns,
			identifier:    "example.br",
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from example.br:\n  parse %%gh&%%ij/domain/example.br: invalid URL escape \"%%gh\""),
		},
		{
			description:   "it should return an error due to invalid json in rdap response",
			kind:          dns,
			identifier:    "example.br",
			responseBody:  "invalid",
			expectedError: fmt.Errorf("error(s) fetching RDAP data from example.br:\n  invalid character 'i' looking for beginning of value"),
		},
		{
			description:    "it should return a valid domain object",
			kind:           dns,
			identifier:     "example.br",
			responseBody:   "{\"objectClassName\": \"domain\"}",
			expectedObject: map[string]interface{}{"objectClassName": "domain"},
		},
		{
			description:   "it should return an error due to non-ok http status code in response",
			kind:          dns,
			identifier:    "example.br",
			status:        http.StatusNotFound,
			responseBody:  "{}",
			expectedError: fmt.Errorf("error(s) fetching RDAP data from example.br:\n  unexpected response: 404 Not Found"),
		},
	}

	for i, test := range tests {
		var object interface{}

		c := NewClient(test.uris, nil)

		if len(test.responseBody) > 0 {
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if test.status > 0 {
						w.WriteHeader(test.status)
					}

					w.Write([]byte(test.responseBody))
				}),
			)

			c.uris = []string{ts.URL}
		}

		err := c.query(test.kind, test.identifier, &object)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedObject, object) {
				t.Fatalf("[%d] “%s”: expected “%v”, got “%v”", i, test.description, test.expectedObject, object)
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
			kind:        kind("ipnetwork"),
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			responseBody:   "{\"objectClassName\": \"ipv4\"}",
			expectedObject: &protocol.IPNetwork{ObjectClassName: "ipv4"},
		},
		{
			description: "it should return the right uris when matching an ipv6 network",
			kind:        kind("ipnetwork"),
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("2001:0200:1000::/48")
				return cidr
			}(),
			responseBody:   "{\"objectClassName\": \"ipv6\"}",
			expectedObject: &protocol.IPNetwork{ObjectClassName: "ipv6"},
		},
		{
			description:    "it should return the right uris when matching an entity",
			kind:           entity,
			identifier:     "example",
			responseBody:   "{\"objectClassName\": \"entity\"}",
			expectedObject: &protocol.Entity{ObjectClassName: "entity"},
		},
		{
			description:    "it should return the right uris when matching a ip",
			kind:           ip,
			identifier:     net.ParseIP("192.168.1.1"),
			responseBody:   "{\"objectClassName\": \"ip\"}",
			expectedObject: &protocol.IPNetwork{ObjectClassName: "ip"},
		},
		{
			description:   "it should return an error when matching a domain due to an invalid uri",
			kind:          dns,
			identifier:    "example.br",
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from example.br:\n  parse %%gh&%%ij/domain/example.br: invalid URL escape \"%%gh\""),
		},
		{
			description:   "it should return an error when matching an as number due to an invalid uri",
			kind:          asn,
			identifier:    uint64(1),
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from 1:\n  parse %%gh&%%ij/autnum/1: invalid URL escape \"%%gh\""),
		},
		{
			description: "it should return an error when matching an ip network due to an invalid uri",
			kind:        kind("ipnetwork"),
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from 192.168.0.0/24:\n  parse %%gh&%%ij/ip/192.168.0.0/24: invalid URL escape \"%%gh\""),
		},
		{
			description:   "it should return an error when matching an ip due to an invalid uri",
			kind:          ip,
			identifier:    net.ParseIP("192.168.1.1"),
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from 192.168.1.1:\n  parse %%gh&%%ij/ip/192.168.1.1: invalid URL escape \"%%gh\""),
		},
		{
			description:   "it should return an error when matching an ip due to an invalid uri",
			kind:          entity,
			identifier:    "example",
			uris:          []string{"%gh&%ij"},
			expectedError: fmt.Errorf("error(s) fetching RDAP data from example:\n  parse %%gh&%%ij/entity/example: invalid URL escape \"%%gh\""),
		},
	}

	for i, test := range tests {
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
		case kind("ipnetwork"):
			object, err = c.IPNetwork(test.identifier.(*net.IPNet))
		case ip:
			object, err = c.IP(test.identifier.(net.IP))
		case entity:
			object, err = c.Entity(test.identifier.(string))
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedObject, object) {
				t.Fatalf("[%d] “%s”: expected “%v”, got “%v”", i, test.description, test.expectedObject, object)
			}
		}
	}
}
