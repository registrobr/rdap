package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/miekg/dns/idn"
	"github.com/registrobr/rdap-client/protocol"
)

type kind string

const (
	dns kind = "domain"
	asn kind = "autnum"
	ip  kind = "ip"
)

type Client struct {
	httpClient *http.Client
	uris       []string
}

func NewClient(uris []string, httpClient *http.Client) *Client {
	return &Client{
		uris:       uris,
		httpClient: httpClient,
	}
}

func (c *Client) Domain(fqdn string) (*protocol.DomainResponse, error) {
	r := &protocol.DomainResponse{}
	fqdn = strings.ToLower(idn.ToPunycode(fqdn))

	if err := c.query(dns, fqdn, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) ASN(as uint64) (*protocol.ASResponse, error) {
	r := &protocol.ASResponse{}

	if err := c.query(asn, as, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) IPNetwork(ipnet *net.IPNet) (*protocol.IPNetwork, error) {
	r := &protocol.IPNetwork{}

	if err := c.query(ip, ipnet, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) IP(netIP net.IP) (*protocol.IPNetwork, error) {
	r := &protocol.IPNetwork{}

	if err := c.query(ip, netIP, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Client) query(kind kind, identifier interface{}, object interface{}) error {
	for _, uri := range c.uris {
		uri := fmt.Sprintf("%s/%s/%v", uri, kind, identifier)
		body, err := c.fetch(uri)

		if err != nil {
			continue
		}

		defer body.Close()

		if err := json.NewDecoder(body).Decode(&object); err != nil {
			continue
		}

		return nil
	}

	return fmt.Errorf("no data available for %v", identifier)
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
