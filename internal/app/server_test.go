package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"goquizbox/internal/project"
)

func TestServerRoutes(t *testing.T) {
	t.Parallel()

	ctx := project.TestContext(t)

	_, s := newTestServer(t)
	h := s.Routes(ctx)

	cases := []struct {
		name string
		path string
	}{
		{"assets", "/assets/styles.css"},
		{"index", "/"},
		{"apps", "/app"},
		{"health_authority", "/healthauthority/0"},
		{"exports", "/exports/0"},
		{"export_importers", "/export-importers/0"},
		{"mirrors", "/mirrors/0"},
		{"siginfo", "/siginfo/0"},
		{"health", "/health"},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r, err := http.NewRequestWithContext(ctx, http.MethodGet, tc.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			w.Flush()

			if got, want := w.Code, 200; got != want {
				t.Errorf("expected status %d to be %d; headers: %#v; body: %s", got, want, w.Header(), w.Body.String())
			}
		})
	}
}
