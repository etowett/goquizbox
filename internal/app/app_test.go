package app

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"goquizbox/internal/project"
	"goquizbox/internal/serverenv"
	"goquizbox/pkg/database"

	"github.com/gin-gonic/gin"
)

var testDatabaseInstance *database.TestInstance

func TestMain(m *testing.M) {
	testDatabaseInstance = database.MustTestInstance()
	defer testDatabaseInstance.MustClose()

	m.Run()
}

func newTestServer(t testing.TB) (*serverenv.ServerEnv, *Server) {
	t.Helper()

	ctx := project.TestContext(t)
	testDB, _ := testDatabaseInstance.NewDatabase(t)

	env := serverenv.New(ctx,
		serverenv.WithDatabase(testDB))

	config := &Config{}

	server, err := NewServer(config, env)
	if err != nil {
		t.Fatalf("error creating test server: %v", err)
	}

	return env, server
}

// Reflectively serialize the fields in f into form
// fields on the https request, r.
func serializeForm(i interface{}) (url.Values, error) {
	if i == nil {
		return url.Values{}, nil
	}

	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("provided interface is not a pointer")
	}

	if v.IsNil() {
		return url.Values{}, nil
	}

	e := v.Elem()
	if e.Kind() != reflect.Struct {
		return nil, fmt.Errorf("provided interface is not a struct")
	}

	t := e.Type()

	form := url.Values{}
	for i := 0; i < t.NumField(); i++ {
		ef := e.Field(i)
		tf := t.Field(i)
		tag := tf.Tag.Get("form")

		if ef.Kind() == reflect.Slice || ef.Kind() == reflect.Array {
			for i := 0; i < ef.Len(); i++ {
				form.Add(tag, fmt.Sprintf("%v", ef.Index(i)))
			}
		} else {
			form.Add(tag, fmt.Sprintf("%v", ef))
		}
	}
	return form, nil
}

func newHTTPServer(t testing.TB, method string, path string, handler gin.HandlerFunc) *httptest.Server {
	t.Helper()

	tmpl, err := template.New("").
		Option("missingkey=zero").
		Funcs(TemplateFuncMap).
		ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		t.Fatalf("failed to parse templates from fs: %v", err)
	}

	r := gin.Default()
	r.SetFuncMap(TemplateFuncMap)
	r.SetHTMLTemplate(tmpl)
	switch method {
	case http.MethodGet:
		r.GET(path, handler)
	case http.MethodPost:
		r.POST(path, handler)
	default:
		t.Fatalf("unsupported http method: %v", method)
	}

	return httptest.NewServer(r)
}

func mustFindStrings(t testing.TB, resp *http.Response, want ...string) {
	t.Helper()
	if len(want) == 0 {
		t.Error("not checking for any strings, error in test?")
	}

	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response: %v", err)
	}

	result := string(bytes)

	for _, wants := range want {
		if !strings.Contains(result, wants) {
			t.Errorf("result missing expected string: %v, got: %v", wants, result)
		}
	}
}

func intPtr(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func int8Ptr(i int8) *int8 {
	if i == 0 {
		return nil
	}
	return &i
}

func int16Ptr(i int16) *int16 {
	if i == 0 {
		return nil
	}
	return &i
}

func int32Ptr(i int32) *int32 {
	if i == 0 {
		return nil
	}
	return &i
}

func int64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func uintPtr(i uint) *uint {
	if i == 0 {
		return nil
	}
	return &i
}

func uint8Ptr(i uint8) *uint8 {
	if i == 0 {
		return nil
	}
	return &i
}

func uint16Ptr(i uint16) *uint16 {
	if i == 0 {
		return nil
	}
	return &i
}

func uint32Ptr(i uint32) *uint32 {
	if i == 0 {
		return nil
	}
	return &i
}

func uint64Ptr(i uint64) *uint64 {
	if i == 0 {
		return nil
	}
	return &i
}

func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}
