package multiminer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// probeHTTP tries a list of path candidates on the given base (host:port) and returns the first path with 200 OK.
func probeHTTP(ctx context.Context, address string, paths []string, timeout time.Duration) (string, bool) {
	client := &http.Client{Timeout: timeout}
	for _, p := range paths {
		url := fmt.Sprintf("http://%s%s", address, p)
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		_ = resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return p, true
		}
	}
	return "", false
}

// joinPath safely concatenates a base path and sub-path.
func joinPath(a, b string) string {
	if strings.HasSuffix(a, "/") && strings.HasPrefix(b, "/") {
		return a + b[1:]
	}
	if !strings.HasSuffix(a, "/") && !strings.HasPrefix(b, "/") {
		return a + "/" + b
	}
	return a + b
}
