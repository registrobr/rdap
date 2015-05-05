package rdap

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

type scenario struct {
	description   string
	endpointURI   string
	kind          kind
	object        interface{}
	registry      *ServiceRegistry
	rdapObject    interface{}
	keepURIs      bool
	expected      *Response
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
			object:      "example.net",
			keepURIs:    true,
			registry: &ServiceRegistry{
				Services: ServicesList{
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
			object:      "example.com",
			registry: &ServiceRegistry{
				Services: ServicesList{{}},
			},
			expectedError: fmt.Errorf("no matches for example.com"),
		},
		{
			description: "it should get an error when querying for an AS number (invalid ASN range)",
			kind:        asn,
			object:      uint64(1),
			registry: &ServiceRegistry{
				Services: ServicesList{
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
			object:        uint64(1),
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
				case fmt.Sprintf("/%s/%v", test.kind, test.object):
					b, _ = json.Marshal(test.rdapObject)
				default:
					t.Fatal("not expecting uri", r.URL)
				}

				w.Write(b)
			}),
		)

		if test.registry != nil && len(test.registry.Services[0][1]) > 0 && !test.keepURIs {
			test.registry.Services[0][1][0] = fmt.Sprintf("%s/%s/%v", ts.URL, test.kind, test.object)
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

		r, err := c.query(test.kind, test.object)

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

func TestQueryByKind(t *testing.T) {
	tests := []scenario{
		{
			description: "it should get the right response when querying for a domain",
			kind:        dns,
			object:      "example.com",
			registry: &ServiceRegistry{
				Services: ServicesList{
					{
						{"com"},
						{""}, // will be replaced by {ts.URL}/{test.kind}/{test.object} in awhile
					},
				},
			},
			rdapObject: Response{
				Body: "hello, world",
			},
			expected: &Response{
				Body: "hello, world",
			},
		},
		{
			description: "it should get the right response when querying for an AS number",
			kind:        asn,
			object:      "123",
			registry: &ServiceRegistry{
				Services: ServicesList{
					{
						{"100-200"},
						{""}, // will be replaced by {ts.URL}/{test.kind}/{test.object} in awhile
					},
				},
			},
			rdapObject: Response{
				Body: "hello, world",
			},
			expected: &Response{
				Body: "hello, world",
			},
		},
		{
			description:   "it should return an error due to invalid AS number",
			kind:          asn,
			object:        "invalid",
			expectedError: fmt.Errorf("strconv.ParseUint: parsing \"invalid\": invalid syntax"),
		},
		{
			description: "it should get the right response when querying for an IP network",
			kind:        ipv4,
			object:      "192.0.2.1/25",
			registry: &ServiceRegistry{
				Services: ServicesList{
					{
						{"192.0.2.0/24"},
						{""},
					},
				},
			},
			rdapObject: Response{
				Body: "hello, world",
			},
			expected: &Response{
				Body: "hello, world",
			},
		},
		{
			description: "it should get the right response when querying for an IP network",
			kind:        ipv6,
			object:      "2001:0200:1000::/48",
			registry: &ServiceRegistry{
				Services: ServicesList{
					{
						{"2001:0200:1000::/36"},
						{""},
					},
				},
			},
			rdapObject: Response{
				Body: "hello, world",
			},
			expected: &Response{
				Body: "hello, world",
			},
		},
		{
			description:   "it should return an error due to invalid CIDR",
			kind:          ipv4,
			object:        "192.168.0.0/invalid",
			expectedError: fmt.Errorf("invalid CIDR address: 192.168.0.0/invalid"),
		},
	}

	for _, test := range tests {
		ts := httptest.NewServer(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var b []byte
				switch r.URL.Path {
				case fmt.Sprintf("/%s", test.kind):
					b, _ = json.Marshal(test.registry)
				case fmt.Sprintf("/%s/%v", test.kind, test.object):
					b, _ = json.Marshal(test.rdapObject)
				default:
					t.Fatal("not expecting uri", r.URL)
				}

				w.Write(b)
			}),
		)

		if test.registry != nil && len(test.registry.Services[0][1]) > 0 && !test.keepURIs {
			test.registry.Services[0][1][0] = fmt.Sprintf("%s/%s/%v", ts.URL, test.kind, test.object)
		}

		dir, err := ioutil.TempDir("/tmp", "rdap-test")

		if err != nil {
			t.Fatal(err)
		}

		c := NewClient(dir)
		c.SetRDAPEndpoint(ts.URL + "/%v")

		var (
			r      *Response
			object = test.object.(string)
		)

		switch test.kind {
		case dns:
			r, err = c.QueryDomain(object)
		case asn:
			r, err = c.QueryASN(object)
		case ipv4, ipv6:
			r, err = c.QueryIPNetwork(object)
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
