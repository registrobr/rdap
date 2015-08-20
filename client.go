package rdap

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/registrobr/rdap/Godeps/_workspace/src/github.com/miekg/dns/idn"
	"github.com/registrobr/rdap/protocol"
)

// Client is responsible for building, sending the request and parsing the
// result. It can set the URIs attribute if you want to query RDAP servers
// directly without using bootstrap
type Client struct {
	// Transport is the network layer that you can fill with a direct query to
	// the RDAP servers or with an extra layer of RDAP bootstrap strategy
	Transport Fetcher

	// URIs store the addresses of the RDAP servers that you want to query
	// directly. Remember that if you use a bootstrap transport layer this
	// information might not be used
	URIs []string
}

// NewClient is an easy way to create a client with bootstrap support or not,
// depending if you inform direct RDAP addresses. Optionally you can define
// an X-Fowarded-For HTTP header to work as a proxy client
func NewClient(URIs []string, xForwardedFor string) *Client {
	client := Client{
		URIs: URIs,
	}

	var httpClient http.Client

	if len(URIs) == 0 {
		client.Transport = NewBootstrapFetcher(&httpClient, xForwardedFor, IANABootstrap, nil)
	} else {
		client.Transport = NewDefaultFetcher(&httpClient, xForwardedFor)
	}

	return &client
}

// Domain will query each RDAP server to retrieve the desired information and
// will parse and store the response into a protocol Domain object. If
// something goes wrong an error will be returned, and if nothing is found
// the error ErrNotFound will be returned
func (c *Client) Domain(fqdn string) (*protocol.Domain, error) {
	fqdn = idn.ToPunycode(strings.ToLower(fqdn))

	resp, err := c.Transport.Fetch(c.URIs, QueryTypeDomain, fqdn)
	if err != nil {
		return nil, err
	}

	domain := &protocol.Domain{}
	if err = json.NewDecoder(resp.Body).Decode(domain); err != nil {
		return nil, err
	}

	return domain, nil
}

// ASN will query each RDAP server to retrieve the desired information and
// will parse and store the response into a protocol AS object. If
// something goes wrong an error will be returned, and if nothing is found
// the error ErrNotFound will be returned
func (c *Client) ASN(asn uint32) (*protocol.AS, error) {
	resp, err := c.Transport.Fetch(c.URIs, QueryTypeAutnum, strconv.FormatUint(uint64(asn), 10))
	if err != nil {
		return nil, err
	}

	as := &protocol.AS{}
	if err = json.NewDecoder(resp.Body).Decode(as); err != nil {
		return nil, err
	}

	return as, nil
}

// Entity will query each RDAP server to retrieve the desired information and
// will parse and store the response into a protocol Entity object. If
// something goes wrong an error will be returned, and if nothing is found
// the error ErrNotFound will be returned
func (c *Client) Entity(identifier string) (*protocol.Entity, error) {
	resp, err := c.Transport.Fetch(c.URIs, QueryTypeEntity, identifier)
	if err != nil {
		return nil, err
	}

	entity := &protocol.Entity{}
	if err = json.NewDecoder(resp.Body).Decode(entity); err != nil {
		return nil, err
	}

	return entity, nil
}

// IPNetwork will query each RDAP server to retrieve the desired information and
// will parse and store the response into a protocol IPNetwork object. If
// something goes wrong an error will be returned, and if nothing is found
// the error ErrNotFound will be returned
func (c *Client) IPNetwork(ipnet *net.IPNet) (*protocol.IPNetwork, error) {
	if ipnet == nil {
		return nil, fmt.Errorf("undefined IP network")
	}

	resp, err := c.Transport.Fetch(c.URIs, QueryTypeIP, ipnet.String())
	if err != nil {
		return nil, err
	}

	ipNetwork := &protocol.IPNetwork{}
	if err = json.NewDecoder(resp.Body).Decode(ipNetwork); err != nil {
		return nil, err
	}

	return ipNetwork, nil
}

// IP will query each RDAP server to retrieve the desired information and
// will parse and store the response into a protocol IP object. If
// something goes wrong an error will be returned, and if nothing is found
// the error ErrNotFound will be returned
func (c *Client) IP(ip net.IP) (*protocol.IPNetwork, error) {
	if ip == nil {
		return nil, fmt.Errorf("undefined IP")
	}

	resp, err := c.Transport.Fetch(c.URIs, QueryTypeIP, ip.String())
	if err != nil {
		return nil, err
	}

	ipNetwork := &protocol.IPNetwork{}
	if err = json.NewDecoder(resp.Body).Decode(ipNetwork); err != nil {
		return nil, err
	}

	return ipNetwork, nil
}
