package rdap

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"

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

// TODO Replace by a protocol object
type Response struct {
	Body string
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

func (c *Client) QueryDomain(fqdn string) (*Response, error) {
	return c.query(dns, fqdn)
}

func (c *Client) QueryASN(number string) (*Response, error) {
	as, err := strconv.ParseUint(number, 10, 32)

	if err != nil {
		return nil, err
	}

	return c.query(asn, as)
}

func (c *Client) QueryIPNetwork(ipnet string) (*Response, error) {
	_, cidr, err := net.ParseCIDR(ipnet)

	if err != nil {
		return nil, err
	}

	kind := ipv4

	if cidr.IP.To4() == nil {
		kind = ipv6
	}

	return c.query(kind, cidr)
}

func (c *Client) query(kind kind, object interface{}) (*Response, error) {
	var (
		err  error
		uris Values
		uri  = fmt.Sprintf(c.rdapEndpoint, kind)
		r    = ServiceRegistry{}
	)

	if err := c.fetchAndUnmarshal(uri, &r); err != nil {
		return nil, err
	}

	switch kind {
	case dns:
		uris, err = r.MatchDomain(object.(string))
	case asn:
		uris, err = r.MatchAS(object.(uint64))
	case ipv4, ipv6:
		uris, err = r.MatchIPNetwork(object.(*net.IPNet))
	}

	if err != nil {
		return nil, err
	}

	if len(uris) == 0 {
		return nil, fmt.Errorf("no matches for %v", object)
	}

	sort.Sort(uris)

	for _, uri := range uris {
		rsp := Response{}

		if err := c.fetchAndUnmarshal(uri, &rsp); err != nil {
			continue
		}

		return &rsp, nil
	}

	return nil, fmt.Errorf("no data available for %v", object)
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
