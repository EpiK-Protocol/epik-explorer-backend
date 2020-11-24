package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/dgraph-io/badger/v2"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"
	"github.com/shopspring/decimal"

	"github.com/gin-gonic/gin"
)

//SetEpikExplorerAPI ...
func SetEpikExplorerAPI(e *gin.Engine) {
	e.POST("BaseInformation", baseInformation)
	e.POST("LatestBlock", latestBlock)
	e.POST("LatestMsg", latestMessage)
	e.POST("BlocktimeGraphical", blocktimeGraphical)
	e.POST("tipset/TipSetTree", tipSetTree)
	e.POST("peer/ActivePeerCount", activePeerCount)
	e.POST("TotalPowerGraphical", totalPowerGraphical)
	e.POST("AvgBlockheaderSizeGraphical", avgBlockheaderSizeGraphical)
	e.GET("Search", epikSearch)
	e.GET("MinerInfo", minerInfo)
	e.GET("MinerStatus", minerStatus)
}

func peerMap(c *gin.Context) {

}

func baseInformation(c *gin.Context) {
	responseJSON(c, errOK, "data", epik.CurrentBaseInfo)
}

func totalPowerGraphical(c *gin.Context) {
	list := epik.GetPowerList()
	responseJSON(c, errOK, "data", list)
}

func getBlockWonList(c *gin.Context) {

}

func minerPowerAtTime(c *gin.Context) {

}

func avgBlockheaderSizeGraphical(c *gin.Context) {
	type headSize struct {
		Time      int64 `json:"time"`
		BlockSize int64 `json:"block_size"`
	}
	list := []*headSize{}
	now := time.Now().Unix()
	for now > time.Now().Add(-24*time.Hour).Unix() {
		hs := &headSize{
			Time:      now,
			BlockSize: 673,
		}
		list = append(list, hs)
		now -= 3600
	}
	responseJSON(c, errOK, "min", 673, "max", 673, "avg_blocksize", 673, "data", list)
}

func latestBlock(c *gin.Context) {
	req := &struct {
		Num int `json:"num"`
	}{}
	parseRequestBody(c, req)
	list := epik.GetLatestBlocks(req.Num, false)
	type blockResult struct {
		epik.EBlockHeader
		Cid string
	}
	resultList := []*blockResult{}
	for _, block := range list {
		result := &blockResult{}
		result.EBlockHeader = *block
		result.Cid = block.Cid().String()
		resultList = append(resultList, result)
	}
	responseJSON(c, errOK, "block_header", resultList)
}

func latestMessage(c *gin.Context) {
	req := &struct {
		Num int `json:"num"`
	}{}
	parseRequestBody(c, req)
	list, _ := epik.GetLatestMessage()
	type msgResult struct {
		types.Message
		Cid string
	}
	resultList := []*msgResult{}
	for _, msg := range list {
		result := &msgResult{}
		result.Message = *msg
		result.Cid = msg.Cid().String()
		resultList = append(resultList, result)
	}
	responseJSON(c, errOK, "msg", resultList)
}

func blocktimeGraphical(c *gin.Context) {
	req := &struct {
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
	}{}
	parseRequestBody(c, req)

	history := epik.GetLatestBlocks(500, true)
	// fmt.Println(history)
	min := float64(100)
	max := float64(0)
	type blockTime struct {
		Time      int64   `json:"time"`
		BlockTime float64 `json:"block_time"`
	}
	list := []blockTime{}
	start := uint64(0)
	end := uint64(0)
	if len(history) > 0 {
		start = history[0].Timestamp
		end = history[len(history)-1].Timestamp
	}
	fmt.Println(start, end)
	var avg = float64(end-start) / float64(len(history))
	sector := uint64(1)
	block := 0
	// fmt.Println(len(history))
	for _, hBlock := range history {
		// fmt.Println(hBlock.Block.Timestamp, now)
		if hBlock.Timestamp < uint64(start+sector*600) {
			block++
		} else {

			bt := blockTime{
				Time:      int64(start),
				BlockTime: 0,
			}
			if block > 0 {
				bt.BlockTime = float64(600) / float64(block)
				if min > bt.BlockTime {
					min = bt.BlockTime
				}
				if max < bt.BlockTime {
					max = bt.BlockTime
				}
			}
			// fmt.Println(bt)
			list = append(list, bt)
			sector++
			start += 600
			block = 0
		}
	}
	responseJSON(c, errOK, "min", min, "max", max, "avg", avg, "data", list)
}

func tipSetTree(c *gin.Context) {
	list := epik.GetLatestTipSets(15)
	responseJSON(c, errOK, "tipsets", list)
}

func activePeerCount(c *gin.Context) {
	responseJSON(c, errOK, "count", 1)
}

func epikSearch(c *gin.Context) {
	_type := c.Query("type")
	word := c.Query("word")
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))

	switch _type {
	case "address":

		type resultMessage struct {
			types.Message
			Cid string
		}
		msgs := []*resultMessage{}
		messages, err := epik.ReadAddrMessage(word, time.Now(), 100)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		for _, message := range messages {
			resultMsg := &resultMessage{}
			resultMsg.Message = *message.Message
			resultMsg.Cid = message.Message.Cid().String()
			msgs = append(msgs, resultMsg)
		}
		responseJSON(c, errOK, "list", msgs)
		return
	case "message":
		msg, err := epik.GetMessage(word)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		responseJSON(c, errOK, "message", msg.Message)
		return
	case "block":
		fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		cid, err := cid.Decode(word)
		if err != nil {
			responseJSON(c, errClientError)
			return
		}
		block, err := fullAPI.ChainGetBlock(context.Background(), cid)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		responseJSON(c, errOK, "block", block)
		return
	}
}

func minerInfo(c *gin.Context) {
	id := c.Query("miner")
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	addr, err := address.NewFromString(id)
	if err != nil {
		responseJSON(c, errClientError)
		return
	}
	miner, err := fullAPI.StateMinerInfo(context.Background(), addr, types.EmptyTSK)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	balance, err := fullAPI.StateMinerAvailableBalance(context.Background(), addr, types.EmptyTSK)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	bal, _ := decimal.NewFromString(balance.String())
	dec18, _ := decimal.NewFromString("1000000000000000000")
	bal = bal.Div(dec18)
	power, err := fullAPI.StateMinerPower(context.Background(), addr, types.EmptyTSK)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}

	responseJSON(c, errOK, "miner_id", addr.String(), "sector_size", miner.SectorSize, "peer_id", miner.PeerId, "balance", bal.String(), "miner_power", power.MinerPower, "total_power", power.TotalPower)
}

func minerStatus(c *gin.Context) {
	type minerStatus struct {
		Total   int64
		Pledged int64
		Won     int64
	}
	statuses := []*minerStatus{}
	storage.PowerKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("MINERSTATUS:"))
		valid := prefix
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			data, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			status := &minerStatus{}
			err = json.Unmarshal(data, status)
			if err == nil {
				statuses = append([]*minerStatus{status}, statuses...)
			}
			if len(statuses) > 100 {
				break
			}
		}
		return
	})

	responseJSON(c, errOK, "list", statuses)
}
