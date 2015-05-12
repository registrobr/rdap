package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/registrobr/rdap-client/bootstrap"
	"github.com/registrobr/rdap-client/protocol"
)

type scenario struct {
	description   string
	endpointURI   string
	kind          kind
	identifier    interface{}
	object        interface{}
	registryBody  string
	registry      *bootstrap.ServiceRegistry
	rdapObject    interface{}
	keepURIs      bool
	expected      interface{}
	expectedError error
}

func TestFetchAndUnmarshal(t *testing.T) {
	tests := []struct {
		description   string
		uri           string
		body          string
		expected      interface{}
		expectedError error
	}{
		{
			description:   "it should have an error due to parsing of an invalid JSON",
			body:          "invalid",
			expectedError: fmt.Errorf("invalid character 'i' looking for beginning of value"),
		},
		{
			description:   "it should return an error due to parsing of an invalid URI scheme",
			uri:           "-",
			expectedError: fmt.Errorf("Get -: unsupported protocol scheme \"\""),
		},
		{
			description:   "it should return an error due to parsing of an invalid URI",
			uri:           "%gh&%ij",
			expectedError: fmt.Errorf("parse %%gh&%%ij: invalid URL escape \"%%gh\""),
		},
		{
			description: "it should fetch and unmarshal a JSON",
			body:        "{\"hello\":\"world\"}",
			expected:    map[string]string{"hello": "world"},
		},
	}

	for i, test := range tests {
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(test.body))
			}),
		)

		dir, err := ioutil.TempDir("/tmp", "rdap-test")

		if err != nil {
			t.Fatal(err)
		}

		c := NewClient(dir)
		obj := make(map[string]string)

		if test.uri == "" {
			test.uri = ts.URL
		}

		err = c.fetchAndUnmarshal(test.uri, &obj)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("At index %d (%s): expected error %s, got %s", i, test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expected, obj) {
				t.Fatalf("At index %d (%s): expected %v, got %v", i, test.description, test.expected, obj)
			}
		}
	}
}

func TestQuery(t *testing.T) {
	tests := []scenario{
		{
			description: "it should get an error when querying for a domain (invalid URI in bootstrap response)",
			kind:        dns,
			identifier:  "example.net",
			object:      &protocol.DomainResponse{},
			keepURIs:    true,
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{
					{
						{"net"},
						{"%%gh&%%ij%s", ""},
					},
				},
			},
			expectedError: fmt.Errorf("no data available for example.net"),
		},
		{
			description: "it should get an error when querying for a domain (no matches in bootstrap response)",
			kind:        dns,
			identifier:  "example.com",
			object:      &protocol.DomainResponse{},
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{{}},
			},
			expectedError: fmt.Errorf("no matches for example.com"),
		},
		{
			description: "it should get an error when querying for an AS number (invalid ASN range)",
			kind:        asn,
			identifier:  uint64(1),
			object:      &protocol.ASResponse{},
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{
					{
						{"2045-invalid"},
						{},
					},
				},
			},
			expectedError: fmt.Errorf("strconv.ParseUint: parsing \"invalid\": invalid syntax"),
		},
		{
			description:   "it should get an error when querying for an object (invalid endpoint URI)",
			kind:          asn,
			identifier:    uint64(1),
			object:        &protocol.ASResponse{},
			endpointURI:   "&&gh&&&ij%s",
			expectedError: fmt.Errorf("Get &&gh&&&ijasn: unsupported protocol scheme \"\""),
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var b []byte
				switch r.URL.Path {
				case fmt.Sprintf("/%s", test.kind):
					b, _ = json.Marshal(test.registry)
				case fmt.Sprintf("/%s/%v", test.kind, test.identifier):
					b, _ = json.Marshal(test.rdapObject)
				default:
					t.Fatal("not expecting uri", r.URL)
				}

				w.Write(b)
			}),
		)

		if test.registry != nil && len(test.registry.Services[0][1]) > 0 && !test.keepURIs {
			test.registry.Services[0][1][0] = fmt.Sprintf("%s/%s/%v", ts.URL, test.kind, test.identifier)
		}

		dir, err := ioutil.TempDir("/tmp", "rdap-test")

		if err != nil {
			t.Fatal(err)
		}

		c := NewClient(dir)

		if len(test.endpointURI) > 0 {
			c.SetRDAPEndpoint(test.endpointURI)
		} else {
			c.SetRDAPEndpoint(ts.URL + "/%v")
		}

		err = c.query(test.kind, test.identifier, test.object)

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error %s, got %s", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expected, test.object) {
				t.Fatalf("%s: expected %v, got %v", test.description, test.expected, test.object)
			}
		}
	}
}

