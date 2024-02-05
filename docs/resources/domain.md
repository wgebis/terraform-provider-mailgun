---
page_title: "Mailgun: mailgun_domain"
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
  smtp_password   = "supersecretpassword1234"
  dkim_key_size   = 1024
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The domain to add to Mailgun
* `region` - (Optional) The region where domain will be created. Default value is `us`.
* `smtp_password` - (Optional) Password for SMTP authentication
* `spam_action` - (Optional) `disabled` or `tag` Disable, no spam
    filtering will occur for inbound messages. Tag, messages
    will be tagged with a spam header. Default value is `disabled`.
* `wildcard` - (Optional) Boolean that determines whether
    the domain will accept email for sub-domains.
* `dkim_key_size` - (Optional) The length of your domain’s generated DKIM key. Default value is `1024`.
* `dkim_selector` - (Optional) The name of your DKIM selector if you want to specify it whereas MailGun will make it's own choice.
* `force_dkim_authority` - (Optional) If set to true, the domain will be the DKIM authority for itself even if the root domain is registered on the same mailgun account. If set to false, the domain will have the same DKIM authority as the root domain registered on the same mailgun account. The default is `false`.
* `open_tracking` - (Optional) (Enum: `yes` or `no`) The open tracking settings for the domain. Default: `no`
* `web_scheme` - (Optional) (`http` or `https`) The tracking web scheme. Default: `http`

## Attributes Reference

The following attributes are exported:

* `name` - The name of the domain.
* `region` - The name of the region.
* `smtp_login` - The login email for the SMTP server.
* `smtp_password` - The password to the SMTP server.
* `wildcard` - Whether or not the domain will accept email for sub-domains.
* `spam_action` - The spam filtering setting.
* `web_scheme` - The tracking web scheme.
* `receiving_records` - A list of DNS records for receiving validation.  **Deprecated** Use `receiving_records_set` instead.
    * `priority` - The priority of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.
* `receiving_records_set` - A set of DNS records for receiving validation.
    * `priority` - The priority of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.
* `sending_records` - A list of DNS records for sending validation. **Deprecated** Use `sending_records_set` instead.
    * `name` - The name of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.
* `sending_records_set` - A set of DNS records for sending validation.
    * `name` - The name of the record.
    * `record_type` - The record type.
    * `valid` - `"valid"` if the record is valid.
    * `value` - The value of the record.

## Import

Domains can be imported using `region:domain_name` via `import` command. Region has to be chosen from `eu` or `us` (when no selection `us` is applied). 

```hcl
terraform import mailgun_domain.test us:example.domain.com
```
