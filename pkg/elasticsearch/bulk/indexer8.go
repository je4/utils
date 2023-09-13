package bulk

import (
	"bytes"
	"context"
	"emperror.dev/errors"
	"encoding/json"
	"fmt"
	"github.com/cenkalti/backoff/v3"
	"github.com/dustin/go-humanize"
	"github.com/elastic/elastic-transport-go/v8/elastictransport"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/rs/zerolog"
	"go.elastic.co/apm/module/apmelasticsearch"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Indexer8 struct {
	index           string
	es              *elasticsearch.Client
	log             *zerolog.Logger
	start           time.Time
	bi              esutil.BulkIndexer
	countSuccessful uint64
	countError      uint64
}

func NewIndexer8(address string, index string, apikey string, logger *zerolog.Logger) (Indexer, error) {
	var err error
	idx := &Indexer8{
		index: index,
		es:    nil,
		log:   logger,
	}
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	//
	// Use a third-party package for implementing the backoff function
	//
	retryBackoff := backoff.NewExponentialBackOff()

	cfg := elasticsearch.Config{
		APIKey: apikey,
		Addresses: []string{
			address,
		},
		// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		// ... using the "apmelasticsearch" wrapper for instrumentation
		Transport: apmelasticsearch.WrapRoundTripper(http.DefaultTransport),
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

		// Retry on 429 TooManyRequests statuses
		//
		RetryOnStatus: []int{502, 503, 504, 429},
		// Configure the backoff function
		//
		RetryBackoff: func(i int) time.Duration {
			if i == 1 {
				retryBackoff.Reset()
			}
			return retryBackoff.NextBackOff()
		},

		// Retry up to 5 attempts
		//
		MaxRetries: 5,

		Logger: &elastictransport.ColorLogger{Output: os.Stdout},
	}

	idx.es, err = elasticsearch.NewClient(cfg)
	if err != nil {
		logger.Fatal().Err(err)
	}
	return idx, nil
}

func (idx *Indexer8) Info() (clientVersion string, serverVersion string, err error) {
	// 1. Get cluster info
	//
	var r map[string]interface{}
	res, err := idx.es.Info()
	if err != nil {
		return "", "", errors.Wrapf(err, "error getting info  response")
	}
	defer res.Body.Close()
	// Check response status
	if res.IsError() {
		return "", "", errors.Errorf("info response has error %s", res.String())
	}
	// Deserialize the response into a map.
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return "", "", errors.Wrap(err, "cannot pars info result body")
	}
	// Print client and server version numbers.
	clientVersion = elasticsearch.Version
	var ok bool
	serverVersion, ok = r["version"].(map[string]interface{})["number"].(string)
	if !ok {
		return "", "", errors.Errorf("invalid info structure %v", r)
	}
	return
}

func (idx *Indexer8) CreateIndex(schema []byte) error {
	// Re-create the index
	//
	res, err := idx.es.Indices.Delete([]string{idx.index}, idx.es.Indices.Delete.WithIgnoreUnavailable(true))
	if err != nil {
		return errors.Wrapf(err, "Cannot delete index %s", idx.index)
	}
	if res.IsError() {
		return errors.Wrapf(err, "Cannot create index %s - %s", idx.index, res.String())
	}
	res.Body.Close()

	res, err = idx.es.Indices.Create(idx.index, idx.es.Indices.Create.WithBody(bytes.NewReader(schema)))
	if err != nil {
		return errors.Wrapf(err, "Cannot create index %s", idx.index)
	}
	if res.IsError() {
		errMsg, err := io.ReadAll(res.Body)
		res.Body.Close()
		return errors.Wrapf(err, "Cannot create index %s - %s // %s", idx.index, res.String(), string(errMsg))
	}
	res.Body.Close()

	/*
		buf := bytes.NewBuffer([]byte("{\n  \"index.mapping.total_fields.limit\": 2000\n}"))
		res, err = idx.es.Indices.PutSettings(buf)
		if err != nil {
			return errors.Wrapf(err, "Cannot put settings index %s", idx.index)
		}
		res.Body.Close()
		if res.IsError() {
			return errors.Wrapf(err, "Cannot create index %s - %s", idx.index, res.String())
		}
	*/
	return nil
}

