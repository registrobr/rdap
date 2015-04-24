package protocol

import (
	"math"
	"math/big"
	"net"
	"strconv"
	"strings"
)

func (s ServiceRegistry) MatchAS(asn uint32) ([]string, error) {
	var (
		uris []string
		size uint32 = math.MaxUint32
	)

	if len(s.Services) > 0 {
		uris = s.Services[0].URIs()
	}

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			asRange := strings.Split(entry, "-")
			b, err := strconv.Atoi(asRange[0])

			if err != nil {
				return nil, err
			}

			e, err := strconv.Atoi(asRange[1])

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

func (s ServiceRegistry) MatchIPNetwork(network *net.IPNet) ([]string, error) {
	var (
		uris   []string
		size   = big.NewInt(0)
		lastIP = lastAddress(network)
	)

	ipSize := net.IPv6len

	if network.IP.To4() != nil {
		ipSize = net.IPv4len
	}

	size.SetBytes(net.CIDRMask(ipSize*8, ipSize*8))

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			_, ipnet, err := net.ParseCIDR(entry)

			if err != nil {
				return nil, err
			}

			entryBegin := big.NewInt(0).SetBytes(ipnet.IP)
			entryEnd := big.NewInt(0).SetBytes(lastAddress(ipnet))
			diff := big.NewInt(0).Sub(entryEnd, entryBegin)

			if ipnet.Contains(network.IP) && ipnet.Contains(lastIP) && diff.Cmp(size) == -1 {
				uris = service.URIs()
				*size = *diff
			}
		}
	}

	return uris, nil
}

func (s ServiceRegistry) MatchDomain(fqdn string) ([]string, error) {
	var (
		uris []string
		size int
	)

	if len(s.Services) > 0 {
		uris = s.Services[0].URIs()
	}

	fqdnParts := strings.Split(fqdn, ".")

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			index := 0
			entryParts := strings.Split(entry, ".")

			if len(fqdnParts) < len(entryParts) {
				fqdnParts, entryParts = entryParts, fqdnParts
			}

			for i := len(entryParts) - 1; i >= 0; i-- {
				if entryParts[i] == fqdnParts[i] {
					index++
				}
			}

			if index > size {
				uris = service.URIs()
				size = index
			}
		}
	}

	return uris, nil
}

func lastAddress(n *net.IPNet) net.IP {
	b := make(net.IP, len(n.IP))
	for i := 0; i <= len(n.IP)-1; i++ {
		b[i] = n.IP[i] | ^n.Mask[i]
	}
	return b
}
