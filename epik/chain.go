package epik

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/filecoin-project/specs-actors/actors/abi/big"

	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/dgraph-io/badger/v2"
)

//CurrentBaseInfo ...
var CurrentBaseInfo *EBaseInformation

func init() {
	CurrentBaseInfo = &EBaseInformation{}
	powerList = []*PowerGraph{}
}

//GetLatestBlocks ...
func GetLatestBlocks(num int, asc bool) (list []*types.BlockHeader) {
	storage.TipsetKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("TS:"))
		valid := []byte(fmt.Sprintf("TS:"))
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			it.Item().Value(func(value []byte) (err error) {
				tipset := &types.TipSet{}
				err = json.Unmarshal(value, tipset)
				if err != nil {
					return
				}
				for _, block := range tipset.Blocks() {
					if asc {
						list = append([]*types.BlockHeader{block}, list...)
					} else {
						list = append(list, block)
					}

					if len(list) >= num {
						return
					}
				}
				return nil
			})
			if len(list) >= num {
				return
			}
		}
		return nil
	})
	return
}

//GetLatestTipSets ...
func GetLatestTipSets(length int) (list []*types.TipSet) {
	storage.TipsetKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("TS:"))
		valid := []byte(fmt.Sprintf("TS:"))
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			tipset := &types.TipSet{}
			err = it.Item().Value(func(value []byte) (err error) {
				return json.Unmarshal(value, tipset)
			})
			if err == nil {
				list = append(list, tipset)
			}
			if len(list) >= length {
				return
			}
		}
		return nil
	})
	return
}

//GetLatestMessages ...
func GetLatestMessages(num int) (list []*types.Message) {
	storage.MessageKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("MS:"))
		valid := []byte(fmt.Sprintf("MS:"))
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			it.Item().Value(func(value []byte) (err error) {
				msg := &RMessage{}
				err = json.Unmarshal(value, msg)
				if err != nil {
					return
				}
				list = append(list, msg.Message)
				return nil
			})
			if len(list) >= num {
				return
			}
		}
		return nil
	})
	return
}

//updateByNewTipset ...
func updateByNewTipset(tipset *types.TipSet) (err error) {
	// httpHeader := http.Header{}
	// httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	// fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	// if err != nil {
	// 	return
	// }
	CurrentBaseInfo.TipsetHeight = int64(tipset.Height())
	// pledge, err := fullAPI.StatePledgeCollateral(context.Background(), tipset.Key())
	// if err != nil {
	// 	return
	// }
	// CurrentBaseInfo.PledgeCollateral = pledge.String()

	return
}

func updatePre10M() (err error) {
	var start, end int64
	storage.MessageKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte("MS:")
		valid := []byte("MS:")
		var msgCount int64
		var gasPrice int64
		var msgLength int64

		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			msg := &RMessage{}
			data, err := it.Item().ValueCopy(nil)
			if err != nil {
				continue
			}
			err = json.Unmarshal(data, msg)
			if err != nil {
				continue
			}
			if start <= 0 {
				start = msg.Height
			}
			end = msg.Height
			msgCount++
			gasPrice += msg.Message.GasPrice.Int64()
			msgLength += int64(len(msg.Message.Params))
			if msgCount >= 1000 {
				break
			}
		}
		if msgCount > 0 {
			CurrentBaseInfo.AvgGasPrice = float64(gasPrice) / float64(msgCount)
			CurrentBaseInfo.AvgMessageSize = float64(msgLength) / float64(msgCount)
			if start > end {
				CurrentBaseInfo.AvgMessagesTipset = float64(msgCount) / float64(start-end)
			}
		}
		return
	})
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	miners, err := fullAPI.StateListMiners(context.Background(), head.Key())
	if err != nil {
		return
	}
	totalPledge := big.NewInt(0)
	for _, miner := range miners {
		sectors, err := fullAPI.StateMinerSectors(context.Background(), miner, nil, false, head.Key())
		if err != nil {
			return err
		}
		for _, sector := range sectors {
			pledge, err := fullAPI.StateMinerInitialPledgeCollateral(context.Background(), miner, sector.Info.Info.SectorNumber, head.Key())
			if err != nil {
				return err
			}
			totalPledge = big.Add(totalPledge, pledge)
		}
	}
	wei, _ := big.FromString("1000000000000000000")
	CurrentBaseInfo.PledgeCollateral = big.Div(totalPledge, wei).String()
	return
}

