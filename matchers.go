package rdap

import (
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/registrobr/rdap-client/Godeps/_workspace/src/github.com/miekg/dns/idn"
)

// MatchAS iterates through a list of services looking for the more
// specific range to which an AS number "asn" belongs.
//
// See http://tools.ietf.org/html/rfc7484#section-5.3
func (s ServiceRegistry) MatchAS(asn uint32) ([]string, error) {
	var (
		uris []string
		size = math.MaxUint32
	)

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			asRange := strings.Split(entry, "-")
			b, err := strconv.ParseUint(asRange[0], 10, 32)

			if err != nil {
				return nil, err
			}

			e, err := strconv.ParseUint(asRange[1], 10, 32)

			if err != nil {
				return nil, err
			}

			begin := uint32(b)
			end := uint32(e)

			if asn >= begin && asn <= end && end-begin < size {
				size = end - begin
				uris = service.URIs()
			}
		}
	}

	return uris, nil
}

// MatchIPNetwork iterates through a list of services looking for the more
// specific IP network to which the IP network "network" belongs.
//
// See http://tools.ietf.org/html/rfc7484#section-5.1
//     http://tools.ietf.org/html/rfc7484#section-5.2
func (s ServiceRegistry) MatchIPNetwork(network *net.IPNet) ([]string, error) {
	var (
		uris   []string
		size   = 0
		lastIP = make(net.IP, len(network.IP))
	)

	for i := 0; i < len(network.IP); i++ {
		lastIP[i] = network.IP[i] | ^network.Mask[i]
	}

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			_, ipnet, err := net.ParseCIDR(entry)

			if err != nil {
				return nil, err
			}

			mask, _ := ipnet.Mask.Size()

			if ipnet.Contains(network.IP) && ipnet.Contains(lastIP) && mask > size {
				uris = service.URIs()
				size = mask
			}
		}
	}

	return uris, nil
}

// MatchDomain iterates through a list of services looking for the label-wise
// longest match of the target domain name "fqdn".
//
// See http://tools.ietf.org/html/rfc7484#section-4
func (s ServiceRegistry) MatchDomain(fqdn string) ([]string, error) {
	var (
		uris      []string
		size      int
		fqdnParts = strings.Split(idn.ToPunycode(fqdn), ".")
	)

	for _, service := range s.Services {
	Entries:
		for _, entry := range service.Entries() {
			entryParts := strings.Split(entry, ".")

			if len(fqdnParts) < len(entryParts) {
				continue
			}

			fqdnExcerpt := fqdnParts[len(fqdnParts)-len(entryParts):]

			for i := len(entryParts) - 1; i >= 0; i-- {
				if fqdnExcerpt[i] != entryParts[i] {
					continue Entries
				}
			}

			if longest := len(entryParts); longest > size {
				uris = service.URIs()
				size = longest
			}
		}
	}

	return uris, nil
}
