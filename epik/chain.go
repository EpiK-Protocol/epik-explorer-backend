package epik

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"

	"github.com/EpiK-Protocol/go-epik/api"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/dgraph-io/badger/v2"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
)

//CurrentBaseInfo ...
var CurrentBaseInfo *EBaseInformation

func init() {
	CurrentBaseInfo = &EBaseInformation{}
	historyBlockList = []*HistoryBlock{}
	historyMessageList = []*HistoryMessage{}
	powerList = []*PowerGraph{}
}

//HistoryBlock ...
type HistoryBlock struct {
	Time  time.Time
	Block *types.BlockHeader
}

var historyBlockList []*HistoryBlock
var historyTipSetList []*types.TipSet

func updateBaseInfoWithNewTipSet(tipset *types.TipSet) {
	CurrentBaseInfo.TipsetHeight = utils.ParseInt64(tipset.Height().String())
	CurrentBaseInfo.HeadUpdate = time.Now().Unix()
	now := time.Now()
	historyTipSetList = append(historyTipSetList, tipset)
	for _, block := range tipset.Blocks() {
		hb := &HistoryBlock{
			Time:  now,
			Block: block,
		}
		hb.Block.Timestamp = uint64(now.Unix())
		historyBlockList = append(historyBlockList, hb)
	}
	has := false
	for i, b := range historyBlockList {
		if b.Time.Add(24 * time.Hour).Before(now) {
			has = true
		}
		if b.Time.Add(24 * time.Hour).After(now) {
			if has {
				historyBlockList = historyBlockList[i:]
			}
			break
		}
	}
	if len(historyTipSetList) > 36000 {
		historyTipSetList = historyTipSetList[len(historyTipSetList)-36000:]
	}

}

func updateBaseInfoWithPledgeCollateral(pledge big.Int) {
	CurrentBaseInfo.PledgeCollateral = pledge.String()
}

//GetLatestBlocks ...
func GetLatestBlocks(num int) (list []*types.BlockHeader) {
	for i := len(historyBlockList) - 1; i > len(historyBlockList)-num-1 && i >= 0; i-- {
		block := historyBlockList[i]
		list = append(list, block.Block)
	}
	return
}

//GetHistoryBlocks ...
func GetHistoryBlocks() (list []*HistoryBlock) {
	return historyBlockList
}

//GetLatestTipSets ...
func GetLatestTipSets(length int) (list []*types.TipSet) {
	start := len(historyTipSetList) - length
	if start < 0 {
		start = 0
	}
	list = historyTipSetList[start:]
	return
}

//HistoryMessage ...
type HistoryMessage struct {
	Time          time.Time
	BlockMessages *api.BlockMessages
}

var historyMessageList []*HistoryMessage

func updateBaseInfoWithBlockMessages(bmsgs *api.BlockMessages) {
	now := time.Now()
	hm := &HistoryMessage{
		Time:          now,
		BlockMessages: bmsgs,
	}
	historyMessageList = append(historyMessageList, hm)
	has := false
	for i, msg := range historyMessageList {
		if msg.Time.Add(48 * time.Hour).Before(now) {
			has = true
		}
		if msg.Time.Add(48 * time.Hour).After(now) {
			if has {
				historyMessageList = historyMessageList[i:]
			}
			break
		}
	}

	//caculate avg messages
	var sumMessage, sumBlock int
	sumGasPrice := big.NewInt(0)
	sumBlock = len(historyMessageList)
	for _, msg := range historyMessageList {
		sumMessage += len(msg.BlockMessages.SecpkMessages)
		for _, m := range msg.BlockMessages.SecpkMessages {
			sumGasPrice = types.BigAdd(sumGasPrice, m.Message.GasPrice)
		}
		sumMessage += len(msg.BlockMessages.SecpkMessages)
		for _, m := range msg.BlockMessages.BlsMessages {
			sumGasPrice = types.BigAdd(sumGasPrice, m.GasPrice)
		}
	}
	if sumBlock > 0 {
		CurrentBaseInfo.AvgMessagesTipset = float64(sumMessage) / float64(sumBlock)
	}
	if sumMessage > 0 {
		CurrentBaseInfo.AvgGasPrice = utils.ParseFloat64(types.BigDiv(sumGasPrice, big.NewInt(int64(sumMessage))).String())
	}
}

