package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type Entity struct {
	Entity *protocol.Entity

	CreatedAt string
	UpdatedAt string

	ContactInfo   ContactInfo
	ContactsInfos []ContactInfo
}

func (e *Entity) setDates() {
	for _, event := range e.Entity.Events {
		date := event.Date.Format(time.RFC3339)

		switch event.Action {
		case protocol.EventActionRegistration:
			e.CreatedAt = date
		case protocol.EventActionLastChanged:
			e.UpdatedAt = date
		}
	}
}

func (e *Entity) ToText(wr io.Writer) error {
	e.setDates()
	e.ContactInfo.setContact(*e.Entity)

	contacts := make(map[string]bool)
	e.ContactsInfos = make([]ContactInfo, 0, len(e.Entity.Entities))
	for _, entity := range e.Entity.Entities {
		if contacts[entity.Handle] == true {
			continue
		}
		contacts[entity.Handle] = true

		var c ContactInfo
		c.Handle = entity.Handle
		c.setContact(entity)
		e.ContactsInfos = append(e.ContactsInfos, c)
	}

	t, err := template.New("entity template").Parse(entityTmpl)
	if err != nil {
		return err
	}

	return t.Execute(wr, e)
}
