package output

import (
	"testing"
	"time"

	"github.com/registrobr/rdap-client/protocol"
)

var TestEntityToTextOutput = `owner:       Joe User
ownerid:     (CPF/CNPJ)
responsible: 
address:      Av Naçoes Unidas 11541 7 andar Sao Paulo SP 04578-000 BR
country:     
phone:       tel:+55-11-5509-3506;ext=3506
owner-c:     XXXX
created:     2015-03-01T12:00:00Z
changed:     2015-03-10T14:00:00Z

nic-hdl-br: XXXX
person: Joe User
e-mail: joe.user@example.com
address:  Av Naçoes Unidas 11541 7 andar Sao Paulo SP 04578-000 BR
phone: tel:+55-11-5509-3506;ext=3506
created: 2015-03-01T12:00:00Z
changed: 2015-03-10T14:00:00Z
`

func TestEntityToText(t *testing.T) {
	entityResponse := protocol.Entity{
		ObjectClassName: "entity",
		Handle:          "XXXX",
		VCardArray: []interface{}{
			"vcard",
			[]interface{}{
				[]interface{}{"version", struct{}{}, "text", "4.0"},
				[]interface{}{"fn", struct{}{}, "text", "Joe User"},
				[]interface{}{"kind", struct{}{}, "text", "individual"},
				[]interface{}{"email", struct{ Type string }{Type: "work"}, "text", "joe.user@example.com"},
				[]interface{}{"lang", struct{ Pref string }{Pref: "1"}, "language-tag", "pt"},
				[]interface{}{"adr", struct{ Type string }{Type: "work"}, "text",
					[]string{
						"Av Naçoes Unidas", "11541", "7 andar", "Sao Paulo", "SP", "04578-000", "BR",
					},
				},
				[]interface{}{"tel", struct{ Type string }{Type: "work"}, "uri", "tel:+55-11-5509-3506;ext=3506"},
			},
		},
		Responsible: "",
		Entities: []protocol.Entity{
			{
				ObjectClassName: "entity",
				Handle:          "XXXX",
				VCardArray: []interface{}{
					"vcard",
					[]interface{}{
						[]interface{}{"version", struct{}{}, "text", "4.0"},
						[]interface{}{"fn", struct{}{}, "text", "Joe User"},
						[]interface{}{"kind", struct{}{}, "text", "individual"},
						[]interface{}{"email", struct{ Type string }{Type: "work"}, "text", "joe.user@example.com"},
						[]interface{}{"lang", struct{ Pref string }{Pref: "1"}, "language-tag", "pt"},
						[]interface{}{"adr", struct{ Type string }{Type: "work"}, "text",
							[]string{
								"Av Naçoes Unidas", "11541", "7 andar", "Sao Paulo", "SP", "04578-000", "BR",
							},
						},
						[]interface{}{"tel", struct{ Type string }{Type: "work"}, "uri", "tel:+55-11-5509-3506;ext=3506"},
					},
				},
				Events: []protocol.Event{
					protocol.Event{Action: protocol.EventActionRegistration, Actor: "", Date: time.Date(2015, 03, 01, 12, 00, 00, 00, time.UTC)},
					protocol.Event{Action: protocol.EventActionLastChanged, Actor: "", Date: time.Date(2015, 03, 10, 14, 00, 00, 00, time.UTC)},
				},
			},
		},
		Events: []protocol.Event{
			protocol.Event{Action: protocol.EventActionRegistration, Actor: "", Date: time.Date(2015, 03, 01, 12, 00, 00, 00, time.UTC)},
			protocol.Event{Action: protocol.EventActionLastChanged, Actor: "", Date: time.Date(2015, 03, 10, 14, 00, 00, 00, time.UTC)},
		},
	}

	entityOutput := Entity{Entity: &entityResponse}

	var w WriterMock
	if err := entityOutput.ToText(&w); err != nil {
		t.Fatal(err)
	}

	if string(w.Content) != TestEntityToTextOutput {
		t.Error("Wrong output")
		t.Log(string(w.Content))
		return
	}
}
