package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	epikwallet "github.com/EpiK-Protocol/go-epik/chain/wallet"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
)

var epikWallet *epikwallet.Wallet

//Page ...
type Page struct {
	Size   int `json:"size"`
	Offset int `json:"offset"`
}

//ParsePage ...
func ParsePage(c *gin.Context) (page *Page) {
	page = &Page{}
	_page := utils.ParseInt(c.Query("page"))
	page.Size = utils.ParseInt(c.Query("size"))
	page.Offset = utils.ParseInt(c.Query("offset"))
	if page.Size <= 0 {
		page.Size = 10
	}
	if _page > 0 && page.Offset <= 0 {
		page.Offset = _page * page.Size
	}
	return
}

//Start ...
func Start() {
	ks := epikwallet.NewMemKeyStore()
	epikWallet, _ = epikwallet.NewWallet(ks)
	keyInfo := &types.KeyInfo{}
	data, _ := hex.DecodeString(etc.Config.EPIK.MainPrivateKey)
	json.Unmarshal(data, keyInfo)
	epikWallet.Import(keyInfo)
	timerTask()
}

func timerTask() {
	ticker1H := time.NewTicker(time.Hour)
	ticker10M := time.NewTicker(time.Minute * 10)
	now := time.Now()

	//21点结算
	next21H := time.Date(now.Year(), now.Month(), now.Day(), 21, 5, 0, 0, now.Location())
	timer21H := time.NewTimer(next21H.Add(24 * time.Hour).Sub(now))
	if next21H.After(now) {
		timer21H.Reset(next21H.Sub(now))
	}

	//10点打币
	// next22H := time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())
	// timer22H := time.NewTimer(next21H.Add(24 * time.Hour).Sub(now))
	// if next22H.After(now) {
	// 	timer22H.Reset(next22H.Sub(now))
	// }
	go func() {
		err := RefreshTestNet(storage.DB)
		if err != nil {
			fmt.Println(err)
		}
		err = RefreshCurrencyPrice()
		if err != nil {
			fmt.Println(err)
		}
		for {
			select {
			case <-ticker1H.C:
				err = RefreshTestNet(storage.DB)
				if err != nil {
					fmt.Println(err)
				}
			case <-ticker10M.C:
				err = RefreshCurrencyPrice()
				if err != nil {
					fmt.Println(err)
				}
			case <-timer21H.C:
				err = GenTestnetMinerBonus()
				if err == nil {
					next21H = time.Date(now.Year(), now.Month(), now.Day(), 21, 5, 0, 0, now.Location())
					timer21H.Reset(next21H.Add(24 * time.Hour).Sub(time.Now()))
				} else {
					fmt.Println(err)
					timer21H.Reset(5 * time.Minute)
				}
				// case <-timer22H.C:
				// 	err = PushMinerERC20Bonus()
				// 	if err == nil {
				// 		next22H = time.Date(now.Year(), now.Month(), now.Day(), 22, 0, 0, 0, now.Location())
				// 		timer22H.Reset(next21H.Add(24 * time.Hour).Sub(time.Now()))
				// 	} else {
				// 		fmt.Println(err)
				// 		timer22H.Reset(5 * time.Minute)
				// 	}
			}
		}
	}()

}

//SendEPK ...
func SendEPK(w *epikwallet.Wallet, to string, amount string) (cidStr string, err error) {
	fromAddr, err := w.GetDefault()
	toAddr, err := address.NewFromString(to)
	if err != nil {
		return
	}
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	if err != nil {
		return
	}
	gasPrice, err := fullAPI.MpoolEstimateGasPrice(context.Background(), 10, fromAddr, 10000, head.Key())
	if err != nil {
		return
	}
	nonce, err := fullAPI.MpoolGetNonce(context.Background(), fromAddr)
	if err != nil {
		return
	}
	epk, err := types.ParseEPK(amount)
	if err != nil {
		return
	}
	msg := types.Message{
		From:     fromAddr,
		To:       toAddr,
		Value:    types.BigInt(epk),
		GasPrice: gasPrice,
		GasLimit: 10000,
		Nonce:    nonce,
	}
	signature, err := w.Sign(context.Background(), fromAddr, msg.Cid().Bytes())
	if err != nil {
		return "", err
	}
	signedMsg := &types.SignedMessage{
		Message:   msg,
		Signature: *signature,
	}
	fmt.Println(signedMsg)
	c, err := fullAPI.MpoolPush(context.Background(), signedMsg)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}
