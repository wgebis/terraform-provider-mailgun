---
page_title: "Provider: Mailgun"
---

# Mailgun Provider

The Mailgun provider is used to interact with the
resources supported by Mailgun. The provider needs to be configured
with the proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

## Example Usage

```hcl
# Configure the Mailgun provider
provider "mailgun" {
  api_key = "${var.mailgun_api_key}"
}

# Create a new domain
resource "mailgun_domain" "default" {
  # ...
}
```

## Argument Reference

The following arguments are supported:

* `api_key` - (Required, Sensitive) Mailgun API key. Can also be supplied via the `MAILGUN_API_KEY` environment variable.

* `requests_per_second` - (Optional) Maximum Mailgun API requests per second for this provider instance. `0` or unset disables pacing. Can also be supplied via the `MAILGUN_REQUESTS_PER_SECOND` environment variable. The limit is per provider process, while Mailgun enforces limits per account, so split the budget across concurrent Terraform runs that share an account.
* `max_retries` - (Optional) Number of times to retry a request that receives HTTP 429, waiting until Mailgun's `X-RateLimit-Reset` time when present and otherwise backing off exponentially (1s, 2s, 4s, ... up to 30s). `0` or unset disables retries. Can also be supplied via the `MAILGUN_MAX_RETRIES` environment variable.
