package mailgun

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

// Server allows at most `limit` requests per rolling second and returns 429
// otherwise, mimicking Mailgun's account-smtp-creds-requests-per-sec limit.
func newThrottlingServer(t *testing.T, limit int) (*httptest.Server, *int64, *int64) {
	var mu sync.Mutex
	var window []time.Time
	var ok, throttled int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		now := time.Now()
		cut := now.Add(-time.Second)
		kept := window[:0]
		for _, ts := range window {
			if ts.After(cut) {
				kept = append(kept, ts)
			}
		}
		window = kept
		if len(window) >= limit {
			mu.Unlock()
			atomic.AddInt64(&throttled, 1)
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message":"Rate Limited"}`))
			return
		}
		window = append(window, now)
		mu.Unlock()
		atomic.AddInt64(&ok, 1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"items":[{"login":"user@example.com"}]}`))
	}))
	t.Cleanup(srv.Close)
	return srv, &ok, &throttled
}

// testClient returns a client from cfg pointed at the test server. A failed
// SetAPIBase is fatal so tests can never fall through to the real API.
func testClient(t *testing.T, cfg *Config, base string) *mailgun.Client {
	t.Helper()
	c, err := cfg.GetClient("us")
	if err != nil {
		t.Fatalf("GetClient: %v", err)
	}
	if err := c.SetAPIBase(base); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return c
}

func listOnce(ctx context.Context, c *mailgun.Client) error {
	it := c.ListCredentials("example.com", nil)
	var page []mtypes.Credential
	it.First(ctx, &page)
	return it.Err()
}

func TestUnpacedGets429(t *testing.T) {
	srv, _, throttled := newThrottlingServer(t, 10)
	cfg := &Config{APIKey: "k"}
	var errs int
	for i := 0; i < 40; i++ {
		if err := listOnce(context.Background(), testClient(t, cfg, srv.URL)); err != nil {
			errs++
			if !strings.Contains(err.Error(), "429") {
				t.Fatalf("unexpected error: %v", err)
			}
		}
	}
	if errs == 0 || atomic.LoadInt64(throttled) == 0 {
		t.Fatalf("expected 429s without pacing, got errs=%d throttled=%d", errs, atomic.LoadInt64(throttled))
	}
	t.Logf("unpaced: %d/40 requests failed with 429", errs)
}

func TestPacedNo429(t *testing.T) {
	srv, ok, throttled := newThrottlingServer(t, 10)
	cfg := &Config{APIKey: "k", RequestsPerSecond: 8, MaxRetries: 5}
	start := time.Now()
	// Concurrent callers, like terraform -parallelism=10, share one limiter.
	var wg sync.WaitGroup
	var errs int64
	for i := 0; i < 40; i++ {
		c := testClient(t, cfg, srv.URL)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := listOnce(context.Background(), c); err != nil {
				atomic.AddInt64(&errs, 1)
			}
		}()
	}
	wg.Wait()
	if errs != 0 || atomic.LoadInt64(throttled) != 0 || atomic.LoadInt64(ok) != 40 {
		t.Fatalf("paced: errs=%d throttled=%d ok=%d", errs, atomic.LoadInt64(throttled), atomic.LoadInt64(ok))
	}
	t.Logf("paced: 40 requests, 0 throttled, %s", time.Since(start).Round(time.Millisecond))
}

func TestRetryRecovers429(t *testing.T) {
	// Pacing off, retries on: 429s must be absorbed by the retry loop.
	srv, ok, throttled := newThrottlingServer(t, 10)
	cfg := &Config{APIKey: "k", MaxRetries: 6}
	var errs int
	for i := 0; i < 30; i++ {
		if err := listOnce(context.Background(), testClient(t, cfg, srv.URL)); err != nil {
			errs++
		}
	}
	if errs != 0 || atomic.LoadInt64(throttled) == 0 {
		t.Fatalf("retry: errs=%d throttled=%d ok=%d", errs, atomic.LoadInt64(throttled), atomic.LoadInt64(ok))
	}
	t.Logf("retry-only: 30 requests succeeded, %d 429s absorbed", atomic.LoadInt64(throttled))
}

func TestRetryRewindsPOSTBody(t *testing.T) {
	var calls int64
	var bodies []string
	var mu sync.Mutex
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		mu.Lock()
		bodies = append(bodies, r.PostForm.Get("login"))
		mu.Unlock()
		if atomic.AddInt64(&calls, 1) == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		_, _ = w.Write([]byte(`{"message":"Created 1 credentials pair(s)"}`))
	}))
	defer srv.Close()
	cfg := &Config{APIKey: "k", MaxRetries: 3}
	c := testClient(t, cfg, srv.URL)
	if err := c.CreateCredential(context.Background(), "example.com", "user@example.com", "s3cretpass"); err != nil {
		t.Fatalf("create after 429: %v", err)
	}
	if len(bodies) != 2 || bodies[0] != bodies[1] || bodies[1] != "user@example.com" {
		t.Fatalf("body not replayed correctly: %q", bodies)
	}
}

func TestRetryPausesOtherRequests(t *testing.T) {
	// The first request gets a 429 with X-RateLimit-Reset; a second request started
	// during the pause must not reach the server until the pause has ended.
	var calls int64
	var mu sync.Mutex
	var arrivals []time.Time
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		arrivals = append(arrivals, time.Now())
		mu.Unlock()
		if atomic.AddInt64(&calls, 1) == 1 {
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(time.Second).UnixMilli(), 10))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"total_count":1,"items":[{"login":"user@example.com"}]}`))
	}))
	defer srv.Close()

	// Pacing off so only the pause can delay the second request.
	cfg := &Config{APIKey: "k", MaxRetries: 3}
	tr := cfg.sharedHTTPClient().Transport.(*rateLimitedTransport)

	first := make(chan error, 1)
	go func() { first <- listOnce(context.Background(), testClient(t, cfg, srv.URL)) }()

	deadline := time.Now().Add(5 * time.Second)
	for tr.pauseRemaining() <= 0 {
		if time.Now().After(deadline) {
			t.Fatal("transport never paused after 429")
		}
		time.Sleep(5 * time.Millisecond)
	}
	tr.mu.Lock()
	pausedUntil := tr.pausedUntil
	tr.mu.Unlock()

	if err := listOnce(context.Background(), testClient(t, cfg, srv.URL)); err != nil {
		t.Fatalf("second request: %v", err)
	}
	if err := <-first; err != nil {
		t.Fatalf("first request: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if len(arrivals) != 3 {
		t.Fatalf("expected 3 server hits (429, retry, second), got %d", len(arrivals))
	}
	for i, at := range arrivals[1:] {
		if at.Before(pausedUntil) {
			t.Fatalf("request %d reached server %s before pause ended", i+2, pausedUntil.Sub(at))
		}
	}
}

