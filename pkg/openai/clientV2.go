package openai

import (
	"context"
	"crypto/sha1"
	"emperror.dev/errors"
	"fmt"
	"github.com/je4/utils/v2/pkg/zLogger"
	oai "github.com/sashabaranov/go-openai"
)

func NewClientV2(apiKey string, kv KVStore, logger zLogger.ZLogger) *ClientV2 {
	return &ClientV2{
		client: oai.NewClient(apiKey),
		apiKey: apiKey,
		kv:     kv,
		logger: logger,
	}
}

type ClientV2 struct {
	client *oai.Client
	apiKey string
	kv     KVStore
	logger zLogger.ZLogger
}

func (c *ClientV2) CreateEmbedding(input string, model oai.EmbeddingModel) (*oai.Embedding, error) {
	var key = []byte(fmt.Sprintf("embedding-%x", sha1.Sum([]byte(input+string(model)))))
	var result *oai.Embedding
	result, err := c.kv.Get(string(key))
	if err != nil {
		if errors.Is(err, ErrNotExists) {
			c.logger.Info().Msgf("cache miss value for key %s", string(key))
		} else {
			return nil, err
		}
	} else {
		c.logger.Info().Msgf("cache hit value for key %s", string(key))
		return result, nil
	}

	if result != nil {
		return result, nil
	}

	// Create an EmbeddingRequest for the user query
	queryReq := oai.EmbeddingRequest{
		Input: []string{input},
		Model: model,
	}
	queryResponse, err := c.client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create embedding")
	}
	if len(queryResponse.Data) == 0 {
		return nil, errors.Errorf("no embedding returned")
	}
	result = &queryResponse.Data[0]
	if err := c.kv.Set(string(key), result); err != nil {
		return nil, errors.Wrapf(err, "cannot set value for key %s", string(key))
	}

	return result, nil
}