func updateMinerPower() (err error) {
	fmt.Println("updateMinerPower")
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		fmt.Println(err)
		return
	}
	head, err := fullAPI.ChainHead(context.Background())
	if err != nil {
		fmt.Println(err)
		return err
	}
	miners, err := fullAPI.StateListMiners(context.Background(), head.Key())
	if err != nil {
		fmt.Println(err)
		return err
	}
	powerList = []*PowerGraph{}
	for i := 0; i < 10; i++ {
		powG := &PowerGraph{
			Time: int64(head.MinTimestamp()),
		}
		if len(miners) > 0 {
			power, err := fullAPI.StateMinerPower(context.Background(), miners[0], head.Key())
			if err != nil {
				fmt.Println(err)
				return err
			}
			powG.Power = power.TotalPower.QualityAdjPower.Int64()
		}
		// fmt.Println(powG)
		powerList = append([]*PowerGraph{powG}, powerList...)
		head, err = fullAPI.ChainGetTipSetByHeight(context.Background(), head.Height()-100, head.Key())
		if err != nil {
			fmt.Println(err)
			return
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

//StartFetch ...
func StartFetch() {
	go func() {

		minute := time.NewTicker(time.Minute)
		minute10 := time.NewTicker(time.Minute * 10)
		err := updatePre10M()
		if err != nil {
			fmt.Println(err)
		}
		err = updateMinerPower()
		if err != nil {
			fmt.Println(err)
		}
	Reconnect:
		fmt.Println("reconnecting")
		httpHeader := http.Header{}
		httpHeader.Set("Content-Timeout", "100s")
		httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
		fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		notify, err := fullAPI.ChainNotify(context.Background())
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		for {
			select {
			case changes := <-notify:
				for _, change := range changes {
					SaveTipset(change.Val)
					msgs := []*types.Message{}
					for _, block := range change.Val.Blocks() {
						blockMessages, err := fullAPI.ChainGetBlockMessages(context.Background(), block.Cid())
						if err != nil {
							fmt.Println(err)
							goto Reconnect
						}
						for _, msg := range blockMessages.BlsMessages {
							msgs = append(msgs, msg)
						}
						for _, msg := range blockMessages.SecpkMessages {
							msgs = append(msgs, &msg.Message)
						}
					}
					SaveMessage(msgs, change.Val)
				}
			case <-minute.C:

			case <-minute10.C:
				err := updatePre10M()
				if err != nil {
					fmt.Println(err)
				}
				err = updateMinerPower()
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

//new---------------------------------------------------------

//SaveTipset ...
func SaveTipset(tipset *types.TipSet) (err error) {
	// fmt.Println(tipset)
	updateByNewTipset(tipset)
	return storage.TipsetKV.Update(func(txn *badger.Txn) (err error) {
		data, err := json.Marshal(tipset)
		if err != nil {
			return
		}
		key := []byte(fmt.Sprintf("TS:%x", math.MaxUint32-tipset.Height()))
		return txn.Set(key, data)
	})
}

//RMessage ...
type RMessage struct {
	Height  int64
	Time    time.Time
	Message *types.Message
}

//SaveMessage ...
func SaveMessage(messages []*types.Message, tipset *types.TipSet) (err error) {
	return storage.MessageKV.Update(func(txn *badger.Txn) (err error) {
		for _, message := range messages {
			data, err := json.Marshal(message)
			if err != nil {
				return err
			}
			key := []byte(fmt.Sprintf("MS:%s", message.Cid().String()))
			err = txn.Set(key, data)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
