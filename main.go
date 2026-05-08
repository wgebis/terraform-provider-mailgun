package main

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6/tf6server"

	"github.com/wgebis/terraform-provider-mailgun/internal/framework"
)

const providerAddress = "registry.terraform.io/wgebis/mailgun"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers")
	flag.Parse()

	ctx := context.Background()

	server, err := framework.MuxedProviderServer(ctx)
	if err != nil {
		log.Fatal(err)
	}

	var serveOpts []tf6server.ServeOpt
	if debug {
		serveOpts = append(serveOpts, tf6server.WithManagedDebug())
	}

	if err := tf6server.Serve(providerAddress, func() tfprotov6.ProviderServer { return server }, serveOpts...); err != nil {
		log.Fatal(err)
	}
}
