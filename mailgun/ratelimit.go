package mailgun

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// rateLimitedTransport paces outgoing requests with a token-bucket limiter and
// retries HTTP 429 responses with backoff. One instance is shared by every
// client the provider creates, so the limit applies to the whole provider
// process (i.e. the whole plan/apply), not per resource. A 429 on any request
// pauses all requests for the retry delay, not just the one that was rejected.
type rateLimitedTransport struct {
	base       http.RoundTripper
	limiter    *rate.Limiter // nil = no pacing
	maxRetries int

	mu          sync.Mutex
	pausedUntil time.Time
}

func newRateLimitedTransport(requestsPerSecond float64, maxRetries int) *rateLimitedTransport {
	t := &rateLimitedTransport{
		base:       http.DefaultTransport,
		maxRetries: maxRetries,
	}
	if requestsPerSecond > 0 {
		// Burst of 1: never send more than one request back-to-back without
		// waiting, which keeps us strictly under a per-second server limit.
		t.limiter = rate.NewLimiter(rate.Limit(requestsPerSecond), 1)
	}
	return t
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	for attempt := 0; ; attempt++ {
		// Re-check the pause after taking a limiter token: a 429 elsewhere may
		// have started one while we were queued, and requests already holding
		// tokens must not all fire the moment it ends.
		for {
			if err := t.waitForPause(ctx); err != nil {
				return nil, err
			}
			if t.limiter != nil {
				if err := t.limiter.Wait(ctx); err != nil {
					return nil, err
				}
			}
			if t.pauseRemaining() <= 0 {
				break
			}
		}

		r := req
		if attempt > 0 {
			// The body was consumed by the previous attempt; rebuild it.
			// net/http sets GetBody for the bytes.Buffer bodies mailgun-go uses.
			if req.Body != nil && req.Body != http.NoBody {
				if req.GetBody == nil {
					return nil, fmt.Errorf("mailgun: cannot retry %s %s: request body is not rewindable", req.Method, req.URL.Path)
				}
				body, err := req.GetBody()
				if err != nil {
					return nil, err
				}
				r = req.Clone(ctx)
				r.Body = body
			}
		}

		resp, err := t.base.RoundTrip(r)
		if err != nil || resp.StatusCode != http.StatusTooManyRequests || attempt >= t.maxRetries {
			return resp, err
		}

		wait := retryDelay(resp, attempt)
		// Drain and close so the connection can be reused.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		t.pauseFor(wait)
		log.Printf("[WARN] Mailgun API returned 429 for %s %s; pausing all requests for %s, retry %d/%d",
			req.Method, req.URL.Path, wait, attempt+1, t.maxRetries)
	}
}

// pauseFor holds off every request on this transport for at least d. An
// existing longer pause is kept.
func (t *rateLimitedTransport) pauseFor(d time.Duration) {
	until := time.Now().Add(d)
	t.mu.Lock()
	if until.After(t.pausedUntil) {
		t.pausedUntil = until
	}
	t.mu.Unlock()
}

func (t *rateLimitedTransport) pauseRemaining() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return time.Until(t.pausedUntil)
}

// waitForPause blocks until no pause is in effect. It loops because another
// 429 can extend the pause while we sleep.
func (t *rateLimitedTransport) waitForPause(ctx context.Context) error {
	for {
		d := t.pauseRemaining()
		if d <= 0 {
			return nil
		}
		if err := sleepCtx(ctx, d); err != nil {
			return err
		}
	}
}

// retryDelay picks how long to wait after a 429: until Mailgun's
// X-RateLimit-Reset time if the header is present and valid, otherwise
// exponential backoff: 1s, 2s, 4s ... capped at 30s.
func retryDelay(resp *http.Response, attempt int) time.Duration {
	if d, ok := rateLimitResetDelay(resp.Header.Get("X-RateLimit-Reset"), time.Now()); ok {
		return d
	}
	d := time.Second << attempt
	if d > 30*time.Second || d <= 0 {
		d = 30 * time.Second
	}
	return d
}

// rateLimitResetDelay parses Mailgun's X-RateLimit-Reset header, documented as
// "Unix milliseconds (UTC) until the limit resets". Values that look like an
// epoch timestamp (after 2001) are treated as the absolute reset time; smaller
// values as milliseconds remaining. A small margin is added either way.
func rateLimitResetDelay(v string, now time.Time) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	ms, err := strconv.ParseInt(v, 10, 64)
	if err != nil || ms < 0 {
		return 0, false
	}
	const margin = 250 * time.Millisecond
	var d time.Duration
	if ms >= 1e12 {
		d = time.UnixMilli(ms).Sub(now)
	} else {
		d = time.Duration(ms) * time.Millisecond
	}
	if d < 0 {
		d = 0
	}
	return d + margin, true
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
