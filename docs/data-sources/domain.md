---
page_title: "Mailgun: mailgun_domain"
---

# mailgun\_domain

`mailgun_domain` provides details about a Mailgun domain.

## Example Usage

```hcl
data "mailgun_domain" "domain" {
  name = "test.example.com"
}

resource "aws_route53_record" "mailgun-mx" {
  zone_id = "${var.zone_id}"
  name    = "${data.mailgun.domain.name}"
  type    = "MX"
  ttl     = 3600
  records = [
    "${data.mailgun_domain.domain.receiving_records.0.priority} ${data.mailgun_domain.domain.receiving_records.0.value}.",
    "${data.mailgun_domain.domain.receiving_records.1.priority} ${data.mailgun_domain.domain.receiving_records.1.value}.",
  ]
}
```

## Argument Reference

* `name` - (Required) The name of the domain.
* `region` - (Optional) The region where domain will be created. Default value is `us`.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the domain.
* `smtp_login` - The login email for the SMTP server.
* `smtp_password` - The password to the SMTP server.
* `wildcard` - Whether or not the domain will accept email for sub-domains.
* `spam_action` - The spam filtering setting.
* `receiving_records` - A list of DNS records for receiving validation.
    * `priority` - The priority of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.
* `sending_records` - A list of DNS records for sending validation.
    * `name` - The name of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.