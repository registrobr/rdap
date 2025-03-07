package rdap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/registrobr/rdap/protocol"
)

func TestNewClient(t *testing.T) {
	client := NewClient([]string{"https://rdap.beta.registro.br"})
	if client.Transport == nil {
		t.Error("Not initializing direct RDAP query tranport layer")
	}
	if !reflect.DeepEqual(client.URIs, []string{"https://rdap.beta.registro.br"}) {
		t.Error("Not setting the URIs")
	}

	client = NewClient(nil)
	if client.Transport == nil {
		t.Error("Not initializing bootstrap RDAP tranport layer")
	}
}

func TestClientDomain(t *testing.T) {
	data := []struct {
		description    string
		fqdn           string
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.Domain
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description: "it should return a valid domain",
			fqdn:        "example.com",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"ticket": []string{"1234"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.Domain{
				ObjectClassName: "domain",
				Handle:          "example.com",
				LDHName:         "example.com",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should detect an invalid unicode domain",
			fqdn:        "xn--東京\uffff!!@...-.jp",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"ticket": []string{"1234"},
			},
			expectedError: fmt.Errorf(`idna: invalid label "東京\uffff!!@"`),
		},
		{
			description: "it should fail to query a domain",
			fqdn:        "example.com",
			client: func() (*http.Response, error) {
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query a domain with no response",
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
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

		domain, header, err := client.Domain(item.fqdn, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientTicket(t *testing.T) {
	data := []struct {
		description    string
		ticketNumber   int
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.Domain
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description:  "it should return a valid ticket",
			ticketNumber: 1234,
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.Domain{
				ObjectClassName: "domain",
				Handle:          "example.com",
				LDHName:         "example.com",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description:  "it should fail to query a ticket",
			ticketNumber: 1234,
			client: func() (*http.Response, error) {
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description:  "it should fail to query a ticket with no response",
			ticketNumber: 1234,
			client: func() (*http.Response, error) {
				return nil, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
		},
		{
			description:  "it should fail to decode the domain response",
			ticketNumber: 1234,
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
				}

				if queryType != QueryTypeTicket {
					return nil, fmt.Errorf("expected query type “%s” and got “%s”", QueryTypeTicket, queryType)
				}

				if queryValue != strconv.Itoa(item.ticketNumber) {
					return nil, fmt.Errorf("expected FQDN “%d” and got “%s”", item.ticketNumber, queryValue)
				}

				return item.client()
			}),
		}

		domain, header, err := client.Ticket(item.ticketNumber, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientASN(t *testing.T) {
	data := []struct {
		description    string
		asn            uint32
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.AS
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description: "it should return a valid AS",
			asn:         1234,
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"full": nil,
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.AS{
				ObjectClassName: "autnum",
				Handle:          "1234",
				StartAutnum:     1234,
				EndAutnum:       1234,
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an ASN",
			asn:         1234,
			client: func() (*http.Response, error) {
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an ASN with no response",
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
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

		as, header, err := client.ASN(item.asn, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientEntity(t *testing.T) {
	data := []struct {
		description    string
		entity         string
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.Entity
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description: "it should return a valid entity",
			entity:      "h_005506560000136-NICBR",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"full": nil,
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.Entity{
				ObjectClassName: "entity",
				Handle:          "005.506.560/0001-36",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an entity",
			entity:      "h_005506560000136-NICBR",
			client: func() (*http.Response, error) {
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an entity with no response",
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
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

		entity, header, err := client.Entity(item.entity, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientIPNetwork(t *testing.T) {
	data := []struct {
		description    string
		ipNetwork      *net.IPNet
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.IPNetwork
		expectedHeader http.Header
		expectedError  error
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
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"full": nil,
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
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
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an IP network with no response",
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
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

		ipNetwork, header, err := client.IPNetwork(item.ipNetwork, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientIP(t *testing.T) {
	data := []struct {
		description    string
		ip             net.IP
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       *protocol.IPNetwork
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description: "it should return a valid IP network",
			ip:          net.ParseIP("200.160.2.3"),
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"full": nil,
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description:   "it should fail for a nil input",
			expectedError: fmt.Errorf("undefined IP"),
		},
		{
			description: "it should fail to query an IP",
			ip:          net.ParseIP("200.160.2.3"),
			client: func() (*http.Response, error) {
				var response http.Response
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, fmt.Errorf("I'm a crazy error!")
			},
			expectedError: fmt.Errorf("I'm a crazy error!"),
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should fail to query an IP with no response",
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
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				expectedURIs := []string{"rdap.example.com"}
				if !reflect.DeepEqual(expectedURIs, uris) {
					return nil, fmt.Errorf("expected uris “%#v” and got “%#v”", expectedURIs, uris)
				}

				if !reflect.DeepEqual(item.header, header) {
					return nil, fmt.Errorf("expected HTTP headers “%#v” and got “%#v”", item.header, header)
				}

				if !reflect.DeepEqual(item.queryString, queryString) {
					return nil, fmt.Errorf("expected query string “%#v” and got “%#v”", item.queryString, queryString)
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

		ipNetwork, header, err := client.IP(item.ip, item.header, item.queryString)

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

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func TestClientQuery(t *testing.T) {
	data := []struct {
		description    string
		object         string
		header         http.Header
		queryString    url.Values
		client         func() (*http.Response, error)
		expected       interface{}
		expectedHeader http.Header
		expectedError  error
	}{
		{
			description: "it should return a valid domain",
			object:      "example.com",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
			queryString: url.Values{
				"ticket": []string{"1234"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.Domain{
				ObjectClassName: "domain",
				Handle:          "example.com",
				LDHName:         "example.com",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should return a valid AS",
			object:      "1234",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.AS{
				ObjectClassName: "autnum",
				Handle:          "1234",
				StartAutnum:     1234,
				EndAutnum:       1234,
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should return a valid entity",
			object:      "h_005506560000136-NICBR",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.Entity{
				ObjectClassName: "entity",
				Handle:          "005.506.560/0001-36",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should return a valid IP network",
			object:      "200.160.0.0/20",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
		{
			description: "it should return a valid IP network",
			object:      "200.160.2.3",
			header: http.Header{
				"X-Forwarded-For": []string{"127.0.0.1"},
			},
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
				response.Header = http.Header{
					"Random-Header": []string{"value"},
				}
				return &response, nil
			},
			expected: &protocol.IPNetwork{
				ObjectClassName: "ipnetwork",
				Handle:          "200.160.0.0/20",
			},
			expectedHeader: http.Header{
				"Random-Header": []string{"value"},
			},
		},
	}

	for i, item := range data {
		client := Client{
			URIs: []string{"rdap.example.com"},
			Transport: fetcherFunc(func(uris []string, queryType QueryType, queryValue string, header http.Header, queryString url.Values) (*http.Response, error) {
				return item.client()
			}),
		}

		resp, header, err := client.Query(item.object, item.header, item.queryString)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Errorf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}

		} else if err != nil {
			t.Errorf("[%d] %s: unexpected error “%s”", i, item.description, err)

		} else {
			if !reflect.DeepEqual(item.expected, resp) {
				t.Errorf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, resp))
			}

			if !reflect.DeepEqual(item.expectedHeader, header) {
				t.Errorf("[%d] “%s”: mismatch HTTP header.\n%v", i, item.description, diff(item.expectedHeader, header))
			}
		}
	}
}

func ExampleClient() {
	c := NewClient([]string{"https://rdap.beta.registro.br"})

	d, _, err := c.Query("nic.br", nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	output, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))

	// Another example for a direct domain query adding a "ticket" parameter

	queryString := make(url.Values)
	queryString.Set("ticket", "5439886")

	d, _, err = c.Domain("rafael.net.br", nil, queryString)
	if err != nil {
		fmt.Println(err)
		return
	}

	output, err = json.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}

func ExampleClient_bootstrap() {
	c := NewClient(nil)

	ipnetwork, _, err := c.Query("214.1.2.3", nil, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	output, err := json.MarshalIndent(ipnetwork, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}

func ExampleClient_aAdvancedBootstrap() {
	var httpClient http.Client

	cacheDetector := CacheDetector(func(resp *http.Response) bool {
		return resp.Header.Get("X-From-Cache") == "1"
	})

	c := Client{
		Transport: NewBootstrapFetcher(&httpClient, IANABootstrap, cacheDetector),
	}

	ipnetwork, _, err := c.Query("214.1.2.3", http.Header{
		"X-Forwarded-For": []string{"127.0.0.1"},
	}, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	output, err := json.MarshalIndent(ipnetwork, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}
