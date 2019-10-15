---
layout: "mailgun"
page_title: "Mailgun: mailgun_route"
sidebar_current: "docs-mailgun-resource-route"
description: |-
  Provides a Mailgun App resource. This can be used to create and manage applications on Mailgun.
---

# mailgun\_route

Provides a Mailgun Route resource. This can be used to create and manage routes on Mailgun.

## Example Usage

```hcl
# Create a new Mailgun route
resource "mailgun_route" "default" {
    priority = "0"
    description = "inbound"
    expression = "match_recipient('.*@foo.example.com')"
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
* `expression` - (Required) A filter expression like `match_recipient('.*@gmail.com')`
* `action` - (Required) Route action. This action is executed when the expression evaluates to True. Example: `forward("alice@example.com")` You can pass multiple `action` parameters.

## Import

Routes can be imported using `ROUTE_ID` via `import` command. Route ID can be found on Mailgun portal in section `Receiving/Routes`.

```hcl
terraform import mailgun_route ROUTE_ID
```