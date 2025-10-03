package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/terraform-providers/terraform-provider-mailgun/mailgun"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/terraform-providers/mailgun",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), mailgun.New, opts)

	if err != nil {
		log.Fatal(err.Error())
	}
}
