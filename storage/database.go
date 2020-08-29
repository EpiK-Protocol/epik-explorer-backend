package storage

import (
	"time"

	"epik-explorer-backend/utils"

	"github.com/dgraph-io/badger/v2"
	"github.com/dgraph-io/badger/v2/options"

	_ "github.com/jinzhu/gorm/dialects/postgres" //postgresql driver
)

var (
	//DB mysql
	// DB *gorm.DB
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
	// source := fmt.Sprintf("sslmode=disable host=%s port=%s user=%s dbname=%s password='%s'",
	// 	utils.ReadConfig("pg.host"),
	// 	utils.ReadConfig("pg.port"),
	// 	utils.ReadConfig("pg.user"),
	// 	utils.ReadConfig("pg.dbname"),
	// 	utils.ReadConfig("pg.password"),
	// )

	// DB, err = gorm.Open("postgres", source)
	// if err != nil {
	// 	panic(err)
	// }
	// DB.DB().SetMaxIdleConns(5)
	// DB.DB().SetMaxOpenConns(20)
	// err = DB.DB().Ping()
	// if err != nil {
	// 	panic(err)
	// }
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

	dir = utils.ReadConfig("db.tipset")
	TipsetKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = utils.ReadConfig("db.power")
	PowerKV, err = badger.Open(opts.WithDir(dir).WithValueDir(dir))
	if err != nil {
		panic(err)
	}
	dir = utils.ReadConfig("db.message")
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
