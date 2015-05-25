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

func (d *Domain) AddContact(c ContactInfo) {
	d.ContactsInfos = append(d.ContactsInfos, c)
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
	AddContacts(d, d.Domain.Entities)

	t, err := template.New("domain").Parse(domainTmpl)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(wr, "domain", d)
}
