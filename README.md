Terraform Provider
==================

- Website: https://registry.terraform.io/providers/wgebis/mailgun/
- [![Gitter chat](https://badges.gitter.im/hashicorp-terraform/Lobby.png)](https://gitter.im/hashicorp-terraform/Lobby)
- Mailing list: [Google Groups](http://groups.google.com/group/terraform-tool)

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.0+
-	[Go](https://golang.org/doc/install) 1.25+ (to build the provider plugin)

Architecture
------------

The provider is being migrated from `terraform-plugin-sdk/v2` to
`terraform-plugin-framework`. Both runtimes are served from a single binary
through `terraform-plugin-mux` (`tf6muxserver`), so resources can be moved
across one at a time without breaking existing configurations.

| Resource / data source | Runtime |
|---|---|
| `mailgun_domain` (resource + data source) | terraform-plugin-framework |
| `mailgun_route` | terraform-plugin-sdk/v2 |
| `mailgun_domain_credential` | terraform-plugin-sdk/v2 |
| `mailgun_webhook` | terraform-plugin-sdk/v2 |
| `mailgun_api_key` | terraform-plugin-sdk/v2 |

The `mailgun_domain` schema was bumped to version `1`; a state upgrader
handles state produced by previous releases (it drops the deprecated
`sending_records` / `receiving_records` lists in favour of the
`*_records_set` attributes).

Building The Provider
---------------------

Clone repository to: `$GOPATH/src/github.com/wgebis/terraform-provider-mailgun`

```sh
$ mkdir -p $GOPATH/src/github.com/wgebis; cd $GOPATH/src/github.com/wgebis
$ git clone git@github.com:wgebis/terraform-provider-mailgun
```

Enter the provider directory and build the provider

```sh
$ cd $GOPATH/src/github.com/wgebis/terraform-provider-mailgun
$ make build
```

Using the provider
----------------------

https://registry.terraform.io/providers/wgebis/mailgun/latest/docs

Developing the Provider
----------------------

If you wish to work on the provider, you'll first need [Go](http://www.golang.org) installed on your machine (version 1.8+ is *required*). You'll also need to correctly setup a [GOPATH](http://golang.org/doc/code.html#GOPATH), as well as adding `$GOPATH/bin` to your `$PATH`.

To compile the provider, run `make build`. This will build the provider and put the provider binary in the `$GOPATH/bin` directory.

```sh
$ make build
...
$ $GOPATH/bin/terraform-provider-mailgun
...
```

In order to test the provider, you can simply run `make test`.

```sh
$ make test
```

In order to run the full suite of Acceptance tests, run `make testacc`.

*Note:* Acceptance tests create real resources, and often cost money to run.

```sh
$ make testacc
```
