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

### Write-only password (Terraform 1.11+)

`password` is stored in the Terraform state. With `password_wo` the password is
never written to the plan or the state, so it can come from an ephemeral value.
Terraform sends it on create and whenever `password_wo_version` changes, so bump
the version to rotate.

```hcl
ephemeral "random_password" "smtp" {
	length  = 32
	special = false
}

resource "mailgun_domain_credential" "foobar" {
	domain              = "toto.com"
	login               = "test"
	password_wo         = ephemeral.random_password.smtp.result
	password_wo_version = 1
	region              = "us"
}

# Hand the same password to whatever consumes it, for example Secret Manager.
resource "google_secret_manager_secret_version" "smtp" {
	secret                 = google_secret_manager_secret.smtp.id
	secret_data_wo         = ephemeral.random_password.smtp.result
	secret_data_wo_version = 1
}
```

## Argument Reference

The following arguments are supported:

* `domain` - (Required) The domain to add credential of Mailgun.
* `login` - (Required) The local-part of the email address to create.
* `password` - (Optional, Sensitive) Password for user authentication. Marked sensitive; not returned by the Mailgun API on read. Exactly one of `password` or `password_wo` must be set.
* `password_wo` - (Optional, Sensitive, Write-only) Password for user authentication that is never stored in the plan or state. Requires Terraform 1.11 or later and `password_wo_version`.
* `password_wo_version` - (Optional) Change this value to send `password_wo` to Mailgun again. Required with `password_wo`.
* `region` - (Optional) The region where domain credential will be created. Default value is `us`.

## Attributes Reference

The following attributes are exported:

* `domain` - The name of the domain.
* `email` - The email address.
* `password` - Password for user authentication.
* `region` - The name of the region.

## Import

Domain credential can be imported using `region:email` via `import` command. Region has to be chosen from `eu` or `us` (when no selection `us` is applied). 
Password is always exported to `null`. With `password_wo`, the first plan after
an import shows an update (because `password_wo_version` is not yet in state),
which sets the configured password on the imported credential.

```hcl
terraform import mailgun_domain_credential.test us:test@domain.com
```
