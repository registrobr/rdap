package output

const domainTmpl = `Domain Name: {{.Domain.LDHName}}
Domain ID: {{.Domain.LDHName}}
Creation Date: {{.Registration}}
Updated Date: {{.Updated}}
Registry Expiry Date: {{.Expiration}}
Trademark Name:
Trademark Date:
Trademark Country:
Trademark Number:
Sponsoring Registrar:
Sponsoring Registrar IANA ID:
WHOIS Server:
Referral URL:
{{range .Domain.Status}}Domain Status: {{.}}
{{end}}Registrant ID:
Registrant Name:
Registrant Organization:
Registrant Street:
Registrant City:
Registrant State/Province:
Registrant Postal Code:
Registrant Country:
Registrant Phone:
Registrant Phone Ext:
Registrant Fax:
Registrant Fax Ext:
Registrant Email:
Admin ID:
Admin Name:
Admin Organization:
Admin Street:
Admin City:
Admin State/Province:
Admin Postal Code:
Admin Country:
Admin Phone:
Admin Phone Ext:
Admin Fax:
Admin Fax Ext:
Admin Email:
Billing ID:
Billing Name:
Billing Organization:
Billing Street:
Billing City:
Billing State/Province:
Billing Postal Code:
Billing Country:
Billing Phone:
Billing Phone Ext:
Billing Fax:
Billing Fax Ext:
Billing Email:
Tech ID:
Tech Name:
Tech Organization:
Tech Street:
Tech City:
Tech State/Province:
Tech Postal Code:
Tech Country:
Tech Phone:
Tech Phone Ext:
Tech Fax:
Tech Fax Ext:
Tech Email:
{{range .Domain.Nameservers}}Nameserver: {{.LDHName}}
{{end}}DNSSEC: 
`
