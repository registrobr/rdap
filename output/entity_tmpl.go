package output

var entityTmpl = `owner:       {{.ContactInfo.Person}}
ownerid:     (CPF/CNPJ)
responsible: 
address:     {{.ContactInfo.Address}}
country:     
phone:       {{.ContactInfo.Phone}}
owner-c:     {{.Entity.Handle}}
created:     {{.CreatedAt}}
changed:     {{.UpdatedAt}}
{{range .ContactsInfos}}
nic-hdl-br: {{.Handle}}
person: {{.Person}}
e-mail: {{.Email}}
address: {{.Address}}
phone: {{.Phone}}
created: {{.ContactCreatedAt}}
changed: {{.ContactUpdatedAt}}
{{end}}`
