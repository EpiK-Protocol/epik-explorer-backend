package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	epikwallet "github.com/EpiK-Protocol/go-epik/chain/wallet"
	"github.com/filecoin-project/go-address"
	"github.com/gin-gonic/gin"
)

var epikWallet *epikwallet.Wallet

//Page ...
type Page struct {
	Size   int64 `json:"size"`
	Offset int64 `json:"offset"`
}

//ParsePage ...
func ParsePage(c *gin.Context) (page *Page) {
	page = &Page{}
	_page, _ := strconv.ParseInt(c.Query("page"), 10, 64)
	page.Size, _ = strconv.ParseInt(c.Query("size"), 10, 64)
	page.Offset, _ = strconv.ParseInt(c.Query("offset"), 10, 64)
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
