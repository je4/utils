package openai

import (
	"emperror.dev/errors"
	oai "github.com/sashabaranov/go-openai"
)

var ErrNotExists = errors.New("key does not exist")

type KVStore interface {
	Get(key string) (*oai.Embedding, error)
	Set(key string, value *oai.Embedding) error
}
