package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap/protocol"
)

type Domain struct {
	Domain *protocol.DomainResponse

	CreatedAt string
	UpdatedAt string
	ExpiresAt string

	Handles       map[string]string
	DS            []ds
	ContactsInfos []ContactInfo
}

type ds struct {
	protocol.DS
	CreatedAt string
}

type textOut struct {
	Domain        string
	LocalTime     string
	Owner         string
	OwnerId       string
	Responsible   string
	Address1      string
	Address2      string
	Country       string
	Phone         string
	AdminHandle   string
	OwnerHandle   string
	TecHandle     string
	BillingHandle string
	Nameservers   []struct {
		Host       string
		CheckedAt  string
		LastStatus string
		LastAAAt   string
	}
	DSSet []struct {
		Record     string // keytag + algorithm + digest
		CheckedAt  string
		LastStatus string
		LastOKAt   string
	}
	CreatedAt string
	ChangedAt string
	Status    string
	Contacts  []struct {
		Handle    string
		Name      string
		Email     string
		Address1  string
		Address2  string
		Phone     string
		CreatedAt string
		ChangedAt string
	}
}

func (d *Domain) AddContact(c ContactInfo) {
	d.ContactsInfos = append(d.ContactsInfos, c)
}

func (d *Domain) setDates() {
	for _, e := range d.Domain.Events {
		date := e.Date.Format("20060102")

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

func (d *Domain) setDS() {
	d.DS = make([]ds, len(d.Domain.SecureDNS.DSData))

	for i, dsdatum := range d.Domain.SecureDNS.DSData {
		myds := ds{DS: dsdatum}

		for _, e := range dsdatum.Events {
			if e.Action == protocol.EventActionRegistration {
				myds.CreatedAt = e.Date.Format("20060102")
			}
		}

		d.DS[i] = myds
	}
}

func (d *Domain) ToText(wr io.Writer) error {
	txtOut := textOut{}
	txtOut.Domain = d.Domain.UnicodeName
	txtOut.LocalTime = time.Now().Format("2006-01-02 15:04:05 (MST -0700)")

	if contactInfo, ok := getOwner(d); ok {
		if len(contactInfo.Persons) > 0 {
			txtOut.Owner = contactInfo.Persons[0]
		}
	}

	// d.setDates()
	// d.setDS()

	// AddContacts(d, d.Domain.Entities)

	// for _, entity := range d.Domain.Entities {
	// 	AddContacts(d, entity.Entities)
	// }

	t, err := template.New("domain").Funcs(domainFuncMap).Parse(domainTmpl)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(wr, "domain", txtOut)
}

func getOwner(d *Domain) (contactInfo ContactInfo, ok bool) {
	for _, e := range d.Domain.Entities {
		for _, r := range e.Roles {
			if r == "registrant" {
				contactInfo.setContact(r)
				return contactInfo, true
			}
		}
	}

	return
}
