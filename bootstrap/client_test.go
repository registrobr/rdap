package bootstrap

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
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
		c := NewClient(nil)
		body := ""
		r, _, err := c.fetch(test.uri)

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
		description   string
		bootstrap     string
		kind          kind
		identifier    interface{}
		responseBody  string
		expectedURIs  []string
		expectedError error
	}{
		{
			description:   "it should return an error due to an invalid bootstrap URI",
			kind:          dns,
			identifier:    "teste",
			bootstrap:     "%%gh&%%ij/%v",
			expectedError: fmt.Errorf("parse %%gh&%%ij/dns: invalid URL escape \"%%gh\""),
		},
		{
			description:   "it should return an error due to invalid JSON in bootstrap response",
			kind:          dns,
			identifier:    "test",
			responseBody:  "invalid",
			expectedError: fmt.Errorf("invalid character 'i' looking for beginning of value"),
		},
		{
			description:  "it should return the right uris when matching a domain",
			kind:         dns,
			identifier:   "example.br",
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"br\"], [\"rdap-domain.example.br\"]]]}",
			expectedURIs: []string{"rdap-domain.example.br"},
		},
		{
			description:  "it should return the right uris when matching an as number",
			kind:         asn,
			identifier:   uint64(5),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"1-10\"], [\"rdap-as.example.br\"]]]}",
			expectedURIs: []string{"rdap-as.example.br"},
		},
		{
			description: "it should return the right uris when matching an ip network",
			kind:        ipv4,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"192.168.0.0/16\"], [\"rdap-ip.example.br\"]]]}",
			expectedURIs: []string{"rdap-ip.example.br"},
		},
		{
			description:  "it should return no uris when matching a domain",
			kind:         dns,
			identifier:   "example.br",
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"net\"], [\"rdap-domain.example.net\"]]]}",
			expectedURIs: nil,
		},
		{
			description:  "it should return an error due to an invalid as range",
			kind:         asn,
			identifier:   uint64(1),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"1-invalid\"], [\"rdap-as.example.net\"]]]}",
			expectedURIs: nil,
		},
		{
			description:   "it should return an error due to incompatible bootstrap spec version",
			kind:          asn,
			identifier:    uint64(1),
			responseBody:  "{\"version\":\"2.0\",\"services\":[[[\"1-invalid\"], [\"rdap-as.example.net\"]]]}",
			expectedError: fmt.Errorf("incompatible bootstrap specification version: 2.0 (expecting 1.0)"),
		},
	}

	for _, test := range tests {
		c := NewClient(nil)

		if test.bootstrap != "" {
			c.Bootstrap = test.bootstrap
		} else {
			ts := httptest.NewServer(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Write([]byte(test.responseBody))
				}),
			)

			c.Bootstrap = ts.URL + "/%v"
		}

		var (
			uris []string
			err  error
		)

		switch test.kind {
		case dns:
			uris, err = c.query(test.kind, test.identifier.(string))
		case asn:
			uris, err = c.query(test.kind, test.identifier.(uint64))
		case ipv4, ipv6:
			uris, err = c.query(test.kind, test.identifier.(*net.IPNet))
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error “%s”, got “%s”", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedURIs, uris) {
				t.Fatalf("“%s”: expected “%v”, got “%v”", test.description, test.expectedURIs, uris)
			}
		}
	}
}

func TestQueriers(t *testing.T) {
	tests := []struct {
		description   string
		bootstrap     string
		kind          kind
		identifier    interface{}
		responseBody  string
		expectedURIs  []string
		expectedError error
	}{
		{
			description:  "it should return the right uris when matching a domain",
			kind:         dns,
			identifier:   "example.br",
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"br\"], [\"rdap-domain.example.br\"]]]}",
			expectedURIs: []string{"rdap-domain.example.br"},
		},
		{
			description:  "it should return the right uris when matching a domain",
			kind:         asn,
			identifier:   uint64(1),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"1-10\"], [\"rdap-as.example.br\"]]]}",
			expectedURIs: []string{"rdap-as.example.br"},
		},
		{
			description: "it should return the right uris when matching an ipv4 network",
			kind:        ipv4,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
				return cidr
			}(),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"192.168.0.0/16\"], [\"rdap-ip.example.br\"]]]}",
			expectedURIs: []string{"rdap-ip.example.br"},
		},
		{
			description: "it should return the right uris when matching an ipv4 network",
			kind:        ipv6,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("2001:0200:1000::/48")
				return cidr
			}(),
			responseBody: "{\"version\":\"1.0\",\"services\":[[[\"2001:0200:1000::/36\"], [\"rdap-ip.example.br\"]]]}",
			expectedURIs: []string{"rdap-ip.example.br"},
		},
	}

	for _, test := range tests {
		c := NewClient(nil)
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(test.responseBody))
			}),
		)

		c.Bootstrap = ts.URL + "/%v"

		var (
			uris []string
			err  error
		)

		switch test.kind {
		case dns:
			uris, err = c.Domain(test.identifier.(string))
		case asn:
			uris, err = c.ASN(test.identifier.(uint64))
		case ipv4, ipv6:
			uris, err = c.IPNetwork(test.identifier.(*net.IPNet))
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error “%s”, got “%s”", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expectedURIs, uris) {
				t.Fatalf("“%s”: expected “%v”, got “%v”", test.description, test.expectedURIs, uris)
			}
		}
	}
}
