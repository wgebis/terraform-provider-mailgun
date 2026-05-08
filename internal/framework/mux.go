package framework

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-mux/tf5to6server"
	"github.com/hashicorp/terraform-plugin-mux/tf6muxserver"

	"github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// MuxedProviderServer returns a tfprotov6.ProviderServer factory that combines
// the legacy SDKv2 mailgun provider (upgraded from protocol v5) with the new
// terraform-plugin-framework provider. It is consumed by the binary entrypoint
// in main.go and by acceptance tests.
func MuxedProviderServer(ctx context.Context) (tfprotov6.ProviderServer, error) {
	upgraded, err := tf5to6server.UpgradeServer(ctx, mailgun.Provider().GRPCProvider)
	if err != nil {
		return nil, err
	}

	mux, err := tf6muxserver.NewMuxServer(ctx,
		func() tfprotov6.ProviderServer { return upgraded },
		providerserver.NewProtocol6(New()()),
	)
	if err != nil {
		return nil, err
	}
	return mux.ProviderServer(), nil
}
