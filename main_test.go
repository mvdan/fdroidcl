package main

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"text/template"

	"github.com/rogpeppe/go-internal/testscript"
)

func TestMain(m *testing.M) {
	if os.Getenv("TESTSCRIPT_COMMAND") == "" {
		startStaticRepo()
	}

	os.Exit(testscript.RunMain(m, map[string]func() int{
		"fdroidcl": main1,
	}))
}

var staticRepoURL string

func startStaticRepo() {
	path := filepath.Join("testdata", "staticrepo")
	fs := http.FileServer(http.Dir(path))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The files are static, so add a unique etag for each file.
		w.Header().Set("Etag", strconv.Quote(r.URL.Path))
		fs.ServeHTTP(w, r)
	})
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	go http.Serve(ln, handler)
	staticRepoURL = "http://" + ln.Addr().String()
}

var testConfigTmpl = template.Must(template.New("").Parse(`
{
	"repos": [
		{
			"id": "local f-droid",
			"url": "{{.}}",
			"enabled": true
		}
	]
}
`[1:]))

func TestScripts(t *testing.T) {
	t.Parallel()
	testscript.Run(t, testscript.Params{
		Dir: filepath.Join("testdata", "scripts"),
		Setup: func(e *testscript.Env) error {
			home := e.WorkDir + "/home"
			if err := os.MkdirAll(home, 0777); err != nil {
				return err
			}
			e.Vars = append(e.Vars, "HOME="+home)
			e.Vars = append(e.Vars, "REPOURL="+staticRepoURL)

			config := home + "/config.json"
			f, err := os.Create(config)
			if err != nil {
				return err
			}
			if err := testConfigTmpl.Execute(f, staticRepoURL); err != nil {
				return err
			}
			if err := f.Close(); err != nil {
				return err
			}
			e.Vars = append(e.Vars, "FDROIDCL_CONFIG="+config)
			return nil
		},
	})
}