func TestRetryDelay(t *testing.T) {
	now := time.Now()
	header := func(kv ...string) *http.Response {
		h := http.Header{}
		for i := 0; i < len(kv); i += 2 {
			h.Set(kv[i], kv[i+1])
		}
		return &http.Response{Header: h}
	}
	within := func(name string, got, want time.Duration) {
		t.Helper()
		if diff := got - want; diff < -100*time.Millisecond || diff > 100*time.Millisecond {
			t.Errorf("%s: got %s, want ~%s", name, got, want)
		}
	}

	within("reset absolute epoch ms",
		retryDelay(header("X-RateLimit-Reset", strconv.FormatInt(now.Add(3*time.Second).UnixMilli(), 10)), 0),
		3*time.Second+250*time.Millisecond)
	within("reset relative ms",
		retryDelay(header("X-RateLimit-Reset", "2000"), 0),
		2*time.Second+250*time.Millisecond)
	within("reset already passed",
		retryDelay(header("X-RateLimit-Reset", strconv.FormatInt(now.Add(-time.Minute).UnixMilli(), 10)), 0),
		250*time.Millisecond)
	within("garbage reset falls back to backoff",
		retryDelay(header("X-RateLimit-Reset", "soon"), 1), 2*time.Second)
	within("no headers backs off exponentially",
		retryDelay(header(), 2), 4*time.Second)
	within("backoff capped at 30s",
		retryDelay(header(), 10), 30*time.Second)
}
