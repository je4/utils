package badgerBuffer

import (
	"emperror.dev/errors"
	"github.com/dgraph-io/badger/v4"
	"sync"
	"sync/atomic"
)

func NewBadgerBuffer(size int, db *badger.DB, cond *sync.Cond, sharedCond *atomic.Bool) (*BadgerBuffer, error) {
	return &BadgerBuffer{
		vals:       make([]*kv, 0, size),
		size:       size,
		db:         db,
		Mutex:      sync.Mutex{},
		cond:       cond,
		sharedCond: sharedCond,
	}, nil
}

type kv struct {
	key, value []byte
}

type BadgerBuffer struct {
	sync.Mutex
	vals       []*kv
	size       int
	db         *badger.DB
	cond       *sync.Cond
	sharedCond *atomic.Bool
}

func (bb *BadgerBuffer) flush() error {
	if bb.sharedCond != nil && bb.cond != nil {
		bb.cond.L.Lock()
		bb.sharedCond.Store(true)
		defer func() {
			bb.sharedCond.Store(false)
			bb.cond.Broadcast()
			bb.cond.L.Unlock()
		}()
	}
	err := bb.db.Update(func(txn *badger.Txn) error {
		for _, v := range bb.vals {
			if err := txn.Set(v.key, v.value); err != nil {
				return errors.Wrapf(err, "cannot store key '%s'", v.key)
			}
		}
		return nil
	})
	if err != nil {
		return errors.WithStack(err)
	}
	//clear(bb.vals)
	bb.vals = make([]*kv, 0, bb.size)
	return nil
}

func (bb *BadgerBuffer) Flush() error {
	bb.Lock()
	defer bb.Unlock()
	return errors.WithStack(bb.flush())
}

func (bb *BadgerBuffer) Add(key, val []byte) error {
	bb.Lock()
	defer bb.Unlock()
	bb.vals = append(bb.vals, &kv{key, val})
	if len(bb.vals) >= bb.size {
		return errors.WithStack(bb.flush())
	}
	return nil
}
