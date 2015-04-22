package protocol

import (
	"strconv"
	"strings"
)

func (s ServiceRegistry) MatchAS(asn uint32) ([]string, error) {
	var chosenEntry []string
	var chosenSize uint32

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
				chosenEntry = service.URIs()
			}
		}
	}

	return chosenEntry, nil
}
