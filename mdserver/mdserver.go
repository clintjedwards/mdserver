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
	"github.com/dustin/go-humanize"
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

	urlPath := path.Clean(r.URL.Path)
	if containsDotDot(urlPath) {
		http.Error(w, "invalid URL path", http.StatusBadRequest)
		return
	}

	// If this is index serve special index page
	if urlPath == "/" {
		files, err := getDirFileInfo(handler.dir)
		if err != nil {
			http.Error(w, "could not serve index", http.StatusBadGateway)
			return
		}

		page := struct {
			Title string
			Style string
			Files []fileInfo
		}{
			Title: "Index",
			Style: handler.theme,
			Files: files,
		}

		err = compiledIndexTemplate.Execute(w, page)
		if err != nil {
			http.Error(w, "could not serve index", http.StatusBadGateway)
			return
		}
		return
	}

	// If it's not a md file see if it can be served by the fileserver
	if !strings.HasSuffix(r.URL.Path, mdSuffix) {
		handler.fileServer.ServeHTTP(w, r)
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
	return
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

// getDirFileInfo
func getDirFileInfo(rootDir string) ([]fileInfo, error) {

	var files []fileInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		// handle error if failure to access path
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, mdSuffix) {
			files = append(files, fileInfo{
				Name:     info.Name(),
				Modified: humanize.Time(info.ModTime()),
				Size:     humanize.Bytes(uint64(info.Size())),
				Path:     strings.TrimPrefix(path, rootDir),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return files, nil
}

type fileInfo struct {
	Name     string
	Modified string
	Size     string
	Path     string
}
