package output

import (
	"io"
	"text/template"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

type domain struct {
	Domain       *protocol.DomainResponse
	Registration string
	Updated      string
	Expiration   string
}

func (d *domain) transform() {
	for _, event := range d.Domain.Events {
		date := event.Date.Format(time.RFC3339)

		switch event.Action {
		case protocol.EventActionRegistration:
			d.Registration = date
		case protocol.EventActionLastChanged:
			d.Updated = date
		case protocol.EventActionExpiration:
			d.Expiration = date
		}
	}
}

func PrintDomain(r *protocol.DomainResponse, wr io.Writer) error {
	d := domain{Domain: r}
	d.transform()

	t, err := template.New("domain").Parse(domainTmpl)

	if err != nil {
		return err
	}

	return t.ExecuteTemplate(wr, "domain", d)
}
