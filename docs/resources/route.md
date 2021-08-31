---
page_title: "Mailgun: mailgun_route"
---

# mailgun\_route

Provides a Mailgun Route resource. This can be used to create and manage routes on Mailgun.

## Example Usage

```hcl
# Create a new Mailgun route
resource "mailgun_route" "default" {
  priority    = "0"
  description = "inbound"
  expression  = "match_recipient('.*@foo.example.com')"
  actions = [
    "forward('http://example.com/api/v1/foos/')",
    "stop()"
  ]
}
```

## Argument Reference

The following arguments are supported:
* `priority` - (Required) Smaller number indicates higher priority. Higher priority routes are handled first.
* `description` - (Required)
* `expression` - (Required) A filter expression like `match_recipient('.*@example.com')`
* `action` - (Required) Route action. This action is executed when the expression evaluates to True. Example: `forward("alice@example.com")` You can pass multiple `action` parameters.
* `region` - (Optional) The region where domain will be created. Default value is `us`.

## Import

Routes can be imported using `ROUTE_ID` and `region` via `import` command. Route ID can be found on Mailgun portal in section `Receiving/Routes`. Region has to be chosen from `eu` or `us` (when no selection `us` is applied). 

```hcl
terraform import mailgun_route.test eu:123456789
```
