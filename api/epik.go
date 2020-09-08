package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/filecoin-project/go-address"
	"github.com/ipfs/go-cid"

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
		types.BlockHeader
		Cid string
	}
	resultList := []*blockResult{}
	for _, block := range list {
		result := &blockResult{}
		result.BlockHeader = *block
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
	list := epik.GetLatestMessages(req.Num)
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
		fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		head, err := fullAPI.ChainHead(context.Background())
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		addr, err := address.NewFromString(word)
		if err != nil {
			responseJSON(c, clientError(fmt.Errorf("address invalid")))
			return
		}
		from, err := fullAPI.StateListMessages(context.Background(), &types.Message{From: addr}, head.Key(), head.Height()-1000)
		to, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: addr}, head.Key(), head.Height()-1000)
		cids := append(from, to...)
		if err != nil {
			return
		}
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		type resultMessage struct {
			types.Message
			Cid string
		}
		msgs := []*resultMessage{}
		for _, cid := range cids {
			msg, err := fullAPI.ChainGetMessage(context.Background(), cid)
			if err != nil {
				responseJSON(c, errServerError)
				return
			}
			resultMsg := &resultMessage{}
			resultMsg.Message = *msg
			resultMsg.Cid = msg.Cid().String()
			msgs = append(msgs, resultMsg)
		}
		responseJSON(c, errOK, "list", msgs)
		return
	case "message":
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
		message, err := fullAPI.ChainGetMessage(context.Background(), cid)
		if err != nil {
			responseJSON(c, errServerError)
			return
		}
		responseJSON(c, errOK, "message", message)
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
