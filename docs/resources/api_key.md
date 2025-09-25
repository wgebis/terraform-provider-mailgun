---
page_title: "Mailgun: mailgun_domain"
---

# mailgun\_api\_key

Provides a Mailgun API key resource. This can be used to  create and manage API keys on Mailgun.

~> **Note:** Please note that due to the limitations of the Terraform SDK v2 this provider uses, the removal of API keys
which have their expiration set cannot be handled properly after expiration. In order to remove such expired keys, it is
recommended to use `terraform state rm`.

## Example Usage

```hcl
# Create a new Mailgun API key
resource "mailgun_api_key" "some_key" {
  role = "basic"
  kind = "user"
  description = "Some key"
}
```

## Argument Reference

The following arguments are supported:

* `role` - (Required) (Enum: `admin`, `basic`, `sending`, `support`, or `developer`) Key role.
* `description` - (Optional) Key description.
* `kind` - (Optional) (Enum:`domain`, `user`, or `web`). API key type. Default: `user`.
* `expiration` - (Optional) Key lifetime in seconds, must be greater than 0 if set.
* `email` - (Optional) API key user's email address; should be provided for all keys of `web` kind.
* `domain_name` - (Optional) Web domain to associate with the key, for keys of `domain` kind.
* `user_id` - (Optional) API key user's string user ID; should be provided for all keys of `web` kind.
* `user_name` - (Optional) API key user's name.

## Attributes Reference

The following attributes are exported:

* `id` - The key ID.
* `description` - Key description.
* `kind` - The type of the key which determines how it can be used.
* `role` - The role of the key which determines its scope in CRUD operations that have role-based access control.
* `domain_name` - The sending domain associated with the key.
* `email` - API key user's email address.
* `requestor` - An email address associated with the key.
* `disabled_reason` - The reason for the key's disablement.
* `expires_at` - When the key will expire.
* `is_disabled` - Whether or not the key is disabled from use.
* `secret` - The full API key secret in plain text.
* `user_id` - API key user's string user ID.
* `user_name` - The API key user's name.
