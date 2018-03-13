package fetch_test

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"testing"
	"time"

	"github.com/axw/pprofbeat/internal/fetch"
	"github.com/stretchr/testify/require"
)

func TestFetch(t *testing.T) {
	server := httptest.NewServer(http.DefaultServeMux)
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go doWork(ctx)

	p, err := fetch.Fetch(fetch.Options{
		URL:      server.URL + "/debug/pprof/profile",
		Duration: time.Second,
	})
	require.NoError(t, err)
	require.NotNil(t, p)
}

func doWork(ctx context.Context) {
	hash := sha256.New()
	buf := make([]byte, 1024)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		rand.Read(buf)
		hash.Write(buf)
	}
}
