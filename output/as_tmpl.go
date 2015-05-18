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

{{range .ContactsInfos}}
nic-hdl-br: {{.Handle}}
person: {{.Person}}
e-mail: {{.Email}}
address: {{.Address}}
phone: {{.Phone}}
created: {{.ContactCreatedAt}}
changed: {{.ContactUpdatedAt}}
{{end}}`
