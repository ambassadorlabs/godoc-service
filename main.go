package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"path/filepath"
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

	http.Handle("/", proxy)

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
