package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func main() {
	goroot := "/var/run/godoc-root"
	ensureDir(goroot)

	token := strings.TrimSpace(os.Getenv("GITHUB_TOKEN"))
	repos := os.Getenv("GITHUB_REPOS")

	cmd := exec.Command("git", "config", "--global",
		fmt.Sprintf("url.https://%s:x-oauth-basic@github.com/.insteadOf", token),
		"https://github.com/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	sync(goroot, repos, os.Stdout)

	cmd = exec.Command("godoc", "-http", "localhost:8081", "-goroot", ".", "-index", "-play")
	cmd.Dir = goroot
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = "http"
			req.URL.Host = "localhost:8081"
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t := &transformer{wrapped: w}
		proxy.ServeHTTP(t, r)
		t.Transform()
	})

	http.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {
		sync(goroot, repos, w)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func sync(goroot, repos string, w io.Writer) {
	base := filepath.Join(goroot, "src/github.com")

	for _, repo := range strings.Split(repos, ";") {
		repo = strings.TrimSpace(repo)
		w.Write([]byte("Updating " + repo + "=============\n"))
		dir := filepath.Join(base, repo)
		ensureDir(dir)

		vcdir := filepath.Join(dir, ".git")

		var cmd *exec.Cmd
		if dirExists(vcdir) {
			cmd = exec.Command("git", "-C", dir, "pull")
		} else {
			url := fmt.Sprintf("https://github.com/%s", repo)
			cmd = exec.Command("git", "-C", dir, "clone", "--depth", "1", url, ".")
		}
		cmd.Stdout = w
		cmd.Stderr = w
		err := cmd.Run()
		if err != nil {
			w.Write([]byte(err.Error()))
		}

		w.Write([]byte("\n"))
	}
}

func ensureDir(dirname string) {
	err := os.MkdirAll(dirname, 0777)
	if err != nil {
		log.Fatal(err)
	}
}

func dirExists(dirname string) bool {
	info, err := os.Stat(dirname)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

type transformer struct {
	wrapped    http.ResponseWriter
	buffer     bytes.Buffer
	statusCode int
}

func (t *transformer) Header() http.Header {
	return t.wrapped.Header()
}

func (t *transformer) Write(bytes []byte) (int, error) {
	return t.buffer.Write(bytes)
}

func (t *transformer) WriteHeader(statusCode int) {
	t.statusCode = statusCode
	t.wrapped.WriteHeader(statusCode)
}

var re = regexp.MustCompile(`((?:src|href)\s*=\s*)"/([^/])`)

func (t *transformer) Transform() {
	bytes := t.buffer.Bytes()
	contentType := t.wrapped.Header().Get("Content-Type")
	if contentType == "text/html" {
		t.wrapped.Header().Del("Content-Length")
		bytes = re.ReplaceAll(bytes, []byte(`$1"./$2`))
	}
	t.wrapped.Write(bytes)
}
