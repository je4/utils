package bulk

import "time"

type Indexer interface {
	Info() (clientVersion string, serverVersion string, err error)
	CreateIndex(schema []byte) error
	StartBulk(workers int, flushbytes int, flushtime time.Duration) error
	CloseBulk() error
	Delete(id string) error
	Index(id string, content any) error
}
