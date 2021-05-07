---
page_title: "Mailgun: mailgun_credential"
---

# mailgun\_credential

Provides a Mailgun Credential resource. This can be used to
create and manage credentials on Mailgun.

## Example Usage

```hcl
# Create a new Mailgun credential
resource "mailgun_credential" "default" {
  login    = "test"
  domain   = "example.com"
  password = "TestTest1!"
}
```

## Argument Reference

The following arguments are supported:

* `login` - (Required) New Mailgun login name
* `domain` - (Required) Existing Mailgun domain
* `password` - (Optional) Login password

## Import

Credentials can be imported using `login@domain` via `import` command.

```hcl
terraform import mailgun_credential.test example@domain.com
```
