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

Here's an example using the [Cloudflare provider](https://registry.terraform.io/providers/cloudflare/cloudflare/latest). Bear in mind that the solution below requires the Cloudflare provider to be included in your project. Also, the Mailgun provider isn't associated with Cloudflare, and other Terraform providers that can control DNS may require a slightly different implementation.

For detailed setup instructions, see Mailgun's [Domain Verification Setup Guide](https://help.mailgun.com/hc/en-us/articles/32884702360603-Domain-Verification-Setup-Guide) or the [Cloudflare DNS Setup Guide](https://help.mailgun.com/hc/en-us/articles/15585722150299-Cloudflare-DNS-Setup-Guide).

```hcl
# Use receiving/sending set attributes to create DNS entries
# TTL is set to 300 seconds (5 minutes) for faster updates as recommended by Mailgun
# You can adjust the TTL to your desired value
resource "cloudflare_dns_record" "default_receiving" {
  for_each = {
    for record in mailgun_domain.default.receiving_records_set : record.id => {
      type     = record.record_type
      value    = record.value
      priority = record.priority
    }
  }

  zone_id  = var.zone_id
  name     = var.domain

  type     = each.value.type
  content  = each.value.value
  priority = each.value.priority
  ttl      = 300
}

resource "cloudflare_dns_record" "default_sending" {
  for_each = {
    for record in mailgun_domain.default.sending_records_set : record.id => {
      name  = record.name
      type  = record.record_type
      value = record.value
    }
  }

  zone_id = var.zone_id

  name    = each.value.name
  type    = each.value.type
  content = each.value.value
  ttl     = 300
}

# Create MX records pointing to Mailgun
# Use "@" for name if using the root domain, or the subdomain name if using a subdomain
resource "cloudflare_dns_record" "mx_records" {
  for_each = toset(["mxa.mailgun.org", "mxb.mailgun.org"])

  zone_id = var.zone_id

  name     = "@"  # Use "@" for root domain or subdomain name (e.g., "mail") for subdomains
  type     = "MX"
  content  = each.value
  priority = 10
  ttl      = 300
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
* `click_tracking` - (Optional) (Enum: `yes` or `no`) The click tracking settings for the domain. Default: `no`
* `web_scheme` - (Optional) (`http` or `https`) The tracking web scheme. Default: `http`
* `use_automatic_sender_security` - (Optional) If true Mailgun manages DKIM key generation and DNS record configuration automatically. Default: `false`

## Attributes Reference

The following attributes are exported:

* `name` - The name of the domain.
* `region` - The name of the region.
* `smtp_login` - The login email for the SMTP server.
* `smtp_password` - The password to the SMTP server.
* `wildcard` - Whether or not the domain will accept email for sub-domains.
* `spam_action` - The spam filtering setting.
* `open_tracking` - The open tracking setting.
* `click_tracking` - The click tracking setting.
* `web_scheme` - The tracking web scheme.
* `use_automatic_sender_security` - Whether or not automatic sender sender security is enabled.
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
