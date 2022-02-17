package mdserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/artyom/httpgzip"
	"github.com/clintjedwards/mdserver/search"
	"github.com/dustin/go-humanize"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/pkg/browser"
	"github.com/rs/zerolog/log"
	"embed"
)

//go:embed public
var assets embed.FS

const mdSuffix = ".md"

// MDServer is the main entry point of the markdown server
type MDServer struct {
	search  *search.Search
	options ServerOptions
}

// NewMDServer creates a new instance of a markdown server which can then be started
func NewMDServer(options ServerOptions) *MDServer {
	search, err := search.InitSearch(options.Dir)
	if err != nil {
		log.Fatal().Err(err).Msg("error initializing search")
	}

	go func() {
		for {
			err := search.BuildIndex()
			if err != nil {
				log.Error().Err(err).Msg("could not build index")
			}
			time.Sleep(60 * time.Minute)
		}
	}()

	return &MDServer{
		search:  search,
		options: options,
	}
}

// mdHandler is a http handler implementation specifically for serving markdown content
type mdHandler struct {
	dir        string // the directory path where markdown files are stored
	fileServer http.Handler
	theme      string
}

// ServerOptions define a set of options that are used to configure the mdserver
type ServerOptions struct {
	Dir   string //dir: the directory where the markdown files live
	Addr  string //addr: <host>:<port> that webserver will listen on
	Open  string //open: name of the file to open to, if not provided a default home page will be provided
	Theme string //theme: css theme for frontend
}

// Run starts the mdserver
func (mdserver *MDServer) Run() error {
	mdHandler := &mdHandler{
		dir: mdserver.options.Dir,
		// We bake frontend files directly into the binary
		// assets is an implementation of an http.filesystem created by
		// github.com/shurcooL/vfsgen that points to the "public" folder
		fileServer: http.FileServer(http.FS(assets)),
		theme:      mdserver.options.Theme,
	}

	router := mux.NewRouter()

	// route matches are evaluated in the order they are added (first match)
	router.Handle("/api/search", handlers.MethodHandler{
		"GET": http.HandlerFunc(mdserver.searchHandler),
	})
	router.PathPrefix("/").Handler(httpgzip.New(mdHandler))

	srv := http.Server{
		Addr:        mdserver.options.Addr,
		Handler:     router,
		ReadTimeout: time.Second,
	}

	if mdserver.options.Open != "" {
		go func() {
			time.Sleep(100 * time.Millisecond)
			browser.OpenURL(fmt.Sprintf("http://%s/%s", mdserver.options.Addr, mdserver.options.Open))
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

func formatName(path string) string {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, mdSuffix)

	return path
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
				Name:     formatName(strings.TrimPrefix(path, rootDir)),
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

func (mdserver *MDServer) searchHandler(w http.ResponseWriter, req *http.Request) {
	term := req.FormValue("term")
	hits, err := mdserver.search.Query(term)
	if err != nil {
		sendJSONErrResponse(w, http.StatusBadGateway, err)
		return
	}

	sendJSONResponse(w, http.StatusOK, hits)
	return
}

// sendJSONResponse converts raw objects and parameters to a json response and passes it to a provided writer
func sendJSONResponse(w http.ResponseWriter, httpStatusCode int, payload interface{}) {
	w.WriteHeader(httpStatusCode)

	enc := json.NewEncoder(w)
	err := enc.Encode(payload)
	if err != nil {
		log.Error().Err(err).Msg("could not send JSON response")
	}
}

// sendJSONErrResponse converts raw objects and parameters to a json response and passes it to a provided writer
func sendJSONErrResponse(w http.ResponseWriter, httpStatusCode int, errStr error) {
	w.WriteHeader(httpStatusCode)

	enc := json.NewEncoder(w)
	err := enc.Encode(map[string]string{"err": errStr.Error()})
	if err != nil {
		log.Error().Err(err).Msg("could not send JSON response")
	}
}
