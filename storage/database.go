package storage

import (
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"
	//postgresql driver
)

var (
	//TipsetKV tokens
	TipsetKV *badger.DB
	//PowerKV vcode
	PowerKV *badger.DB
	//MessageKV vcode
	MessageKV *badger.DB
)

//RunMode debug
var RunMode = "debug"

//InitDatabase ...
func InitDatabase() {
	var err error
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

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				TipsetKV.RunValueLogGC(0.1)
				PowerKV.RunValueLogGC(0.1)
				MessageKV.RunValueLogGC(0.1)
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
}
