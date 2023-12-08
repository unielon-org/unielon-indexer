package storage

import (
	"database/sql"
	"errors"
	"github.com/dogecoinw/go-dogecoin/log"
	_ "github.com/mattn/go-sqlite3"
	"github.com/unielon-org/unielon-indexer/utils"
	"math/big"
	"sync"
)

var (
	ErrNotFound = errors.New("not found")
)

type DBClient struct {
	SqlDB *sql.DB
	lock  *sync.RWMutex
}

func NewSqliteClient(cfg utils.SqliteConfig) *DBClient {
	db, err := sql.Open("sqlite3", cfg.Database)
	if err != nil {
		log.Error("NewMysqlClient", "err", err)
		return nil
	}

	lock := new(sync.RWMutex)
	conn := &DBClient{
		SqlDB: db,
		lock:  lock,
	}

	return conn
}

func (conn *DBClient) Stop() {
	conn.SqlDB.Close()
}

func (conn *DBClient) UpdateBlock(height int64, blockHash string) error {
	conn.lock.Lock()
	defer conn.lock.Unlock()

	stmt, err := conn.SqlDB.Prepare("INSERT OR REPLACE INTO block (block_hash, block_number) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(blockHash, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) InstallRevert(tx *sql.Tx, tick, from, to string, amt *big.Int, height int64) error {
	exec := "INSERT INTO cardinals_revert (tick, from_address, to_address, amt, block_number) VALUES (?, ?, ?, ?, ?)"
	_, err := tx.Exec(exec, tick, from, to, amt.String(), height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) DelRevert(tx *sql.Tx, height int64) error {
	query := "delete from cardinals_revert where block_number > ?"
	_, err := tx.Exec(query, height)
	if err != nil {
		return err
	}
	return nil
}

func (c *DBClient) FindRevertByNumber(height int64) ([]*utils.CardinalsRevert, error) {
	query := "SELECT tick, from_address, to_address, amt FROM cardinals_revert where block_number > ? order by id desc "
	rows, err := c.SqlDB.Query(query, height)
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	reverts := make([]*utils.CardinalsRevert, 0)
	for rows.Next() {
		revert := &utils.CardinalsRevert{}
		var amt string
		err := rows.Scan(&revert.Tick, &revert.FromAddress, &revert.ToAddress, &amt)
		if err != nil {
			return nil, err
		}

		revert.Amt, _ = utils.ConvetStr(amt)
		reverts = append(reverts, revert)
	}

	return reverts, nil
}

func (conn *DBClient) FindBlockByHeight(height int64) (string, error) {
	conn.lock.RLock()
	defer conn.lock.RUnlock()

	var blockHash string
	err := conn.SqlDB.QueryRow("SELECT block_hash FROM block WHERE block_number = ?", height).Scan(&blockHash)
	if err != nil {
		return "", err
	}
	return blockHash, nil
}

func (conn *DBClient) LastBlock() (int64, error) {
	conn.lock.RLock()
	defer conn.lock.RUnlock()

	var blockNumber int64
	err := conn.SqlDB.QueryRow("SELECT block_number FROM block ORDER BY block_number DESC LIMIT 1").Scan(&blockNumber)
	if err != nil {
		return 0, err
	}
	return blockNumber, nil
}
