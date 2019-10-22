---
layout: "mailgun"
page_title: "Mailgun: mailgun_domain"
sidebar_current: "docs-mailgun-resource-domain"
description: |-
  Provides a Mailgun App resource. This can be used to create and manage applications on Mailgun.
---

# mailgun\_domain

Provides a Mailgun App resource. This can be used to
create and manage applications on Mailgun.

After DNS records are set, domain verification should be triggered manually using [PUT /domains/\<domain\>/verify](https://documentation.mailgun.com/en/latest/api-domains.html#domains)

## Example Usage

```hcl
# Create a new Mailgun domain
resource "mailgun_domain" "default" {
  name          = "test.example.com"
  region        = "us"
  spam_action   = "disabled"
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain to add to Mailgun
* `region` - (Optional) The region where domain will be created. Default value is `us`.
* `spam_action` - (Optional) `disabled` or `tag` Disable, no spam
    filtering will occur for inbound messages. Tag, messages
    will be tagged with a spam header.
* `wildcard` - (Optional) Boolean that determines whether
    the domain will accept email for sub-domains.

## Attributes Reference

The following attributes are exported:

* `name` - The name of the domain.
* `region` - The name of the region.
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

## Import

Domains can be imported using `region:domain_name` via `import` command. Region has to be chosen from `eu` or `us` (when no selection `us` is applied). 

```hcl
terraform import mailgun_domain.test us:example.domain.com
```
