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
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/filecoin-project/specs-actors/actors/abi/big"
	"github.com/ipfs/go-cid"

	"github.com/EpiK-Protocol/go-epik/api"
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
func GetLatestBlocks(num int, asc bool) (list []*EBlockHeader) {
	storage.TipsetKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("TS:"))
		valid := []byte(fmt.Sprintf("TS:"))
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			it.Item().Value(func(value []byte) (err error) {
				tipset := &ETipSet{}
				err = json.Unmarshal(value, tipset)
				if err != nil {
					return
				}
				for _, block := range tipset.Blocks {
					if asc {
						list = append([]*EBlockHeader{block}, list...)
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

//UpdateByNewTipset ...
func UpdateByNewTipset(tipset *types.TipSet) (err error) {
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

func updatePre10M(fullAPI api.FullNode) (err error) {
	var start, end abi.ChainEpoch
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
			if msg.Message == nil {
				fmt.Println(msg)
				continue
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
	head, err := fullAPI.ChainHead(context.Background())
	miners, err := fullAPI.StateListMiners(context.Background(), head.Key())
	if err != nil {
		return
	}
	wei, _ := big.FromString("1000000000000000000")
	pledged := int64(0)
	won := int64(0)
	totalPledge := big.NewInt(0)
	for _, miner := range miners {
		actor, err := fullAPI.StateGetActor(context.Background(), miner, types.EmptyTSK)
		if err != nil {
			fmt.Println(err)
			continue
		}
		totalPledge = big.Add(totalPledge, actor.Balance)
		if big.Cmp(actor.Balance, big.Zero()) > 0 {
			pledged++
		}
		if big.Cmp(actor.Balance, big.Mul(big.NewInt(1000), wei)) > 0 {
			won++
		}
	}
	storage.PowerKV.Update(func(txn *badger.Txn) (err error) {
		type minerStatus struct {
			Total   int64
			Pledged int64
			Won     int64
		}

		status := &minerStatus{
			Total:   int64(len(miners)),
			Pledged: pledged,
			Won:     won,
		}
		data, _ := json.Marshal(status)
		now := time.Now()
		t := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		key := []byte(fmt.Sprintf("MINERSTATUS:%x", math.MaxInt64-t.Unix()))
		return txn.Set(key, data)
	})
	CurrentBaseInfo.PledgeCollateral = big.Div(totalPledge, wei).String()
	return
}

func updateMinerPower(fullAPI api.FullNode) (err error) {
	fmt.Println("updateMinerPower")

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
			powG.RawPower = power.TotalPower.RawBytePower.Int64()
			powG.QualityPower = power.TotalPower.QualityAdjPower.Int64()
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
	Time         int64
	RawPower     int64
	QualityPower int64
}

var powerList []*PowerGraph

//GetPowerList ...
func GetPowerList() (list []*PowerGraph) {
	return powerList
}

//StartUpdateByTime ...
func StartUpdateByTime() {
	minute10 := time.NewTicker(time.Minute * 60)

Reconnect:
	time.Sleep(5 * time.Second)
	fmt.Println("reconnecting")
	httpHeader := http.Header{}
	httpHeader.Set("Connection-Timeout", "30s")
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, closer, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	err = updatePre10M(fullAPI)
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	err = updateMinerPower(fullAPI)
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	for {
		select {
		case <-minute10.C:
			err := updatePre10M(fullAPI)
			if err != nil {
				fmt.Println(err)
				closer()
				goto Reconnect
			}
			err = updateMinerPower(fullAPI)
			if err != nil {
				fmt.Println(err)
				closer()
				goto Reconnect
			}
		}
	}
}

//StartFetch ...
func StartFetch() {

Reconnect:
	time.Sleep(5 * time.Second)
	fmt.Println("reconnecting")
	httpHeader := http.Header{}
	httpHeader.Set("Connection-Timeout", "30s")
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, closer, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	err = updatePre10M(fullAPI)
	if err != nil {
		fmt.Println(err)
	}
	err = updateMinerPower(fullAPI)
	if err != nil {
		fmt.Println(err)
	}
	notify, err := fullAPI.ChainNotify(context.Background())
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	resetTimer := time.NewTimer(time.Minute * 5)
	for {
		select {
		case <-resetTimer.C:
			closer()
			goto Reconnect
		case changes := <-notify:
			resetTimer.Reset(time.Minute * 5)
			for _, change := range changes {
				UpdateByNewTipset(change.Val)
				lastTipset, err := fullAPI.ChainGetTipSetByHeight(context.Background(), change.Val.Height()-1, types.EmptyTSK)
				if err != nil {
					fmt.Println("get tipset height", err)
					closer()
					goto Reconnect
				}
				err = SaveTipset(lastTipset, fullAPI)
				if err != nil {
					fmt.Println("save tipset", err)
					closer()
					goto Reconnect
				}
				msgs := []*types.Message{}
				for _, block := range lastTipset.Blocks() {
					blockMessages, err := fullAPI.ChainGetBlockMessages(context.Background(), block.Cid())
					if err != nil {
						fmt.Println("block messages", err)
						closer()
						goto Reconnect
					}
					for _, msg := range blockMessages.BlsMessages {
						msgs = append(msgs, msg)
					}
					for _, msg := range blockMessages.SecpkMessages {
						msgs = append(msgs, &msg.Message)
					}
				}
				err = SaveMessage(msgs, lastTipset)
				SaveLatestMessage(msgs)
				if err != nil {
					fmt.Println("save message", err)
					closer()
					goto Reconnect
				}
			}
		}
	}
}

//FetchHistory ...
func FetchHistory() {
Reconnect:
	time.Sleep(5 * time.Second)
	fmt.Println("reconnecting")
	httpHeader := http.Header{}
	httpHeader.Set("Connection-Timeout", "30s")
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, closer, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		fmt.Println(err)
		closer()
		goto Reconnect
	}
	tipset, err := fullAPI.ChainHead(context.Background())
	for tipset.Height() > 1 {

		storage.TipsetKV.Update(func(txn *badger.Txn) (err error) {
			key := []byte(fmt.Sprintf("TS:%x", math.MaxUint32-int(tipset.Height())))
			_, err = txn.Get(key)
			if err == badger.ErrKeyNotFound {
				err = SaveTipset(tipset, fullAPI)
				if err != nil {
					fmt.Println(err)
					return
				}
				msgs := []*types.Message{}
				for _, block := range tipset.Blocks() {
					blockMessages, err := fullAPI.ChainGetBlockMessages(context.Background(), block.Cid())
					if err != nil {
						fmt.Println(err)
						return err
					}
					for _, msg := range blockMessages.BlsMessages {
						msgs = append(msgs, msg)
					}
					for _, msg := range blockMessages.SecpkMessages {
						msgs = append(msgs, &msg.Message)
					}
				}
				err = SaveMessage(msgs, tipset)
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			return
		})
		tipset, err = fullAPI.ChainGetTipSetByHeight(context.Background(), tipset.Height()-1, types.EmptyTSK)
		if err != nil {
			fmt.Println(err)
			closer()
			goto Reconnect
		}
	}
}

//new---------------------------------------------------------

//EBlockHeader ...
type EBlockHeader struct {
	types.BlockHeader
	MinerReward types.BigInt
	GasReward   types.BigInt
}

//ETipSet ...
type ETipSet struct {
	Cids   []cid.Cid       `json:"Cids"`
	Blocks []*EBlockHeader `json:"Blocks"`
	Height abi.ChainEpoch  `json:"Height"`
}

//SaveTipset ...
func SaveTipset(tipset *types.TipSet, fullAPI api.FullNode) (err error) {
	if tipset == nil {
		return fmt.Errorf("tipset is nil")
	}
	fmt.Println("save Tipset:", tipset.Height())
	var ets *ETipSet
	data, _ := json.Marshal(tipset)
	ets = &ETipSet{}
	json.Unmarshal(data, ets)
	err = storage.TipsetKV.Update(func(txn *badger.Txn) (err error) {
		for _, bh := range ets.Blocks {
			reward, err := fullAPI.ChainGetBlockRewards(context.Background(), bh.Cid())
			if err != nil {
				return err
			}
			bh.MinerReward = reward.MinerReward
			bh.GasReward = reward.GasReward
		}
		data, err = json.Marshal(ets)
		key := []byte(fmt.Sprintf("TS:%x", math.MaxUint32-int(ets.Height)))
		return txn.Set(key, data)
	})

	return
}

//SaveReward ...
func SaveReward(height abi.ChainEpoch, fullAPI api.FullNode) (err error) {
	err = storage.TipsetKV.Update(func(txn *badger.Txn) (err error) {
		key := []byte(fmt.Sprintf("TS:%x", math.MaxUint32-int(height)))
		var ets = &ETipSet{}
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		err = json.Unmarshal(data, ets)
		if err != nil {
			return err
		}
		for _, bh := range ets.Blocks {
			reward, err := fullAPI.ChainGetBlockRewards(context.Background(), bh.Cid())
			if err != nil {
				return err
			}

			bh.MinerReward = reward.MinerReward
			bh.GasReward = reward.GasReward
		}
		data, err = json.Marshal(ets)
		return txn.Set(key, data)
	})
	return
}

//RMessage ...
type RMessage struct {
	Height  abi.ChainEpoch
	Time    uint64
	Message *types.Message
}

//SaveMessage ...
func SaveMessage(messages []*types.Message, tipset *types.TipSet) (err error) {

	return storage.MessageKV.Update(func(txn *badger.Txn) (err error) {
		for _, message := range messages {
			rm := &RMessage{
				Height:  tipset.Height(),
				Time:    tipset.MinTimestamp(),
				Message: message,
			}
			data, err := json.Marshal(rm)
			if err != nil {
				return err
			}
			key := []byte(fmt.Sprintf("MS:%s", message.Cid().String()))
			err = txn.Set(key, data)
			if err != nil {
				return err
			}

			from := []byte(fmt.Sprintf("UMS:%s:%s:F", message.From.String(), fmt.Sprintf("%x", math.MaxInt64-time.Now().UnixNano())))
			txn.Set(from, key)
			if err != nil {
				return err
			}
			to := []byte(fmt.Sprintf("UMS:%s:%s:T", message.To.String(), fmt.Sprintf("%x", math.MaxInt64-time.Now().UnixNano())))
			txn.Set(to, key)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

//ReadAddrMessage ...
func ReadAddrMessage(addr string, from time.Time, size int64) (messages []*RMessage, err error) {
	messages = []*RMessage{}
	err = storage.MessageKV.View(func(txn *badger.Txn) (err error) {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		prefix := []byte(fmt.Sprintf("UMS:%s:%s", addr, fmt.Sprintf("%x", math.MaxInt64-from.UnixNano())))
		valid := []byte(fmt.Sprintf("UMS:%s:", addr))
		for it.Seek(prefix); it.ValidForPrefix(valid); it.Next() {
			key, err := it.Item().ValueCopy(nil)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(string(key))
			item, err := txn.Get(key)
			if err != nil {
				fmt.Println(err)
				break
			}
			data, err := item.ValueCopy(nil)
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(string(data))
			msg := &RMessage{}
			err = json.Unmarshal(data, msg)
			if err == nil {
				messages = append(messages, msg)
			}
			if len(messages) >= int(size) {
				break
			}
		}
		return
	})
	return
}

//GetMessage ...
func GetMessage(cid string) (message *RMessage, err error) {
	message = &RMessage{}
	err = storage.MessageKV.View(func(txn *badger.Txn) (err error) {
		key := []byte(fmt.Sprintf("MS:%s", cid))
		item, err := txn.Get(key)
		if err != nil {
			return
		}
		data, err := item.ValueCopy(nil)
		if err != nil {
			return
		}
		err = json.Unmarshal(data, message)
		return err
	})
	return
}

var latestMessages = []*types.Message{}

//SaveLatestMessage ...
func SaveLatestMessage(msgs []*types.Message) (err error) {
	latestMessages = append(msgs, latestMessages...)
	if len(latestMessages) > 100 {
		latestMessages = latestMessages[0:100]
	}
	return
}

//GetLatestMessage ...
func GetLatestMessage() (msgs []*types.Message, err error) {
	return latestMessages, nil
}
