package openai

import (
	"bytes"
	"emperror.dev/errors"
	"encoding/json"
	"github.com/andybalholm/brotli"
	"github.com/dgraph-io/badger/v4"
	oai "github.com/sashabaranov/go-openai"
	"io"
)

func NewKVBadger(db *badger.DB) *KVBadger {
	return &KVBadger{
		badger: db,
	}
}

type KVBadger struct {
	badger *badger.DB
}

func (k *KVBadger) Get(key string) (*oai.Embedding, error) {
	var result *oai.Embedding
	err := k.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return ErrNotExists
			}
			return errors.Wrapf(err, "cannot get item for key %s", string(key))
		}
		return item.Value(func(val []byte) error {
			br := brotli.NewReader(bytes.NewReader(val))
			jsonBytes, err := io.ReadAll(br)
			if err != nil {
				return errors.Wrapf(err, "cannot read value from item for key %s", string(key))
			}
			result = &oai.Embedding{}
			if err := json.Unmarshal(jsonBytes, &result); err != nil {
				return errors.Wrapf(err, "cannot unmarshal json for key %s", string(key))
			}
			return nil
		})
	})
	return result, errors.WithStack(err)
}

func (k *KVBadger) Set(key string, value *oai.Embedding) error {
	return errors.WithStack(k.badger.Update(func(txn *badger.Txn) error {
		jsonStr, err := json.Marshal(value)
		if err != nil {
			return errors.Wrapf(err, "cannot marshal json for key %s", string(key))
		}
		buf := &bytes.Buffer{}
		wr := brotli.NewWriter(buf)
		if _, err := wr.Write(jsonStr); err != nil {
			return errors.Wrapf(err, "cannot write value for key %s", string(key))
		}
		if err := wr.Close(); err != nil {
			return errors.Wrapf(err, "cannot close writer for key %s", string(key))
		}
		return txn.Set([]byte(key), buf.Bytes())
	}))
}

var _ KVStore = (*KVBadger)(nil)
