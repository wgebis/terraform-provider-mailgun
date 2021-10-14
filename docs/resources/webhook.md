---
page_title: "Mailgun: mailgun_webhook"
---

# mailgun\_webhook

Provides a Mailgun App resource. This can be used to
create and manage applications on Mailgun.

## Example Usage

```hcl
# Create a new Mailgun webhook
resource "mailgun_webhook" "default" {
  domain        = "test.example.com"
  region        = "us"
  kind          = "delivered"
  urls          = ["https://example.com"]
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain to add to Mailgun
* `region` - (Optional) The region where domain will be created. Default value is `us`.
* `kind` - (Required) The kind of webhook. Supported values (`clicked` `complained` `delivered` `opened` `permanent_fail`, `temporary_fail` `unsubscribed`)
* `urls` - (Required) The urls of webhook

## Attributes Reference

The following attributes are exported:

* `domain` - The name of the domain.
* `region` - The name of the region.
* `kind` - The kind of the webhook.
* `urls` - The urls of the webhook.
