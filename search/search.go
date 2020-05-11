package search

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
)

// searchSyntax is a wrapper for all search terms to improve fuzzy searching
// We use the wildcard query type and then wrap user search terms with *
// meaning search for 0 or more characters
const searchSyntax string = "*%s*"

const mdSuffix = ".md"

// Search represents a index that can be used to look up markdown items
type Search struct {
	path  string               // path where bleve index will be stored (usually user home)
	dir   string               // directory where markdown files exist
	log   map[string]time.Time // a mapping of files to last time they were indexed, so we cut down on file operations
	index bleve.Index
}

type mdfile struct {
	Name    string
	Content string
}

// InitSearch creates a search index object which can then be queried for search results
func InitSearch(path, dir string) (*Search, error) {

	index, err := createOrGetIndex(path)
	if err != nil {
		return nil, err
	}

	return &Search{
		path:  path,
		index: index,
		dir:   dir,
		log:   map[string]time.Time{},
	}, nil
}

// createOrGetIndex returns an index is one already exists or creates a new empty one
func createOrGetIndex(path string) (index bleve.Index, err error) {

	index, err = bleve.Open(path)
	if err != nil {
		log.Println("could not find current index, creating new one")
		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
		if err != nil {
			log.Println("failed to create search index")
			return
		}
	}

	return
}

// BuildIndex will walk the search directory and add files into index
func (si *Search) BuildIndex() error {
	// TODO: Log how long it took to build the index in prometheus
	startTime := time.Now()
	batchCount := 0
	batch := si.index.NewBatch()

	err := filepath.Walk(si.dir, func(path string, info os.FileInfo, err error) error {
		// handle error if failure to access path
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, mdSuffix) {
			name := strings.TrimPrefix(path, si.dir)

			if lastScrapedTime, ok := si.log[name]; ok {
				// We need to check if the last scraped time is after the file's mod time
				// if so we skip it
				if lastScrapedTime.After(info.ModTime()) || lastScrapedTime.Equal(info.ModTime()) {
					log.Printf("skipping file %s since it was recently scrapped", name)
					return nil
				}
			}

			// if file is larger than 3GB skip it
			if info.Size() > 3221225472 {
				log.Printf("file %s is too large(%d) to index skipping", name, info.Size())
				return nil
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				log.Printf("could not open file %s", name)
				return nil
			}

			batch.Index(name, mdfile{
				Name:    name,
				Content: string(content),
			})
			batchCount++
			si.log[name] = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = si.index.Batch(batch)
	if err != nil {
		log.Fatal(err)
	}

	indexDuration := time.Since(startTime)
	indexDurationSeconds := float64(indexDuration) / float64(time.Second)
	timePerDoc := float64(indexDuration) / float64(batchCount)
	log.Printf("Indexed %d documents, in %.2fs (average %.2fms/doc)", batchCount, indexDurationSeconds, timePerDoc/float64(time.Millisecond))
	return nil
}
