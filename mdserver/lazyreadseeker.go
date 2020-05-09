// Lazyreadseeker was created to improve performance on requests for markdown pages.
//
// Typically clients will request a page and pass the if-modified-since header to
// decide if they should just load the page from cache or download and render a new
// version.
//
// The http.ServeContent method handles if-modified-since requests, but to do so you have to pass
// it a readseeker. Typically a readseeker needs to have the full content already present when
// instantiated. This becomes wasteful if http.ServeContent(which is called every time a file is requested)
// ends up not actually reading the file due to the client having the most recent version.
//
// To get around this we create a wrapper around readseeker and only load the file into memory if
// the readseeker actually ends up needing to read.

package mdserver

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
)

type lazyReadSeeker struct {
	filePath string
	options  htmlPageOptions
	reader   *bytes.Reader // initially nil, initialized with init()
}

type htmlPageOptions struct {
	theme string
}

func newLazyReadSeeker(filepath string, options htmlPageOptions) (*lazyReadSeeker, time.Time, error) {
	file, err := os.Stat(filepath)
	if err != nil {
		return nil, time.Time{}, err
	}

	return &lazyReadSeeker{
		filePath: filepath,
		options:  options,
	}, file.ModTime(), nil
}

func (l *lazyReadSeeker) init() error {
	if l.reader != nil {
		return nil
	}
	content, err := ioutil.ReadFile(l.filePath)
	if err != nil {
		return err
	}

	html, err := compileMDToHTML(filepath.Base(l.filePath), l.options.theme, content)
	if err != nil {
		return err
	}

	l.reader = bytes.NewReader(html)
	return nil
}

func (l *lazyReadSeeker) Read(p []byte) (n int, err error) {
	if l.reader == nil {
		if err := l.init(); err != nil {
			return 0, err
		}
	}
	return l.reader.Read(p)
}

func (l *lazyReadSeeker) Seek(offset int64, whence int) (int64, error) {
	if l.reader == nil {
		if err := l.init(); err != nil {
			return 0, err
		}
	}
	return l.reader.Seek(offset, whence)
}

func compileMDToHTML(title, theme string, content []byte) ([]byte, error) {

	// set some common extensions and render settings
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs ^ parser.MathJax
	parsedMarkdown := parser.NewWithExtensions(extensions).Parse(content)
	body := markdown.Render(parsedMarkdown, html.NewRenderer(html.RendererOptions{Title: title, Flags: html.CommonFlags}))

	// use bluemonday to sanatize html content
	policy := bluemonday.UGCPolicy().AllowAttrs("class").OnElements("code")
	body = policy.SanitizeBytes(body)

	page := struct {
		Title string
		Body  template.HTML
		Style string
	}{
		Title: title,
		Body:  template.HTML(body),
		Style: theme,
	}

	buf := bytes.NewBuffer(content[:0]) // reuse content to reduce allocations
	err := compiledPageTemplate.Execute(buf, page)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
