package framework

import (
	"context"
	"fmt"
	"log"
	"time"

	mailgunpkg "github.com/wgebis/terraform-provider-mailgun/mailgun"
)

// waitForDeletion polls probe until the resource is gone (Mailgun is
// eventually consistent after a delete call).
//
// probe is expected to fetch the resource; a Mailgun 404 means the deletion
// has propagated. Any other error is returned as-is rather than being treated
// as a successful deletion, so a transient failure during polling does not
// cause Terraform to drop a resource that may still exist. If the resource is
// still present when the timeout elapses, a timeout error is returned.
func waitForDeletion(ctx context.Context, kind, id string, timeout, interval time.Duration, probe func() error) error {
	deadline := time.Now().Add(timeout)
	for {
		err := probe()
		if mailgunpkg.IsNotFound(err) {
			return nil
		}
		if err != nil {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("%s %s still exists after %s", kind, id, timeout)
		}
		log.Printf("[INFO] Retrying until %s disappears...", kind)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}
	}
}
