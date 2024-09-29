package storage

import (
	"fmt"
	"github.com/dogecoinw/go-dogecoin/rlp"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/unielon-org/unielon-indexer/utils"
	"sync"
)

type LevelDB struct {
	DB   *leveldb.DB
	lock *sync.RWMutex
}

func NewLevelDB(cfg utils.LevelDBConfig) *LevelDB {
	db, err := leveldb.OpenFile(cfg.Path, nil)
	if err != nil {
		panic(fmt.Sprintf("Leveldb err %s", err))
	}

	lock := new(sync.RWMutex)
	conn := &LevelDB{
		DB:   db,
		lock: lock,
	}

	return conn
}

func (conn *LevelDB) Stop() {
	conn.DB.Close()
}

func (conn *LevelDB) SetCacheOrderAddress(key string, value *utils.OrderAddressCache) error {
	conn.lock.Lock()
	defer conn.lock.Unlock()

	if data, err := rlp.EncodeToBytes(value); err != nil {
		return err
	} else {
		if err := conn.DB.Put([]byte("order-"+key), data, nil); err != nil {
			return err
		}
	}

	return nil
}

func (conn *LevelDB) GetCacheOrderAddress(key string) (*utils.OrderAddressCache, error) {
	conn.lock.RLock()
	defer conn.lock.RUnlock()

	var value *utils.OrderAddressCache
	if data, err := conn.DB.Get([]byte("order-"+key), nil); err != nil {
		return value, err
	} else {
		if err := rlp.DecodeBytes(data, &value); err != nil {
			return value, err
		}
	}
	return value, nil

}
