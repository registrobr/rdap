package rdap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/registrobr/rdap/protocol"
)

type httpClientFunc func(*http.Request) (*http.Response, error)

func (h httpClientFunc) Do(r *http.Request) (*http.Response, error) {
	return h(r)
}

func TestDefaultFetcherFetch(t *testing.T) {
	data := []struct {
		description   string
		uris          []string
		queryType     QueryType
		queryValue    string
		xForwardedFor string
		httpClient    HTTPClient
		expected      *http.Response
		expectedError error
	}{
		{
			description:   "it should fetch correctly",
			uris:          []string{"https://rdap.beta.registro.br"},
			queryType:     QueryTypeDomain,
			queryValue:    "example.com",
			xForwardedFor: "200.160.2.3",
			httpClient: httpClientFunc(func(r *http.Request) (*http.Response, error) {
				expectedURL := "https://rdap.beta.registro.br/domain/example.com"
				if r.URL.String() != expectedURL {
					return nil, fmt.Errorf("expected url “%s” and got “%s”", expectedURL, r.URL.String())
				}

				expectedXForwardedFor := "200.160.2.3"
				if r.Header.Get("X-Forwarded-For") != expectedXForwardedFor {
					return nil, fmt.Errorf("expected HTTP header X-Forwarded-For “%s” and got “%s”", expectedXForwardedFor, r.Header.Get("X-Forwarded-For"))
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
				response.StatusCode = http.StatusOK
				response.Header = http.Header{
					"Content-Type": []string{"application/rdap+json"},
				}
				response.Body = nopCloser{bytes.NewBuffer(data)}
				return &response, nil
			}),
			expected: func() *http.Response {
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
				response.StatusCode = http.StatusOK
				response.Header = http.Header{
					"Content-Type": []string{"application/rdap+json"},
				}
				response.Body = nopCloser{bytes.NewBuffer(data)}
				return &response
			}(),
		},
	}

	for i, item := range data {
		fetcher := NewDefaultFetcher(item.httpClient, item.xForwardedFor)
		response, err := fetcher.Fetch(item.uris, item.queryType, item.queryValue)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, response) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, response))
			}
		}
	}
}

func TestBootstrap(t *testing.T) {
	data := []struct {
		description   string
		uris          []string
		queryType     QueryType
		queryValue    string
		httpClient    HTTPClient
		xForwardedFor string
		bootstrapURI  string
		cacheDetector CacheDetector
		expected      *http.Response
		expectedError error
	}{}

	for i, item := range data {
		fetcher := NewBootstrapFetcher(item.httpClient, item.xForwardedFor, item.bootstrapURI, item.cacheDetector)
		response, err := fetcher.Fetch(item.uris, item.queryType, item.queryValue)

		if item.expectedError != nil {
			if fmt.Sprintf("%v", item.expectedError) != fmt.Sprintf("%v", err) {
				t.Fatalf("[%d] %s: expected error “%s”, got “%s”", i, item.description, item.expectedError, err)
			}
		} else {
			if !reflect.DeepEqual(item.expected, response) {
				t.Fatalf("[%d] “%s”: mismatch results.\n%v", i, item.description, diff(item.expected, response))
			}
		}
	}
}
