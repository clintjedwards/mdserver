package mdserver

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/artyom/httpgzip"
	"github.com/pkg/browser"
)

const mdSuffix = ".md"

// mdHandler is a http handler implementation specifically for markdown serving
type mdHandler struct {
	dir        string // the directory path where markdown files are stored
	fileServer http.Handler
	theme      string
}

// RunOptions define a set of options that are used to configure the mdserver
type RunOptions struct {
	Dir   string
	Addr  string
	Open  string
	Theme string
}

// Run starts the markdown webserver
// 	dir: the directory where the markdown files live
// 	addr: <host>:<port> that webserver will listen on
// 	open: name of the file to open to, if not provided a default home page will be provided
func Run(options RunOptions) error {
	mdHandler := &mdHandler{
		dir: options.Dir,
		// We bake frontend files directly into the binary
		// assets is an implementation of an http.filesystem created by
		// github.com/shurcooL/vfsgen that points to the "public" folder
		fileServer: http.FileServer(assets),
		theme:      options.Theme,
	}

	srv := http.Server{
		Addr:        options.Addr,
		Handler:     httpgzip.New(mdHandler),
		ReadTimeout: time.Second,
	}

	if options.Open != "" {
		go func() {
			time.Sleep(100 * time.Millisecond)
			browser.OpenURL(fmt.Sprintf("http://%s/%s", options.Addr, options.Open))
		}()
	}
	return srv.ListenAndServe()
}

func (handler *mdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Frame-Options", "SAMEORIGIN") // prevent site from being framed by other sites

	// If it's not a md file see if it can be served by the fileserver
	if !strings.HasSuffix(r.URL.Path, mdSuffix) {
		handler.fileServer.ServeHTTP(w, r)
		return
	}

	urlPath := path.Clean(r.URL.Path)
	if containsDotDot(urlPath) {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}

	filePath := filepath.Join(handler.dir, filepath.FromSlash(urlPath))

	reader, modTime, err := newLazyReadSeeker(filePath, htmlPageOptions{theme: handler.theme})
	if err != nil {
		if os.IsNotExist(err) {
			http.NotFound(w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, "page.html", modTime, reader)
}

func containsDotDot(v string) bool {
	if !strings.Contains(v, "..") {
		return false
	}
	for _, ent := range strings.FieldsFunc(v, func(r rune) bool { return r == '/' || r == '\\' }) {
		if ent == ".." {
			return true
		}
	}
	return false
}