func (idx *Indexer8) StartBulk(workers int, flushbytes int, flushtime time.Duration) error {
	var err error
	idx.start = time.Now()
	idx.bi, err = esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         idx.index,  // The default index name
		Client:        idx.es,     // The Elasticsearch client
		NumWorkers:    workers,    // The number of worker goroutines
		FlushBytes:    flushbytes, // The flush threshold in bytes
		FlushInterval: flushtime,  // The periodic flush interval
		//DebugLogger:   &esLogger{logger},
	})
	if err != nil {
		return errors.Wrap(err, "cannot initialize bulk indexer")
	}
	return nil
}
func (idx *Indexer8) CloseBulk() error {
	if idx.bi == nil {
		return errors.New("bulk indexer not initialized")
	}
	// >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
	// Close the indexer
	//
	if err := idx.bi.Close(context.Background()); err != nil {
		return errors.Wrap(err, "cannot close bulk indexer")
	}
	// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<

	biStats := idx.bi.Stats()

	// Report the results: number of indexed docs, number of errors, duration, indexing rate
	//
	idx.log.Info().Msg(strings.Repeat("▔", 65))

	dur := time.Since(idx.start)

	if biStats.NumFailed > 0 {
		msg := fmt.Sprintf(
			"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			humanize.Comma(int64(biStats.NumFailed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed+biStats.NumFailed))),
		)
		idx.log.Info().Msg(msg)
		return errors.Errorf(
			msg,
		)
	} else {
		idx.log.Info().Msgf(
			"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
			humanize.Comma(int64(biStats.NumFlushed)),
			dur.Truncate(time.Millisecond),
			humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
		)
	}
	return nil
}
func (idx *Indexer8) Delete(id string) error {
	if err := idx.bi.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			// Action field configures the operation to perform (index, create, delete, update)
			Action: "delete",

			Index: idx.index,

			// DocumentID is the (optional) document ID
			DocumentID: id,

			// OnSuccess is called for each successful operation
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				atomic.AddUint64(&idx.countSuccessful, 1)
				idx.log.Info().Msgf("%s: %s", item.Action, item.DocumentID)
			},

			// OnFailure is called for each failed operation
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				atomic.AddUint64(&idx.countError, 1)
				if err != nil {
					idx.log.Error().Msgf("%s", err)
				} else {
					if res.Error.Type != "" {
						idx.log.Error().Msgf("%s: %s", res.Error.Type, res.Error.Reason)
					} else {
						idx.log.Error().Msgf("[%v] %s - %s - %s", res.Status, item.Action, item.DocumentID, res.Result)
					}
				}
			},
		},
	); err != nil {
		return errors.Wrapf(err, "cannot delete document %s", id)
	}
	return nil
}

func (idx *Indexer8) Index(id string, doc any) error {
	if doc == nil {
		return errors.Errorf("document %s is nil", id)
	}
	jsBuf := bytes.NewBuffer(nil)
	enc := json.NewEncoder(jsBuf)
	//enc.SetIndent("", "  ")
	err := enc.Encode(doc)
	if err != nil {
		return errors.Wrap(err, "cannot json encode document")
	}
	jsonBytes := jsBuf.Bytes()
	//logger.Print(string(jsonBytes))
	if err := idx.bi.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			// Action field configures the operation to perform (index, create, delete, update)
			Action: "index",

			Index: idx.index,

			// DocumentID is the (optional) document ID
			DocumentID: id,

			// Body is an `io.Reader` with the payload
			Body: bytes.NewReader(jsonBytes),

			// OnSuccess is called for each successful operation
			OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
				atomic.AddUint64(&idx.countSuccessful, 1)
			},

			// OnFailure is called for each failed operation
			OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
				atomic.AddUint64(&idx.countError, 1)
				if err != nil {
					idx.log.Error().Msgf("ERROR: %s", err)
				} else {
					idx.log.Error().Msgf("ERROR [%s] '%s' caused by [%s] '%s' in document '%s'", res.Error.Type, res.Error.Reason, res.Error.Cause.Type, res.Error.Cause.Reason, item.DocumentID)
				}
				item.Body.Seek(0, io.SeekStart)
				body, _ := io.ReadAll(item.Body)
				idx.log.Debug().Msgf("%s", string(body))
			},
		},
	); err != nil {
		return errors.Wrapf(err, "cannot index document %s", id)
	}
	return nil
}
