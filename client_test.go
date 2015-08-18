package rdap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/registrobr/rdap/protocol"
)

func TestClientDomain(t *testing.T) {
	data := []struct {
		description   string
		fqdn          string
		client        Client
		expected      *protocol.Domain
		expectedError error
	}{
		{
			description: "it should return a valid domain",
			fqdn:        "example.com",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeDomain {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeDomain, queryType)
					}

					expectedFQDN := "example.com"
					if queryValue != expectedFQDN {
						return nil, fmt.Errorf("expected FQDN “%s” and got “%s”", expectedFQDN, queryValue)
					}

					domain := protocol.Domain{
						ObjectClassName: "domain",
						Handle:          "example.com",
						LDHName:         "example.com",
					}

					data, err := json.Marshal(domain)
					if err != nil {
						t.Fatal(err)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBuffer(data)}
					return &response, nil
				}),
			},
			expected: &protocol.Domain{
				ObjectClassName: "domain",
				Handle:          "example.com",
				LDHName:         "example.com",
			},
		},
		{
			description: "it should fail to query a domain",
			fqdn:        "example.com",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeDomain {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeDomain, queryType)
					}

					expectedFQDN := "example.com"
					if queryValue != expectedFQDN {
						return nil, fmt.Errorf("expected FQDN “%s” and got “%s”", expectedFQDN, queryValue)
					}

					return nil, fmt.Errorf("I'm a crazy error!")
				}),
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the domain response",
			fqdn:        "example.com",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeDomain {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeDomain, queryType)
					}

					expectedFQDN := "example.com"
					if queryValue != expectedFQDN {
						return nil, fmt.Errorf("expected FQDN “%s” and got “%s”", expectedFQDN, queryValue)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
					return &response, nil
				}),
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		domain, err := item.client.Domain(item.fqdn)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, domain) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, domain))
			}
		}
	}
}

func TestClientASN(t *testing.T) {
	data := []struct {
		description   string
		asn           uint32
		client        Client
		expected      *protocol.AS
		expectedError error
	}{
		{
			description: "it should return a valid entity",
			asn:         1234,
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeAutnum {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeAutnum, queryType)
					}

					expectedASN := "1234"
					if queryValue != expectedASN {
						return nil, fmt.Errorf("expected ASN “%s” and got “%s”", expectedASN, queryValue)
					}

					as := protocol.AS{
						ObjectClassName: "autnum",
						Handle:          "1234",
						StartAutnum:     1234,
						EndAutnum:       1234,
					}

					data, err := json.Marshal(as)
					if err != nil {
						t.Fatal(err)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBuffer(data)}
					return &response, nil
				}),
			},
			expected: &protocol.AS{
				ObjectClassName: "autnum",
				Handle:          "1234",
				StartAutnum:     1234,
				EndAutnum:       1234,
			},
		},
		{
			description: "it should fail to query an ASN",
			asn:         1234,
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeAutnum {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeAutnum, queryType)
					}

					expectedASN := "1234"
					if queryValue != expectedASN {
						return nil, fmt.Errorf("expected ASN “%s” and got “%s”", expectedASN, queryValue)
					}

					return nil, fmt.Errorf("I'm a crazy error!")
				}),
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the AS response",
			asn:         1234,
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeAutnum {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeAutnum, queryType)
					}

					expectedASN := "1234"
					if queryValue != expectedASN {
						return nil, fmt.Errorf("expected ASN “%s” and got “%s”", expectedASN, queryValue)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
					return &response, nil
				}),
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		as, err := item.client.ASN(item.asn)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, as) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, as))
			}
		}
	}
}

func TestClientEntity(t *testing.T) {
	data := []struct {
		description   string
		entity        string
		client        Client
		expected      *protocol.Entity
		expectedError error
	}{
		{
			description: "it should return a valid entity",
			entity:      "h_005506560000136-NICBR",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeEntity {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeEntity, queryType)
					}

					expectedEntity := "h_005506560000136-NICBR"
					if queryValue != expectedEntity {
						return nil, fmt.Errorf("expected entity “%s” and got “%s”", expectedEntity, queryValue)
					}

					entity := protocol.Entity{
						ObjectClassName: "entity",
						Handle:          "005.506.560/0001-36",
					}

					data, err := json.Marshal(entity)
					if err != nil {
						t.Fatal(err)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBuffer(data)}
					return &response, nil
				}),
			},
			expected: &protocol.Entity{
				ObjectClassName: "entity",
				Handle:          "005.506.560/0001-36",
			},
		},
		{
			description: "it should fail to query an entity",
			entity:      "h_005506560000136-NICBR",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeEntity {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeEntity, queryType)
					}

					expectedEntity := "h_005506560000136-NICBR"
					if queryValue != expectedEntity {
						return nil, fmt.Errorf("expected entity “%s” and got “%s”", expectedEntity, queryValue)
					}

					return nil, fmt.Errorf("I'm a crazy error!")
				}),
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the entity response",
			entity:      "h_005506560000136-NICBR",
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeEntity {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeEntity, queryType)
					}

					expectedEntity := "h_005506560000136-NICBR"
					if queryValue != expectedEntity {
						return nil, fmt.Errorf("expected entity “%s” and got “%s”", expectedEntity, queryValue)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
					return &response, nil
				}),
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		entity, err := item.client.Entity(item.entity)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, entity) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, entity))
			}
		}
	}
}

