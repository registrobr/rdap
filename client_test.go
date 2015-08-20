package rdap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"testing"

	"github.com/registrobr/rdap/protocol"
)

func TestNewClient(t *testing.T) {
	client := NewClient([]string{"https://rdap.beta.registro.br"}, "200.160.2.3")
	if client.Transport == nil {
		t.Error("Not initializing direct RDAP query tranport layer")
	}
	if !reflect.DeepEqual(client.URIs, []string{"https://rdap.beta.registro.br"}) {
		t.Error("Not setting the URIs")
	}

	client = NewClient(nil, "200.160.2.3")
	if client.Transport == nil {
		t.Error("Not initializing bootstrap RDAP tranport layer")
	}
}

func TestClientDomain(t *testing.T) {
	data := []struct {
		description   string
		fqdn          string
		client        func() (*http.Response, error)
		expected      *protocol.Domain
		expectedError error
	}{
		{
			description: "it should return a valid domain",
			fqdn:        "example.com",
			client: func() (*http.Response, error) {
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
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the domain response",
			fqdn:        "example.com",
			client: func() (*http.Response, error) {
				var response http.Response
				response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
				return &response, nil
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if queryType != QueryTypeDomain {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeDomain, queryType)
				}

				if queryValue != item.fqdn {
					return nil, fmt.Errorf("expected FQDN “%s” and got “%s”", item.fqdn, queryValue)
				}

				return item.client()
			}),
		}

		domain, err := client.Domain(item.fqdn)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, domain) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, domain))
			}
		}
	}
}

func TestClientASN(t *testing.T) {
	data := []struct {
		description   string
		asn           uint32
		client        func() (*http.Response, error)
		expected      *protocol.AS
		expectedError error
	}{
		{
			description: "it should return a valid entity",
			asn:         1234,
			client: func() (*http.Response, error) {
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
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the AS response",
			asn:         1234,
			client: func() (*http.Response, error) {
				var response http.Response
				response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
				return &response, nil
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if queryType != QueryTypeAutnum {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeAutnum, queryType)
				}

				if queryValue != strconv.FormatUint(uint64(item.asn), 10) {
					return nil, fmt.Errorf("expected ASN “%d” and got “%s”", item.asn, queryValue)
				}

				return item.client()
			}),
		}

		as, err := client.ASN(item.asn)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, as) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, as))
			}
		}
	}
}

func TestClientEntity(t *testing.T) {
	data := []struct {
		description   string
		entity        string
		client        func() (*http.Response, error)
		expected      *protocol.Entity
		expectedError error
	}{
		{
			description: "it should return a valid entity",
			entity:      "h_005506560000136-NICBR",
			client: func() (*http.Response, error) {
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
			},
			expected: &protocol.Entity{
				ObjectClassName: "entity",
				Handle:          "005.506.560/0001-36",
			},
		},
		{
			description: "it should fail to query an entity",
			entity:      "h_005506560000136-NICBR",
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the entity response",
			entity:      "h_005506560000136-NICBR",
			client: func() (*http.Response, error) {
				var response http.Response
				response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
				return &response, nil
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if queryType != QueryTypeEntity {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeEntity, queryType)
				}

				if queryValue != item.entity {
					return nil, fmt.Errorf("expected entity “%s” and got “%s”", item.entity, queryValue)
				}

				return item.client()
			}),
		}

		entity, err := client.Entity(item.entity)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, entity) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, entity))
			}
		}
	}
}

func TestClientIPNetwork(t *testing.T) {
	data := []struct {
		description   string
		ipNetwork     *net.IPNet
		client        func() (*http.Response, error)
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
			client: func() (*http.Response, error) {
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
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
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
			client: func() (*http.Response, error) {
				var response http.Response
				response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
				return &response, nil
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if queryType != QueryTypeIP {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIP, queryType)
				}

				if queryValue != item.ipNetwork.String() {
					return nil, fmt.Errorf("expected IP network “%s” and got “%s”", item.ipNetwork, queryValue)
				}

				return item.client()
			}),
		}

		ipNetwork, err := client.IPNetwork(item.ipNetwork)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, ipNetwork) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, ipNetwork))
			}
		}
	}
}

func TestClientIP(t *testing.T) {
	data := []struct {
		description   string
		ip            net.IP
		client        func() (*http.Response, error)
		expected      *protocol.IPNetwork
		expectedError error
	}{
		{
			description: "it should return a valid IP network",
			ip:          net.ParseIP("200.160.2.3"),
			client: func() (*http.Response, error) {
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
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description: "it should fail to decode the IP network response",
			ip:          net.ParseIP("200.160.2.3"),
			client: func() (*http.Response, error) {
				var response http.Response
				response.Body = nopCloser{bytes.NewBufferString(`{{{{`)}
				return &response, nil
			},
			expectedError: fmt.Errorf("invalid character '{' looking for beginning of object key string"),
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if queryType != QueryTypeIP {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeIP, queryType)
				}

				if queryValue != item.ip.String() {
					return nil, fmt.Errorf("expected IP “%s” and got “%s”", item.ip, queryValue)
				}

				return item.client()
			}),
		}

		ipNetwork, err := client.IP(item.ip)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, ipNetwork) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, ipNetwork))
			}
		}
	}
}

func ExampleClient() {
	c := NewClient([]string{"https://rdap.beta.registro.br"}, "")

	d, err := c.Domain("nic.br")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v", d)
}

func ExampleBootstrapClient() {
	c := NewClient(nil, "")
	ip := net.ParseIP("214.1.2.3")

	ipnetwork, err := c.IP(ip)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v", ipnetwork)
}

func ExampleAdvancedBootstrapClient() {
	var httpClient http.Client

	cacheDetector := CacheDetector(func(resp *http.Response) bool {
		return resp.Header.Get("X-From-Cache") == "1"
	})

	c := Client{
		Transport: NewBootstrapFetcher(&httpClient, "", IANABootstrap, cacheDetector),
	}
	ip := net.ParseIP("214.1.2.3")

	ipnetwork, err := c.IP(ip)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%#v", ipnetwork)
}
