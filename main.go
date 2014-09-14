package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"bitbucket.org/kardianos/osext"
	"github.com/aarzilli/golua/lua"
)

var (
	bind    = flag.String("bind", ":8080", "[host]:port where to serve on")
	asserts = flag.String("asserts", getWorkDir(), "path to assers")
	//verbose = flag.Bool("verbose", false, "path to assers")
)

func getWorkDir() string {
	p, err := build.Default.Import("github.com/nordicdyno/luawshop2014-gopher-on-lua", "", build.FindOnly)
	if err != nil {
		filename, _ := osext.Executable()
		return filepath.Join(filepath.Dir(filename), "resources")
	}

	return p.Dir
}

func init() {
	flag.Parse()
}

type appHandler struct{}

func main() {
	h := appHandler{}
	http.Handle("/", h)

	if err := http.ListenAndServe(*bind, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}

func (h appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println("Catch error. Recovering...")
			var doc bytes.Buffer
			err := errorTemplate.Execute(&doc, &ErrorPage{
				Code:    http.StatusInternalServerError,
				Message: rec,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html;charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, doc.String())
		}
	}()

	dir := *asserts
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			panic("Dir" + dir + " not exists")
		}
		panic(err)
	}

	if r.Method == "GET" {
		if strings.HasSuffix(r.URL.Path, ".lua") {
			serveLua(dir, w, r)
			return
		}
		fs := http.FileServer(http.Dir(dir))
		fs.ServeHTTP(w, r)
		return
	} else {
		http.Error(w, "Invalid request method.", 405)
	}
}

func serveLua(dir string, w http.ResponseWriter, r *http.Request) {
	file := filepath.Join(dir, r.URL.Path)

	content, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}

	luaPrint := func(L *lua.State) int {
		s := L.ToString(1)
		io.WriteString(w, s)
		io.WriteString(w, "\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		return 0
	}

	L := lua.NewState()
	defer L.Close()
	L.OpenLibs()
	L.Register("print", luaPrint)
	L.MustDoString(string(content))
}

type ErrorPage struct {
	Code    int
	Message interface{}
}

var errorTemplate = template.Must(template.New("").Parse(`
<html><body>
<h2>This app is crashed with error:</h2>
<h2>Code: {{.Code}}<br>
Message: «{{.Message}}»
</h2>
</body></html>
`))
