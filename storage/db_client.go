package storage

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/unielon-org/unielon-indexer/utils"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"sync"
	"time"
)

const (
	NETWORK          = "tcp"
	stakePoolAddress = "DS8eFcobjXp6oL8YoXoVazDQ32bcDdWwui"
)

type DBClient struct {
	DB   *gorm.DB
	lock *sync.RWMutex
}

func NewSqliteClient(cfg utils.SqliteConfig) *DBClient {

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// github.com/mattn/go-sqlite3
	db, err := gorm.Open(sqlite.Open(cfg.Database), &gorm.Config{Logger: newLogger})
	if err != nil {
		fmt.Printf("Open failed,err:%v  ", err)
		os.Exit(0)
	}

	_ = db.Exec("PRAGMA journal_mode=WAL;")

	sqlDB, dbError := db.DB()
	if dbError != nil {
		fmt.Printf("get db failed,err:%v  ", dbError)
		os.Exit(0)
	}

	sqlDB.SetMaxIdleConns(10)

	sqlDB.SetMaxOpenConns(100)

	lock := new(sync.RWMutex)
	conn := &DBClient{
		DB:   db,
		lock: lock,
	}

	return conn
}

func NewMysqlClient(cfg utils.MysqlConfig) *DBClient {

	dsn := fmt.Sprintf("%s:%s@%s(%s:%d)/%s?parseTime=true", cfg.UserName, cfg.PassWord, NETWORK, cfg.Server, cfg.Port, cfg.Database)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second, //
			LogLevel:                  logger.Warn, //
			IgnoreRecordNotFoundError: true,        //
			Colorful:                  true,        //
		},
	)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		fmt.Printf("Open failed,err:%v  ", err)
		os.Exit(0)
	}

	lock := new(sync.RWMutex)
	conn := &DBClient{
		DB:   db,
		lock: lock,
	}

	return conn
}

func (conn *DBClient) Stop() {
	sqlDB, err := conn.DB.DB()
	if err != nil {
		fmt.Printf("get db failed,err:%v  ", err)
		return
	}
	sqlDB.Close()
}
