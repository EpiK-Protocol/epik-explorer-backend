package api

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime/debug"

	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/dgraph-io/badger/v2"
	"github.com/gin-gonic/gin"
)

//SetWalletAPI ...
func SetWalletAPI(e *gin.Engine) {
	e.GET("wallet/price", walletPrice)
	e.GET("wallet/config")
}

func walletPrice(c *gin.Context) {

	prices := []*struct {
		ID     string `json:"id"`
		Price  string `json:"price"`
		Change string `json:"change"`
	}{
		{ID: "BTC"},
		{ID: "ETH"},
	}
	storage.WalletKV.View(func(txn *badger.Txn) (err error) {
		for _, price := range prices {
			item, err := txn.Get([]byte(fmt.Sprintf("CURRENCYPRICE:%s", price.ID)))
			if err != nil {
				continue
			}
			item.Value(func(value []byte) error {
				price.Price = string(value)
				return nil
			})
			item, err = txn.Get([]byte(fmt.Sprintf("CURRENCYCHANGE:%s", price.ID)))
			if err != nil {
				continue
			}
			item.Value(func(value []byte) error {
				price.Change = string(value)
				return nil
			})
		}
		return nil
	})
	responseJSON(c, errOK, "prices", prices)
}

//RefreshCurrencyPrice ..
func RefreshCurrencyPrice() (err error) {
	fmt.Println("refresh wallet price")
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get("https://api.nomics.com/v1/currencies/ticker?key=f6dd9ae39a202ae7e5496155be90998d&ids=BTC,ETH,USDT&interval=1d,30d&convert=USD")
	if err != nil {
		return err
	}
	data, _ := ioutil.ReadAll(resp.Body)
	result := []*struct {
		ID    string `json:"id"`
		Price string `json:"price"`
		D1    struct {
			PriceChangePct string `json:"price_change_pct"`
		} `json:"1d"`
	}{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return err
	}
	storage.WalletKV.Update(func(txn *badger.Txn) (err error) {
		for _, currency := range result {
			txn.Set([]byte(fmt.Sprintf("CURRENCYPRICE:%s", currency.ID)), []byte(currency.Price))
			txn.Set([]byte(fmt.Sprintf("CURRENCYCHANGE:%s", currency.ID)), []byte(currency.D1.PriceChangePct))

		}
		return nil
	})
	return
}
