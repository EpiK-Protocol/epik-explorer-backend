package storage

import (
	"fmt"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"

	"gorm.io/driver/postgres" //postgresql driver
	"gorm.io/gorm"
)

var (
	//DB mysql
	DB *gorm.DB
	//TipsetKV tokens
	TipsetKV *badger.DB
	//PowerKV vcode
	PowerKV *badger.DB
	//MessageKV vcode
	MessageKV *badger.DB
	//TestNetKV ...
	TestNetKV *badger.DB
	//WalletKV ...
	WalletKV *badger.DB
	//TokenKV tokens
	TokenKV *badger.DB
)

//RunMode debug
var RunMode = "debug"

//InitDatabase ...
func InitDatabase() {
	var err error
	source := fmt.Sprintf("sslmode=disable host=%s port=%d user=%s dbname=%s password='%s'",
		etc.Config.Postgres.Host,
		etc.Config.Postgres.Port,
		etc.Config.Postgres.User,
		etc.Config.Postgres.Database,
		etc.Config.Postgres.Password,
	)

	DB, err = gorm.Open(postgres.Open(source), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	d, err := DB.DB()
	if err != nil {
		panic(err)
	}
	err = d.Ping()
	if err != nil {
		panic(err)
	}
	opts := badger.DefaultOptions("").
		WithNumVersionsToKeep(1).
		WithSyncWrites(true).
		WithTruncate(true).
		WithValueLogLoadingMode(options.FileIO).
		WithNumMemtables(1).
		WithNumLevelZeroTables(1).
		WithNumLevelZeroTablesStall(2).
		WithTableLoadingMode(options.FileIO).
		WithMaxCacheSize(20 * 1024 * 1024).
		WithMaxTableSize(10 * 1024 * 1024).
		WithValueLogFileSize(20 * 1024 * 1024)

	dir := ""

	dir = etc.Config.BadgerDB.TipSet
	TipsetKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = etc.Config.BadgerDB.Power
	PowerKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = etc.Config.BadgerDB.Message
	MessageKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = etc.Config.BadgerDB.TestNet
	TestNetKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = etc.Config.BadgerDB.Wallet
	WalletKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = etc.Config.BadgerDB.Token
	TokenKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:

				TipsetKV.RunValueLogGC(0.1)
				PowerKV.RunValueLogGC(0.1)
				MessageKV.RunValueLogGC(0.1)
				TestNetKV.RunValueLogGC(0.1)
				WalletKV.RunValueLogGC(0.1)
				TokenKV.RunValueLogGC(0.1)
			}
		}
	}()
}

//CloseDatabase ...
func CloseDatabase() {
	// DB.Close()
	TipsetKV.Close()
	PowerKV.Close()
	MessageKV.Close()
	TestNetKV.Close()
	WalletKV.Close()
	TokenKV.Close()
}
