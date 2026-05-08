---
page_title: "Mailgun: mailgun_domain"
---

# mailgun\_domain_credential

Provides a Mailgun domain credential resource. This can be used to create and manage credential in domain of Mailgun.

~> **Note:** The Mailgun API does not return previously created credential passwords on read. The provider therefore does not refresh `password` from the API and only sends it to Mailgun on create or when the value in configuration changes. If your configured `password` value drifts (for example you rotate it out-of-band), use a `lifecycle.ignore_changes = [ password ]` block to avoid spurious updates.

## Example Usage

```hcl
# Create a new Mailgun credential
resource "mailgun_domain_credential" "foobar" {
	domain   = "toto.com"
	login    = "test"
	password = "supersecretpassword1234"
	region   = "us"
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain to add credential of Mailgun.
* `login` - (Required) The local-part of the email address to create.
* `password` - (Required, Sensitive) Password for user authentication. Marked sensitive; not returned by the Mailgun API on read.
* `region` - (Optional) The region where domain credential will be created. Default value is `us`.

## Attributes Reference

The following attributes are exported:

* `domain` - The name of the domain.
* `email` - The email address.
* `password` - Password for user authentication.
* `region` - The name of the region.

## Import

Domain credential can be imported using `region:email` via `import` command. Region has to be chosen from `eu` or `us` (when no selection `us` is applied). 
Password is always exported to `null`.

```hcl
terraform import mailgun_domain_credential.test us:test@domain.com
```
