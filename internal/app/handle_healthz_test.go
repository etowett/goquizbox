package app

import (
	"io"
	"net/http"
	"testing"

	"goquizbox/internal/project"
)

func TestHandleHealthz(t *testing.T) {
	t.Parallel()

	ctx := project.TestContext(t)

	_, s := newTestServer(t)
	server := newHTTPServer(t, http.MethodGet, "/", s.HandleHealthz())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatal(err)
	}
	client := server.Client()

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("error making http call: %v", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, 200; got != want {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		t.Errorf("expected status %d to be %d; headers: %#v; body: %s", got, want, resp.Header, b)
	}

	mustFindStrings(t, resp, "ok")
}