func TestClientIPNetwork(t *testing.T) {
	data := []struct {
		description   string
		ipNetwork     *net.IPNet
		client        Client
		expected      *protocol.IPNetwork
		expectedError error
	}{
		{
			description: "it should return a valid IP network",
			ipNetwork: func() *net.IPNet {
				_, ipNetwork, err := net.ParseCIDR("200.160.0.0/20")
				if err != nil {
					t.Fatal(err)
				}

				return ipNetwork
			}(),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIPNetwork {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIPNetwork, queryType)
					}

					expectedIPNetwork := "200.160.0.0/20"
					if queryValue != expectedIPNetwork {
						return nil, fmt.Errorf("expected IP network “%s” and got “%s”", expectedIPNetwork, queryValue)
					}

					ipNetwork := protocol.IPNetwork{
						ObjectClassName: "ipnetwork",
						Handle:          "200.160.0.0/20",
					}

					data, err := json.Marshal(ipNetwork)
					if err != nil {
						t.Fatal(err)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBuffer(data)}
					return &response, nil
				}),
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
		},
		{
			description:   "it should fail for a nil input",
			expectedError: fmt.Errorf("undefined IP network"),
		},
		{
			description: "it should fail to query an IP network",
			ipNetwork: func() *net.IPNet {
				_, ipNetwork, err := net.ParseCIDR("200.160.0.0/20")
				if err != nil {
					t.Fatal(err)
				}

				return ipNetwork
			}(),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIPNetwork {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIPNetwork, queryType)
					}

					expectedIPNetwork := "200.160.0.0/20"
					if queryValue != expectedIPNetwork {
						return nil, fmt.Errorf("expected IP network “%s” and got “%s”", expectedIPNetwork, queryValue)
					}

					return nil, fmt.Errorf("I'm a crazy error!")
				}),
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the IP network response",
			ipNetwork: func() *net.IPNet {
				_, ipNetwork, err := net.ParseCIDR("200.160.0.0/20")
				if err != nil {
					t.Fatal(err)
				}

				return ipNetwork
			}(),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIPNetwork {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIPNetwork, queryType)
					}

					expectedIPNetwork := "200.160.0.0/20"
					if queryValue != expectedIPNetwork {
						return nil, fmt.Errorf("expected IP network “%s” and got “%s”", expectedIPNetwork, queryValue)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
					return &response, nil
				}),
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		ipNetwork, err := item.client.IPNetwork(item.ipNetwork)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, ipNetwork) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, ipNetwork))
			}
		}
	}
}

func TestClientIP(t *testing.T) {
	data := []struct {
		description   string
		ip            net.IP
		client        Client
		expected      *protocol.IPNetwork
		expectedError error
	}{
		{
			description: "it should return a valid IP network",
			ip:          net.ParseIP("200.160.2.3"),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIP {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIP, queryType)
					}

					expectedIP := "200.160.2.3"
					if queryValue != expectedIP {
						return nil, fmt.Errorf("expected IP “%s” and got “%s”", expectedIP, queryValue)
					}

					ipNetwork := protocol.IPNetwork{
						ObjectClassName: "ipnetwork",
						Handle:          "200.160.0.0/20",
					}

					data, err := json.Marshal(ipNetwork)
					if err != nil {
						t.Fatal(err)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBuffer(data)}
					return &response, nil
				}),
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
		},
		{
			description:   "it should fail for a nil input",
			expectedError: fmt.Errorf("undefined IP"),
		},
		{
			description: "it should fail to query an IP network",
			ip:          net.ParseIP("200.160.2.3"),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIP {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIP, queryType)
					}

					expectedIP := "200.160.2.3"
					if queryValue != expectedIP {
						return nil, fmt.Errorf("expected IP “%s” and got “%s”", expectedIP, queryValue)
					}

					return nil, fmt.Errorf("I'm a crazy error!")
				}),
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the IP network response",
			ip:          net.ParseIP("200.160.2.3"),
			client: Client{
				URIs: []string{"rdap.example.com"},
				Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
					expectedURIs := []string{"rdap.example.com"}
					if !reflect.DeepEqual(expectedURIs, uris) {
						return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
					}

					if queryType != QueryTypeIP {
						return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIP, queryType)
					}

					expectedIP := "200.160.2.3"
					if queryValue != expectedIP {
						return nil, fmt.Errorf("expected IP “%s” and got “%s”", expectedIP, queryValue)
					}

					var response http.Response
					response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
					return &response, nil
				}),
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		ipNetwork, err := item.client.IP(item.ip)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, ipNetwork) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, ipNetwork))
			}
		}
	}
}
