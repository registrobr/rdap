package output

import (
	"strings"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type ContactInfo struct {
	Handle    string
	Persons   []string
	Emails    []string
	Addresses []string
	Phones    []string
	CreatedAt string
	UpdatedAt string
}

func (c *ContactInfo) setContact(entity protocol.Entity) {
	for _, vCardValues := range entity.VCardArray {
		vCardValue, ok := vCardValues.([]interface{})
		if !ok {
			continue
		}

		for _, value := range vCardValue {
			v, ok := value.([]interface{})
			if !ok {
				continue
			}

			switch v[0] {
			case "fn":
				c.Persons = append(c.Persons, v[3].(string))
			case "email":
				c.Emails = append(c.Emails, v[3].(string))
			case "adr":
				address := make([]string, 0)

				for _, v := range v[3].([]string) {
					if len(v) > 0 {
						address = append(address, v)
					}
				}

				c.Addresses = append(c.Addresses, strings.Join(address, ", "))
			case "tel":
				c.Phones = append(c.Phones, v[3].(string))
			}
		}
	}

	for _, event := range entity.Events {
		date := event.Date.Format(time.RFC3339)

		switch event.Action {
		case protocol.EventActionRegistration:
			c.CreatedAt = date
		case protocol.EventActionLastChanged:
			c.UpdatedAt = date
		}
	}
}
