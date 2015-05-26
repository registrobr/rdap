package bootstrap

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sort"

	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/gregjones/httpcache"
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
	cacheKey   string
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

func (c *Client) IP(ip net.IP) ([]string, error) {
	kind := ipv4

	if ip.To4() == nil {
		kind = ipv6
	}

	return c.query(kind, ip)
}

func (c *Client) CheckDomain(fqdn string, cached bool) (uris []string, err error) {
	r := serviceRegistry{}
	uris, err = r.MatchDomain(fqdn)
	if err != nil {
		return
	}

	if len(uris) > 0 {
		return
	}

	if !cached {
		return
	}

	nsSet, err := net.LookupNS(fqdn)
	if err != nil {
		return nil, nil
	}

	if len(nsSet) == 0 {
		return nil, nil
	}

	transport := c.httpClient.Transport.(*httpcache.Transport)
	transport.Cache.Delete(c.cacheKey)
	c.httpClient.Transport = transport

	body, cached, err := c.fetch(c.Bootstrap + string(dns))
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(&r); err != nil {
		return nil, err
	}

	return r.MatchDomain(fqdn)
}

func (c *Client) query(kind kind, identifier interface{}) ([]string, error) {
	uris := []string{}
	r := serviceRegistry{}
	uri := fmt.Sprintf(c.Bootstrap, kind)
	body, cached, err := c.fetch(uri)

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
		uris, err = c.CheckDomain(identifier.(string), cached)
	case asn:
		uris, err = r.MatchAS(identifier.(uint64))
	case ipv4, ipv6:
		if ip, ok := identifier.(net.IP); ok {
			uris, err = r.MatchIP(ip)
			break
		}
		if ipNet, ok := identifier.(*net.IPNet); ok {
			uris, err = r.MatchIPNetwork(ipNet)
			break
		}
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

func (c *Client) fetch(uri string) (_ io.ReadCloser, cached bool, err error) {
	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return nil, cached, err
	}

	c.cacheKey = req.URL.String()

	if c.httpClient == nil {
		c.httpClient = &http.Client{}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, cached, err
	}

	if resp.Header.Get(httpcache.XFromCache) == "1" {
		cached = true
	}

	return resp.Body, cached, nil
}
