package openai

import (
	"emperror.dev/errors"
	"github.com/bluele/gcache"
	oai "github.com/sashabaranov/go-openai"
)

func NewKVGCache(cache gcache.Cache) *KVGCache {
	return &KVGCache{
		cache: cache,
	}
}

type KVGCache struct {
	cache gcache.Cache
}

func (k *KVGCache) Get(key string) (*oai.Embedding, error) {
	e, err := k.cache.Get(key)
	if err != nil {
		if errors.Is(err, gcache.KeyNotFoundError) {
			return nil, ErrNotExists
		}
		return nil, errors.WithStack(err)
	}
	embedding, ok := e.(*oai.Embedding)
	if !ok {
		return nil, errors.New("cannot cast to *oai.Embedding")
	}
	return embedding, nil
}

func (k *KVGCache) Set(key string, value *oai.Embedding) error {
	return errors.WithStack(k.cache.Set(key, value))
}
