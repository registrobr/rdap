package output

const asTmpl = `aut-num:     {{.AS.Handle}}
owner:       FUNDAÇÃO DE AMPARO À PESQUISA DO ESTADO SÃO PAULO
ownerid:     043.828.151/0001-45
responsible: Luis Fernandez Lopez
address:     Rua Pio XI, 1.500, 3.Andar/ANSP
address:     05468-150 - São Paulo - SP
country:     {{.AS.Country}}
phone:       (11) 38384000 [4072]
owner-c:     LFL180
routing-c:   JMA135
abuse-c:     JMA135
created:     {{.AS.CreateAt}}
changed:     {{.AS.UpdatedAt}}

inetnum:

{{range .AS.Entities}}
nic-hdl-br: {{.Handle}}
person: 
e-mail: 
address: 
address: 
phone: 
created: 
changed: 
{{end}}`
