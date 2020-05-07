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
	dir string // the directory path where markdown files are stored
}

// Run starts the markdown webserver
// 	dir: the directory where the markdown files live
// 	addr: <host>:<port> that webserver will listen on
// 	open: name of the file to open to, if not provided a default home page will be provided
func Run(dir, addr, open string) error {
	mdHandler := &mdHandler{
		dir: dir,
	}

	srv := http.Server{
		Addr:        addr,
		Handler:     httpgzip.New(mdHandler),
		ReadTimeout: time.Second,
	}

	if open != "" {
		go func() {
			time.Sleep(100 * time.Millisecond)
			browser.OpenURL(fmt.Sprintf("http://%s/%s", addr, open))
		}()
	}
	return srv.ListenAndServe()
}

func (handler *mdHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Frame-Options", "SAMEORIGIN") // prevent site from being framed by other sites

	urlPath := path.Clean(r.URL.Path)
	if containsDotDot(urlPath) {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}

	//We need to handle fileserver interactions here to assets folder

	filePath := filepath.Join(handler.dir, filepath.FromSlash(urlPath))

	reader, modTime, err := newLazyReadSeeker(filePath)
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
