package output

var entityTmpl = `{{range .ContactInfo.Persons}}
owner:       {{.}}
{{end}}ownerid:     (CPF/CNPJ)
responsible: {{.Entity.Responsible}}
{{range .ContactInfo.Addresses}}address:     {{.}}
{{end}}{{range .ContactInfo.Phones}}phone:     {{.}}
{{end}}
owner-c:     {{.Entity.Handle}}
created:     {{.CreatedAt}}
changed:     {{.UpdatedAt}}

{{range .ContactsInfos}}nic-hdl-br: {{.Handle}}
{{range .Persons}}person: {{.}}
{{end}}{{range .Emails}}e-mail: {{.}}
{{end}}{{range .Addresses}}address: {{.}}
{{end}}{{range .Phones}}phone: {{.}}
{{end}}created: {{.CreatedAt}}
changed: {{.UpdatedAt}}
{{end}}`
