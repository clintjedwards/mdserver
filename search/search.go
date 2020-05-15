package search

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/blevesearch/bleve"
	"github.com/rs/zerolog/log"
)

// searchSyntax is a wrapper for all search terms to improve fuzzy searching
// We use the wildcard query type and then wrap user search terms with *
// meaning search for 0 or more characters
const searchSyntax string = "*%s*"

const mdSuffix = ".md"
const searchFileName = "mdserver.index"

// Search represents a index that can be used to look up markdown items
type Search struct {
	path        string               // path where bleve index will be stored (usually user home)
	dir         string               // directory where markdown files exist
	lastScraped map[string]time.Time // mapping of files to last time they were indexed; used to reduce file operations
	bleve.Index                      // bleve search index
}

// mdfile is used to index the metadata and content of markdown files
type mdfile struct {
	Name    string
	Content string
}

// InitSearch creates a search index object which can then be queried for search results
func InitSearch(dir string) (*Search, error) {

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/.%s", homeDir, searchFileName)

	index, err := getOrCreateIndex(path)
	if err != nil {
		return nil, err
	}

	return &Search{
		path,
		dir,
		map[string]time.Time{},
		index,
	}, nil
}

// getOrCreateIndex returns an index if one already exists or creates a new empty one
func getOrCreateIndex(path string) (index bleve.Index, err error) {

	index, err = bleve.Open(path)
	if err != nil {
		log.Info().Msg("could not find current index, creating new one")

		mapping := bleve.NewIndexMapping()
		index, err = bleve.New(path, mapping)
		if err != nil {
			return nil, err
		}
	}

	return index, nil
}

// BuildIndex will walk the search directory and add markdown files into index
func (si *Search) BuildIndex() error {
	startTime := time.Now()
	count := 0
	batch := si.NewBatch()

	err := filepath.Walk(si.dir, func(path string, info os.FileInfo, err error) error {
		// handle error if failure to access path
		if err != nil {
			return err
		}
		if strings.HasSuffix(path, mdSuffix) {
			name := strings.TrimPrefix(path, si.dir)

			if lastScrapedTime, ok := si.lastScraped[name]; ok {
				// check if the last scraped time is after the file's mod time
				// if so skip it
				if lastScrapedTime.After(info.ModTime()) || lastScrapedTime.Equal(info.ModTime()) {
					log.Debug().Msgf("skipping file %s since it has not changed since last scrape", name)
					return nil
				}
			}

			// if file is larger than 1GB skip it
			if info.Size() > 1073741824 {
				log.Debug().Msgf("file %s is too large(%db) to index skipping", name, info.Size())
				return nil
			}

			content, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			batch.Index(name, mdfile{
				Name:    name,
				Content: string(content),
			})
			count++
			si.lastScraped[name] = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = si.Batch(batch)
	if err != nil {
		log.Fatal().Err(err)
	}

	indexDuration := time.Since(startTime)
	indexDurationSeconds := float64(indexDuration) / float64(time.Second)
	timePerDoc := float64(indexDuration) / float64(count)
	log.Info().Msgf("Indexed %d documents, in %.2fs (average %.2fms/doc)", count, indexDurationSeconds, timePerDoc/float64(time.Millisecond))
	return nil
}

// Query runs the actual search query against the index.
// It uses the boolean query which is a type of query builder
// The search phrase given is separated into separate search terms, made into a wildcard query
// and then passed to the boolean query. The boolean query checks that all terms are found in any hits
// it returns.
// Example: "hello world" is searched as .*hello.* .*world.* and only when both are present in a document
// that document will be present in the results
func (si *Search) Query(searchPhrase string) ([]string, error) {
	queryBuilder := bleve.NewBooleanQuery()

	for _, term := range strings.Split(searchPhrase, " ") {
		query := bleve.NewWildcardQuery(fmt.Sprintf(searchSyntax, term))
		queryBuilder.AddMust(query)
	}

	searchRequest := bleve.NewSearchRequest(queryBuilder)
	searchResult, err := si.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	var matchingIDs []string
	for _, result := range searchResult.Hits {
		matchingIDs = append(matchingIDs, result.ID)
	}

	return matchingIDs, nil
}
