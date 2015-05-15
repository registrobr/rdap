package bootstrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"
)

type kind string

const (
	dns           kind = "dns"
	asn           kind = "asn"
	ipv4          kind = "ipv4"
	ipv6          kind = "ipv6"
	RDAPBootstrap      = "https://data.iana.org/rdap/%s.json"
)

type Client struct {
	httpClient *http.Client
	Bootstrap  string
}

func NewClient(httpClient *http.Client) *Client {
	return &Client{
		Bootstrap:  RDAPBootstrap,
		httpClient: httpClient,
	}
}

func (c *Client) Domain(fqdn string) ([]string, error) {
	return c.query(dns, fqdn)
}

func (c *Client) ASN(as uint64) ([]string, error) {
	return c.query(asn, as)
}

func (c *Client) IPNetwork(ipnet *net.IPNet) ([]string, error) {
	kind := ipv4

	if ipnet.IP.To4() == nil {
		kind = ipv6
	}

	return c.query(kind, ipnet)
}

func (c *Client) query(kind kind, identifier interface{}) ([]string, error) {
	uris := []string{}
	r := serviceRegistry{}
	uri := fmt.Sprintf(c.Bootstrap, kind)
	body, err := c.fetch(uri)

	if err != nil {
		return nil, err
	}

	defer body.Close()

	if err := json.NewDecoder(body).Decode(&r); err != nil {
		return nil, err
	}

	if r.Version != version {
		return nil, fmt.Errorf("incompatible bootstrap specification version: %s (expecting %s)", r.Version, version)
	}

	switch kind {
	case dns:
		uris, err = r.MatchDomain(identifier.(string))
	case asn:
		uris, err = r.MatchAS(identifier.(uint64))
	case ipv4, ipv6:
		uris, err = r.MatchIPNetwork(identifier.(*net.IPNet))
	}

	if err != nil {
		return nil, err
	}

	if len(uris) == 0 {
		return nil, fmt.Errorf("no matches for %v", identifier)
	}

	sort.Sort(prioritizeHTTPS(uris))

	return uris, nil
}

func (c *Client) fetch(uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, err
	}

	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}

	resp, err := c.httpClient.Do(req)

	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}
