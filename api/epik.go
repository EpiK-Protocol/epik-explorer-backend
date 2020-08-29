package api

import (
	"epik-explorer-backend/epik"
	"fmt"
	"time"

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
	list := epik.GetLatestBlocks(req.Num)
	responseJSON(c, errOK, "block_header", list)
}

func latestMessage(c *gin.Context) {
	req := &struct {
		Num int `json:"num"`
	}{}
	parseRequestBody(c, req)
	list := epik.GetLatestMessages(req.Num)
	responseJSON(c, errOK, "msg", list)
}

func blocktimeGraphical(c *gin.Context) {
	req := &struct {
		StartTime int64 `json:"start_time"`
		EndTime   int64 `json:"end_time"`
	}{}
	parseRequestBody(c, req)

	history := epik.GetHistoryBlocks()
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
		start = history[0].Block.Timestamp
		end = history[len(history)-1].Block.Timestamp
	}
	var avg = float64(end-start) / float64(len(history))
	sector := uint64(1)
	block := 0
	fmt.Println(len(history))
	for _, hBlock := range history {
		// fmt.Println(hBlock.Block.Timestamp, now)
		if hBlock.Block.Timestamp < uint64(start+sector*600) {
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
			fmt.Println(bt)
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
