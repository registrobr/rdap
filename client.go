package rdap

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"

	"github.com/registrobr/rdap-client/protocol"

	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/gregjones/httpcache"
	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/gregjones/httpcache/diskcache"
)

const (
	ianaRDAPEndpoint = "https://data.iana.org/rdap/%v.json"
)

type kind string

const (
	dns  kind = "dns"
	asn  kind = "asn"
	ipv4 kind = "ipv4"
	ipv6 kind = "ipv6"
)

type Client struct {
	cacheDir     string
	rdapEndpoint string
}

func NewClient(cacheDir string) *Client {
	return &Client{
		cacheDir:     cacheDir,
		rdapEndpoint: ianaRDAPEndpoint,
	}
}

func (c *Client) SetRDAPEndpoint(uri string) {
	c.rdapEndpoint = uri
}

func (c *Client) QueryDomain(fqdn string) (*protocol.DomainResponse, error) {
	r := &protocol.DomainResponse{}

	if err := c.query(dns, fqdn, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) QueryASN(number string) (*protocol.ASResponse, error) {
	as, err := strconv.ParseUint(number, 10, 32)

	if err != nil {
		return nil, err
	}

	r := &protocol.ASResponse{}

	if err := c.query(asn, as, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) QueryIPNetwork(ipnet string) (*protocol.IPNetwork, error) {
	_, cidr, err := net.ParseCIDR(ipnet)

	if err != nil {
		return nil, err
	}

	kind := ipv4

	if cidr.IP.To4() == nil {
		kind = ipv6
	}

	r := &protocol.IPNetwork{}

	if err := c.query(kind, cidr, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) query(kind kind, identifier interface{}, object interface{}) error {
	var (
		err  error
		uris Values
		uri  = fmt.Sprintf(c.rdapEndpoint, kind)
		r    = ServiceRegistry{}
	)

	if err := c.fetchAndUnmarshal(uri, &r); err != nil {
		return err
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
		return err
	}

	if len(uris) == 0 {
		return fmt.Errorf("no matches for %v", identifier)
	}

	sort.Sort(uris)

	for _, uri := range uris {
		if err := c.fetchAndUnmarshal(uri, object); err != nil {
			continue
		}

		return nil
	}

	return fmt.Errorf("no data available for %v", identifier)
}

func (c *Client) fetchAndUnmarshal(uri string, object interface{}) error {
	cli := http.Client{
		Transport: httpcache.NewTransport(
			diskcache.New(c.cacheDir),
		),
	}

	req, err := http.NewRequest("GET", uri, nil)

	if err != nil {
		return err
	}

	resp, err := cli.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&object); err != nil {
		return err
	}

	return nil
}
