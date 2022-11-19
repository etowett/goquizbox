package app

// import (
// 	"io"
// 	"net/http"
// 	"testing"

// 	exportmodel "toiler/internal/export/model"
// 	exportimportmodel "toiler/internal/exportimport/model"
// 	mirrormodel "toiler/internal/mirror/model"
// 	"toiler/internal/project"
// 	verificationmodel "toiler/internal/verification/model"
// )

// func TestRenderIndex(t *testing.T) {
// 	t.Parallel()

// 	m := TemplateMap{}
// 	m["healthauthorities"] = []*verificationmodel.HealthAuthority{}
// 	m["exports"] = []*exportmodel.ExportConfig{}
// 	m["exportImporters"] = []*exportimportmodel.ExportImport{}
// 	m["siginfos"] = []*exportmodel.SignatureInfo{}
// 	m["mirrors"] = []*mirrormodel.Mirror{}

// 	testRenderTemplate(t, "index", m)
// }

// func TestHandleIndex(t *testing.T) {
// 	t.Parallel()

// 	ctx := project.TestContext(t)

// 	env, s := newTestServer(t)
// 	db := env.Database()
// 	_ = db

// 	cases := []struct {
// 		name   string
// 		status int
// 		want   []string
// 	}{
// 		{
// 			name:   "default",
// 			status: 200,
// 			want:   []string{"Admin Console"},
// 		},
// 	}

// 	for _, tc := range cases {
// 		tc := tc
// 		t.Run(tc.name, func(t *testing.T) {
// 			t.Parallel()

// 			server := newHTTPServer(t, http.MethodGet, "/", s.HandleIndex())

// 			req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
// 			if err != nil {
// 				t.Fatal(err)
// 			}
// 			client := server.Client()

// 			resp, err := client.Do(req)
// 			if err != nil {
// 				t.Fatalf("error making http call: %v", err)
// 			}
// 			defer resp.Body.Close()

// 			if got, want := resp.StatusCode, tc.status; got != want {
// 				b, err := io.ReadAll(resp.Body)
// 				if err != nil {
// 					t.Fatal(err)
// 				}
// 				t.Errorf("expected status %d to be %d; headers: %#v; body: %s", got, want, resp.Header, b)
// 			}

// 			if len(tc.want) > 0 {
// 				mustFindStrings(t, resp, tc.want...)
// 			}
// 		})
// 	}
// }
