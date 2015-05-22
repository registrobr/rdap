package output

const asTmpl = `aut-num:     {{.AS.Handle}}
owner:       (name)
ownerid:     (CPF/CNPJ)
responsible: (name)
address:     
address:     
country:     {{.AS.Country}}
phone:       
owner-c:     (handle)
routing-c:   (handle)
abuse-c:     (handle)
created:     {{.CreatedAt}}
changed:     {{.UpdatedAt}}

inetnum:     (ip networks)

{{range .ContactsInfos}}nic-hdl-br: {{.Handle}}
{{range .Persons}}person: {{.}}
{{end}}{{range .Emails}}e-mail: {{.}}
{{end}}{{range .Addresses}}address: {{.}}
{{end}}{{range .Phones}}phone: {{.}}
{{end}}created: {{.CreatedAt}}
changed: {{.UpdatedAt}}
{{end}}`
