package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type IPNetwork struct {
	IPNetwork *protocol.IPNetwork

	CreatedAt string
	UpdatedAt string

	ContactsInfos []ContactInfo
}

func (i *IPNetwork) setDates() {
	for _, e := range i.IPNetwork.Events {
		date := e.Date.Format(time.RFC3339)

		switch e.Action {
		case protocol.EventActionRegistration:
			i.CreatedAt = date
		case protocol.EventActionLastChanged:
			i.UpdatedAt = date
		}
	}
}

func (i *IPNetwork) ToText(wr io.Writer) error {
	i.setDates()

	contacts := make(map[string]bool)
	i.ContactsInfos = make([]ContactInfo, 0, len(i.IPNetwork.Entities))
	for _, entity := range i.IPNetwork.Entities {
		if contacts[entity.Handle] == true {
			continue
		}
		contacts[entity.Handle] = true

		var c ContactInfo
		c.setContact(entity)
		i.ContactsInfos = append(i.ContactsInfos, c)
	}

	t, err := template.New("ipnetwork template").Parse(ipnetTmpl)
	if err != nil {
		return err
	}

	return t.Execute(wr, i)
}
