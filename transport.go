package rdap

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"

	"github.com/registrobr/rdap/protocol"
)

const (
	QueryTypeDomain    QueryType = "domain"
	QueryTypeAutnum    QueryType = "autnum"
	QueryTypeIP        QueryType = "ip"
	QueryTypeIPNetwork QueryType = "ipnetwork"
	QueryTypeEntity    QueryType = "entity"
)

type QueryType string

var (
	// ErrNotFound is used when the RDAP server doesn't contain any
	// information of the requested object
	ErrNotFound = errors.New("not found")
)

type Fetcher interface {
	Fetch(uris []string, queryType QueryType, queryValue string) (*http.Response, error)
}

// fetcherFunc is a function type that implements the Fetcher interface
type fetcherFunc func([]string, QueryType, string) (*http.Response, error)

func (f fetcherFunc) Fetch(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
	return f(uris, queryType, queryValue)
}

type decorator func(Fetcher) Fetcher

func decorate(f Fetcher, ds ...decorator) Fetcher {
	for _, decorate := range ds {
		f = decorate(f)
	}

	return f
}

type CacheDetector func(*http.Response) bool

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type defaultFetcher struct {
	httpClient    HTTPClient
	xForwardedFor string
}

func NewDefaultFetcher(httpClient HTTPClient, xForwardedFor string) Fetcher {
	return &defaultFetcher{
		httpClient:    httpClient,
		xForwardedFor: xForwardedFor,
	}
}

func (d *defaultFetcher) Fetch(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
	var lastErr error

	for _, uri := range uris {
		uri = fmt.Sprintf("%s/%s/%s", uri, queryType, queryValue)

		req, err := http.NewRequest("GET", uri, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Add("Accept", "application/json")

		if d.xForwardedFor != "" {
			req.Header.Add("X-Forwarded-For", d.xForwardedFor)
		}

		resp, err := d.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode == http.StatusNotFound {
			lastErr = ErrNotFound
			continue
		}

		if resp.Header.Get("Content-Type") != "application/rdap+json" {
			lastErr = fmt.Errorf("unexpected response: %d %s",
				resp.StatusCode, http.StatusText(resp.StatusCode))
		}

		if resp.StatusCode != http.StatusOK {
			var responseErr protocol.Error
			if err := json.NewDecoder(resp.Body).Decode(&responseErr); err != nil {
				lastErr = err
				continue
			}
			lastErr = responseErr
			continue
		}

		return resp, nil
	}

	return nil, lastErr
}

func NewBootstrapFetcher(httpClient HTTPClient, xForwardedFor string, bootstrapURI string, cacheDetector CacheDetector) Fetcher {
	return decorate(
		NewDefaultFetcher(httpClient, xForwardedFor),
		bootstrap(bootstrapURI, httpClient, cacheDetector),
	)
}

func bootstrap(bootstrapURI string, httpClient HTTPClient, cacheDetector CacheDetector) decorator {
	return func(f Fetcher) Fetcher {
		return fetcherFunc(func(uris []string, queryType QueryType, queryValue string) (*http.Response, error) {
			bootstrapURI := fmt.Sprintf("%s/%s", bootstrapURI, queryType)

			serviceRegistry, cached, err := bootstrapFetch(httpClient, bootstrapURI, false, cacheDetector)
			if err != nil {
				return nil, err
			}

			switch queryType {
			case QueryTypeDomain:
				uris, err = serviceRegistry.matchDomain(queryValue)
				if err == nil && len(uris) == 0 && cached {
					var nsSet []*net.NS
					if nsSet, err = net.LookupNS(queryValue); err == nil && len(nsSet) > 0 {
						serviceRegistry, cached, err = bootstrapFetch(httpClient, bootstrapURI, true, cacheDetector)
						if err == nil {
							uris, err = serviceRegistry.matchDomain(queryValue)
						}
					}
				}

			case QueryTypeAutnum:
				asn, err := strconv.ParseUint(queryValue, 10, 32)
				if err == nil {
					uris, err = serviceRegistry.matchAS(uint32(asn))
				}

			case QueryTypeIP:
				ip := net.ParseIP(queryValue)
				if ip != nil {
					uris, err = serviceRegistry.matchIP(ip)
				}

			case QueryTypeIPNetwork:
				_, cidr, err := net.ParseCIDR(queryValue)
				if err == nil {
					uris, err = serviceRegistry.matchIPNetwork(cidr)
				}
			}

			if err != nil {
				return nil, err
			}

			if len(uris) == 0 {
				return nil, fmt.Errorf("no matches for %v", queryValue)
			}

			sort.Sort(prioritizeHTTPS(uris))
			return f.Fetch(uris, queryType, queryValue)
		})
	}
}

func bootstrapFetch(httpClient HTTPClient, uri string, reloadCache bool, cacheDetector CacheDetector) (*serviceRegistry, bool, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Add("Accept", "application/json")

	if reloadCache {
		req.Header.Add("Cache-Control", "max-age=0")
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}

	cached := cacheDetector(resp)

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotModified {
		return nil, cached, fmt.Errorf("unexpected status code %d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}

	var serviceRegistry serviceRegistry
	if err := json.NewDecoder(resp.Body).Decode(&serviceRegistry); err != nil {
		return nil, cached, err
	}

	if serviceRegistry.Version != version {
		return nil, false, fmt.Errorf("incompatible bootstrap specification version: %s (expecting %s)", serviceRegistry.Version, version)
	}

	return &serviceRegistry, cached, nil
}
