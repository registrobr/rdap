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

func (a *AS) ToText(wr io.Writer) error {
	a.setDates()

	contacts := make(map[string]bool)
	a.ContactsInfos = make([]ContactInfo, 0, len(a.AS.Entities))
	for _, entity := range a.AS.Entities {
		if contacts[entity.Handle] == true {
			continue
		}
		contacts[entity.Handle] = true

		var c ContactInfo
		c.Handle = entity.Handle
		c.setContact(entity)
		a.ContactsInfos = append(a.ContactsInfos, c)
	}

	t, err := template.New("as template").Parse(asTmpl)
	if err != nil {
		return err
	}

	return t.Execute(wr, a)
}
