package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type AS struct {
	AS *protocol.ASResponse

	CreatedAt string
	UpdatedAt string

	ContactsInfos []ContactInfo
}

type ContactInfo struct {
	Handle           string
	Person           string
	Email            string
	Address          string
	Phone            string
	ContactCreatedAt string
	ContactUpdatedAt string
}

func (a *AS) setDates() {
	for _, e := range a.AS.Events {
		date := e.Date.Format(time.RFC3339)

		switch e.Action {
		case protocol.EventActionRegistration:
			a.CreatedAt = date
		case protocol.EventActionLastChanged:
			a.UpdatedAt = date
		}
	}
}

func (a *AS) setContact() {
	a.ContactsInfos = make([]ContactInfo, 0, len(a.AS.Entities))
	for _, entity := range a.AS.Entities {
		var contactInfo ContactInfo

		contactInfo.Handle = entity.Handle
		for _, vCardValues := range entity.VCardArray {
			if _, ok := vCardValues.([]interface{}); !ok {
				continue
			}

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
					contactInfo.Person = v[3].(string)
				case "email":
					contactInfo.Email = v[3].(string)
				case "adr":
					for _, v := range v[3].([]string) {
						contactInfo.Address += " " + v
					}
				case "tel":
					contactInfo.Phone = v[3].(string)
				}
			}
		}

		for _, event := range entity.Events {
			date := event.Date.Format(time.RFC3339)

			switch event.Action {
			case protocol.EventActionRegistration:
				contactInfo.ContactCreatedAt = date
			case protocol.EventActionLastChanged:
				contactInfo.ContactUpdatedAt = date
			}
		}

		a.ContactsInfos = append(a.ContactsInfos, contactInfo)
	}
}

func (a *AS) ToText(wr io.Writer) error {
	a.setDates()
	a.setContact()

	t, err := template.New("as template").Parse(asTmpl)
	if err != nil {
		return err
	}

	return t.Execute(wr, a)
}
