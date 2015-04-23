package protocol

import (
	"math/big"
	"net"
	"strconv"
	"strings"
)

func (s ServiceRegistry) MatchAS(asn uint32) ([]string, error) {
	var (
		chosenURIs []string
		chosenSize uint32
	)

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

			if asn >= begin && asn <= end && begin-end > chosenSize {
				chosenSize = begin - end
				chosenURIs = service.URIs()
			}
		}
	}

	return chosenURIs, nil
}

func (s ServiceRegistry) MatchIPNetwork(network *net.IPNet) ([]string, error) {
	var (
		chosenURIs []string
		chosenSize = big.NewInt(0)
		begin      = big.NewInt(0).SetBytes(network.IP)
		mask       = big.NewInt(0).SetBytes(network.Mask)
		end        = big.NewInt(0).Xor(begin, mask)
	)

	for _, service := range s.Services {
		for _, entry := range service.Entries() {
			_, ipnet, err := net.ParseCIDR(entry)

			if err != nil {
				return nil, err
			}

			entryBegin := big.NewInt(0).SetBytes(ipnet.IP)
			mask := big.NewInt(0).SetBytes(ipnet.Mask)
			entryEnd := big.NewInt(0).Xor(entryBegin, mask)
			diff := big.NewInt(0).Sub(entryBegin, entryEnd)

			if entryBegin.Cmp(begin) >= 0 && entryEnd.Cmp(end) <= 0 && chosenSize.Cmp(diff) <= 0 {
				chosenURIs = service.URIs()
				chosenSize.Sub(entryEnd, entryBegin)
			}
		}
	}

	return chosenURIs, nil
}
