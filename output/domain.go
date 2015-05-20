package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type Domain struct {
	Domain *protocol.DomainResponse

	CreatedAt string
	UpdatedAt string
	ExpiresAt string

	ContactsInfos []ContactInfo
}

func (d *Domain) setDates() {
	for _, e := range d.Domain.Events {
		date := e.Date.Format(time.RFC3339)

		switch e.Action {
		case protocol.EventActionRegistration:
			d.CreatedAt = date
		case protocol.EventActionLastChanged:
			d.UpdatedAt = date
		case protocol.EventActionExpiration:
			d.ExpiresAt = date
		}
	}
}

func (d *Domain) ToText(wr io.Writer) error {
	d.setDates()

	contacts := make(map[string]bool)
	d.ContactsInfos = make([]ContactInfo, 0, len(d.Domain.Entities))
	for _, entity := range d.Domain.Entities {
		if contacts[entity.Handle] == true {
			continue
		}
		contacts[entity.Handle] = true

		var c ContactInfo
		c.setContact(entity)
		d.ContactsInfos = append(d.ContactsInfos, c)
	}

	t, err := template.New("domain").Parse(domainTmpl)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(wr, "domain", d)
}
