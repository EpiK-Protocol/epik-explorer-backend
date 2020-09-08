package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/EpiK-Protocol/epik-explorer-backend/api"
	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
)

func init() {
	err := etc.Load("./conf/config.yml")
	panicErr(err)
	storage.InitDatabase()
	epik.RegisterModels(storage.DB)
}

func TestGenMinerProfit(t *testing.T) {

	err := api.GenTestnetMinerBonusByBlock()
	panicErr(err)
}

func panicErr(err error) {
	if err != nil {
		panic(err)
	}
}

func TestMinerBind(t *testing.T) {

	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	panicErr(err)
	head, err := fullAPI.ChainHead(context.Background())
	panicErr(err)
	miners, err := fullAPI.StateListMiners(context.Background(), head.Key())
	panicErr(err)
	for _, miner := range miners {
		data, _ := json.Marshal(miner)
		t.Log(string(data))
		info, err := fullAPI.StateMinerInfo(context.Background(), miner, head.Key())
		panicErr(err)
		data, _ = json.Marshal(info)
		t.Log(string(data))
		balance, err := fullAPI.StateMinerAvailableBalance(context.Background(), miner, head.Key())
		panicErr(err)
		t.Log("balance:", balance.String())
		cids, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: miner}, head.Key(), 0)
		panicErr(err)
		if len(cids) > 0 {
			firstMsg, err := fullAPI.ChainGetMessage(context.Background(), cids[0])
			panicErr(err)
			data, _ = json.Marshal(firstMsg)
			t.Log(string(data))
		}
	}
}

func TestGenMinerProfitByEPK(t *testing.T) {
	err := api.GenTestnetMinerBonusByEPK()
	panicErr(err)
}

func TestPledgedMiner(t *testing.T) {
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	panicErr(err)
	head, err := fullAPI.ChainHead(context.Background())
	panicErr(err)
	miners, err := fullAPI.StateListMiners(context.Background(), head.Key())
	panicErr(err)
	o := storage.DB
	bonus := []*epik.Miner{}
	for _, addr := range miners {
		balance, _ := fullAPI.StateMinerAvailableBalance(context.Background(), addr, head.Key())
		miner := &epik.Miner{}
		err := o.Model(epik.Miner{}).Where("miner_id = ?", addr.String()).First(miner).Error
		if err == nil && balance.String() != "0" {
			fmt.Printf("{\nID:%s\nMinerID:%s\nAddress:%s\n,PledgeEPK:%s\n}\n", miner.ID, miner.MinerID, miner.EpikAddress, balance.String())
			bonus = append(bonus, miner)
		}
	}
	for _, miner := range bonus {
		fmt.Printf("%s,%s,%s,%s,\n", miner.ID, miner.Erc20Address, miner.EpikAddress, miner.WeiXin)
	}
}