func TestQueryByKind(t *testing.T) {
	tests := []scenario{
		{
			description: "it should get the right response when querying for a domain",
			kind:        dns,
			identifier:  "example.com",
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{
					{
						{"com"},
						{""}, // will be replaced by {ts.URL}/{test.kind}/{test.identifier} in awhile
					},
				},
			},
			rdapObject: protocol.DomainResponse{
				ObjectClassName: "test",
			},
			expected: &protocol.DomainResponse{
				ObjectClassName: "test",
			},
		},
		{
			description: "it should get the right response when querying for an AS number",
			kind:        asn,
			identifier:  uint64(123),
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{
					{
						{"100-200"},
						{""}, // will be replaced by {ts.URL}/{test.kind}/{test.identifier} in awhile
					},
				},
			},
			rdapObject: protocol.ASResponse{
				ObjectClassName: "test",
			},
			expected: &protocol.ASResponse{
				ObjectClassName: "test",
			},
		},
		{
			description: "it should get the right response when querying for an IP network",
			kind:        ip,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.0.2.1/25")
				return cidr
			}(),
			registry: &bootstrap.ServiceRegistry{
				Services: bootstrap.ServicesList{
					{
						{"192.0.2.0/24"},
						{""},
					},
				},
			},
			rdapObject: protocol.IPNetwork{
				ObjectClassName: "test",
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "test",
			},
		},
		{
			description:   "it should return an error due to invalid JSON in bootstrap response when querying for a domain",
			kind:          dns,
			identifier:    "example.com",
			registryBody:  "invalid",
			expectedError: fmt.Errorf("invalid character 'i' looking for beginning of value"),
		},
		{
			description:   "it should return an error due to invalid JSON in bootstrap response when querying for an AS number",
			kind:          asn,
			identifier:    uint64(1),
			registryBody:  "invalid",
			expectedError: fmt.Errorf("invalid character 'i' looking for beginning of value"),
		},
		{
			description: "it should return an error due to invalid JSON in bootstrap response when querying for an IP network",
			kind:        ip,
			identifier: func() *net.IPNet {
				_, cidr, _ := net.ParseCIDR("192.0.2.1/25")
				return cidr
			}(),
			registryBody:  "invalid",
			expectedError: fmt.Errorf("invalid character 'i' looking for beginning of value"),
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var b []byte
				switch r.URL.Path {
				case fmt.Sprintf("/%s", test.kind):
					if len(test.registryBody) > 0 {
						b = []byte(test.registryBody)
					} else {
						b, _ = json.Marshal(test.registry)
					}
				case fmt.Sprintf("/%s/%v", test.kind, test.identifier):
					b, _ = json.Marshal(test.rdapObject)
				default:
					t.Fatal("not expecting uri", r.URL)
				}

				w.Write(b)
			}),
		)

		if test.registry != nil && len(test.registry.Services[0][1]) > 0 && !test.keepURIs {
			test.registry.Services[0][1][0] = fmt.Sprintf("%s/%s", ts.URL, test.kind)
		}

		dir, err := ioutil.TempDir("/tmp", "rdap-test")

		if err != nil {
			t.Fatal(err)
		}

		c := NewClient(dir)
		c.SetRDAPEndpoint(ts.URL + "/%v")

		var r interface{}

		switch test.kind {
		case dns:
			r, err = c.QueryDomain(test.identifier.(string))
		case asn:
			r, err = c.QueryASN(test.identifier.(uint64))
		case ip:
			r, err = c.QueryIPNetwork(test.identifier.(*net.IPNet))
		}

		if test.expectedError != nil {
			if fmt.Sprintf("%v", test.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("%s: expected error %s, got %s", test.description, test.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(test.expected, r) {
				t.Fatalf("%s: expected %v, got %v", test.description, test.expected, r)
			}
		}
	}
}
