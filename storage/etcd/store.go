package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"github.com/laincloud/deployd/storage"

	client "github.com/coreos/etcd/clientv3"
)

type EtcdStore struct {
	keysApi *client.Client
	ctx     context.Context

	sync.RWMutex
	keyHashes map[string]uint64
}

func (store *EtcdStore) Get(key string, v interface{}) error {
	if resp, err := store.keysApi.Get(store.ctx, key); err != nil {
		return err
	} else {
		if len(resp.Kvs) == 0 {
			return storage.ErrNoSuchKey
		}
		if len(resp.Kvs) > 1 {
			return fmt.Errorf("Etcd Store returns this is a directory node")
		}

		value := resp.Kvs[0].Value
		if err := json.Unmarshal([]byte(value), v); err != nil {
			return err
		}
	}
	return nil
}

func (store *EtcdStore) KeysByPrefix(prefix string) ([]string, error) {
	// Prefix should corresponding to a directory name, and will return all the nodes inside the directory
	keys := make([]string, 0)
	if resp, err := store.keysApi.Get(store.ctx, prefix, client.WithPrefix()); err != nil {
		return keys, err
	} else {
		if len(resp.Kvs) == 0 {
			return keys, storage.ErrNoSuchKey
		}

		for _, ev := range resp.Kvs {
			if string(ev.Key) == prefix {
				return []string{}, fmt.Errorf("Etcd store returns a non-directory node")
			}
			keys = append(keys, string(ev.Key))
		}
	}
	return keys, nil
}

func (store *EtcdStore) Set(key string, v interface{}, force ...bool) error {
	if data, err := json.Marshal(v); err != nil {
		return err
	} else {
		h := fnv.New64a()
		h.Write(data)
		dataHash := h.Sum64()
		forceSave := false
		if len(force) > 0 {
			forceSave = force[0]
		}

		store.Lock()
		defer store.Unlock()
		if !forceSave {
			if lastHash, ok := store.keyHashes[key]; ok && lastHash == dataHash {
				return nil
			}
		}
		_, err := store.keysApi.Put(store.ctx, key, string(data))
		if err == nil {
			store.keyHashes[key] = dataHash
		}
		return err
	}
}

func (store *EtcdStore) Remove(key string) error {
	_, err := store.keysApi.Delete(store.ctx, key)
	if err != nil {
		store.Lock()
		delete(store.keyHashes, key)
		store.Unlock()
	}
	return err
}

func (store *EtcdStore) RemoveDir(key string) error {
	err := store.deleteDir(key, true)
	return err
}

func (store *EtcdStore) TryRemoveDir(key string) {
	store.deleteDir(key, false)
}

func (store *EtcdStore) deleteDir(key string, recursive bool) error {
	if !recursive {
		keys, _ := store.KeysByPrefix(key)
		if len(keys) > 0 {
			return fmt.Errorf("Etcd exist subKeys")
		}
	}

	_, err := store.keysApi.Delete(store.ctx, key, client.WithPrefix())
	return err
}

func NewStore(addr string, isDebug bool) (storage.Store, error) {
	c, err := client.New(client.Config{
		Endpoints: strings.Split(addr, ","),
	})
	if err != nil {
		return nil, err
	}

	s := &EtcdStore{
		keysApi:   c,
		ctx:       context.Background(),
		keyHashes: make(map[string]uint64),
	}
	return s, nil
}
