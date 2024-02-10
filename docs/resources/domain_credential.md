---
page_title: "Mailgun: mailgun_domain"
---

# mailgun\_domain_credential

Provides a Mailgun domain credential resource. This can be used to create and manage credential in domain of Mailgun.

~> **Note:** Please note that starting of v0.6.1 due to using new Mailgun Client API (v4), there is no possibility to retrieve previously created secrets via API. In order get it worked, it's recommended to mark `password` as ignored under `lifecycle` block. See below.

## Example Usage

```hcl
# Create a new Mailgun credential
resource "mailgun_domain_credential" "foobar" {
	domain = "toto.com"
	login = "test"
	password = "supersecretpassword1234"
	region = "us"
	
	lifecycle {
	    ignore_changes = [ "password" ]
	}
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain to add credential of Mailgun.
* `login` - (Required) The local-part of the email address to create.
* `password` - (Required) Password for user authentication.
* `region` - (Optional) The region where domain will be created. Default value is `us`.

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