//GetLatestMessages ...
func GetLatestMessages(num int) (list []*types.Message) {
	total := 0
	for i := len(historyMessageList) - 1; i > len(historyMessageList)-num-1 && i >= 0; i-- {
		msg := historyMessageList[i]

		for _, message := range msg.BlockMessages.BlsMessages {
			list = append(list, message)
			total++
			if total > num {
				break
			}
		}
		for _, message := range msg.BlockMessages.SecpkMessages {
			list = append(list, &message.Message)
			total++
			if total > num {
				break
			}
		}
	}
	return
}

//PowerGraph ...
type PowerGraph struct {
	Time  int64
	Power int64
}

var powerList []*PowerGraph

//GetPowerList ...
func GetPowerList() (list []*PowerGraph) {
	return powerList
}

func updateMinerPower(power *PowerGraph) {
	powerList = append(powerList, power)
	has := false
	for i, p := range powerList {
		if p.Time < time.Now().Unix()-24*60*60 {
			has = true
		}
		if p.Time > time.Now().Unix()-24*60*60 {
			if has {
				powerList = powerList[i:]
			}
			break
		}
	}
}

//SaveData ...
func SaveData() {
	storage.TipsetKV.Update(func(txn *badger.Txn) (err error) {
		data, _ := json.Marshal(historyBlockList)
		txn.Set([]byte("HISTORYBLOCKLIST"), data)
		data, _ = json.Marshal(historyMessageList)
		txn.Set([]byte("HISTORYMESSAGELIST"), data)
		data, _ = json.Marshal(historyTipSetList)
		txn.Set([]byte("HISTORYTIPSETLIST"), data)
		data, _ = json.Marshal(powerList)
		txn.Set([]byte("POWERLIST"), data)
		return nil
	})
}

//LoadData ...
func LoadData() {
	storage.TipsetKV.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte("HISTORYBLOCKLIST"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			json.Unmarshal(val, &historyBlockList)
			return nil
		})
		item, err = txn.Get([]byte("HISTORYMESSAGELIST"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			json.Unmarshal(val, &historyMessageList)
			return nil
		})
		item, err = txn.Get([]byte("HISTORYTIPSETLIST"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			json.Unmarshal(val, &historyTipSetList)
			return nil
		})
		item, err = txn.Get([]byte("POWERLIST"))
		if err != nil {
			return err
		}
		item.Value(func(val []byte) error {
			json.Unmarshal(val, &powerList)
			return nil
		})
		return
	})
}

//StartFetch ...
func StartFetch() {
	go func() {

		minute := time.NewTicker(time.Minute)
		minute10 := time.NewTicker(time.Minute * 10)
	Reconnect:
		fmt.Println("reconnecting")
		client, err := NewClient(etc.Config.EPIK.RPCHost, etc.Config.EPIK.RPCToken)
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		notify, err := client.FullNodeAPI.ChainNotify(context.Background())
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		var latestTipSet *types.TipSet
		for {
			select {
			case changes := <-notify:
				for _, change := range changes {
					latestTipSet = change.Val
					updateBaseInfoWithNewTipSet(change.Val)
					for _, block := range change.Val.Blocks() {
						blockMessages, err := client.FullNodeAPI.ChainGetBlockMessages(context.Background(), block.Cid())
						if err == nil {
							updateBaseInfoWithBlockMessages(blockMessages)
						} else {
							fmt.Println(err)
							goto Reconnect
						}
					}
				}
			case <-minute.C:
				if latestTipSet != nil {
					pledge, err := client.FullNodeAPI.StatePledgeCollateral(context.Background(), latestTipSet.Key())
					if err == nil {
						updateBaseInfoWithPledgeCollateral(pledge)
					} else {
						fmt.Println(err)
						goto Reconnect
					}

				}
			case <-minute10.C:
				addrs, err := client.FullNodeAPI.StateListMiners(context.Background(), latestTipSet.Key())
				powerG := &PowerGraph{
					Time: time.Now().Unix(),
				}
				if err == nil {
					for _, addr := range addrs {
						power, err := client.FullNodeAPI.StateMinerPower(context.Background(), addr, latestTipSet.Key())
						if err == nil {
							powerG.Power += power.MinerPower.QualityAdjPower.Int64()
						} else {
							fmt.Println(err)
							goto Reconnect
						}
					}
				}
				updateMinerPower(powerG)
			}
		}
	}()
}
