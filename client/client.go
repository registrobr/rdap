package client

import (
	"encoding/json"
	"errors"
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
	dns    kind = "domain"
	asn    kind = "autnum"
	ip     kind = "ip"
	entity kind = "entity"
)

var SERVER_RETURNED_NOT_200 error = errors.New("The HTTP status code returned from server is not 200 OK!")

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

	err := c.query(dns, fqdn, r)
	if err != nil {
		if err == SERVER_RETURNED_NOT_200 {
			// TODO - handle http status code returned from server properly
			return nil, nil
		}
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

func (c *Client) Entity(identifier string) (*protocol.Entity, error) {
	r := &protocol.Entity{}

	if err := c.query(entity, identifier, r); err != nil {
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

func (c *Client) query(kind kind, identifier interface{}, object interface{}) (err error) {
	for _, uri := range c.uris {
		uri := fmt.Sprintf("%s/%s/%v", uri, kind, identifier)

		var body io.ReadCloser
		body, err = c.fetch(uri)

		if err != nil {
			continue
		}

		defer body.Close()
		if err = json.NewDecoder(body).Decode(&object); err != nil {
			continue
		}

		return err
	}

	if err != nil {
		return err
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

	if resp.StatusCode != http.StatusOK {
		// TODO - handle the error properly
		return resp.Body, SERVER_RETURNED_NOT_200
	}

	return resp.Body, nil
}
