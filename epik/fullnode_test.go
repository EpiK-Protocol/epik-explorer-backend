package epik

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/EpiK-Protocol/go-epik/api"
	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/filecoin-project/go-jsonrpc"
)

var fullNodeAPI api.FullNode
var fullNodeCloser jsonrpc.ClientCloser

func init() {
	api, closer, err := client.NewFullNodeRPC("ws://120.55.82.202:1234/rpc/v0", nil)
	if err != nil {
		panic(err)
	}
	fullNodeAPI = api
	fullNodeCloser = closer
}

func TestChainHead(t *testing.T) {
	tipset, err := fullNodeAPI.ChainHead(context.Background())
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tipset)
	fmt.Println(string(data))
}

func TestChainNotify(t *testing.T) {
	notify, err := fullNodeAPI.ChainNotify(context.Background())
	if err != nil {
		panic(err)
	}
	for {
		select {

		case change := <-notify:
			data, _ := json.Marshal(change)
			fmt.Println(string(data))
		}
	}
}

func TestStateListMiners(t *testing.T) {
	tipset, err := fullNodeAPI.ChainHead(context.Background())
	if err != nil {
		panic(err)
	}
	data, _ := json.Marshal(tipset)
	fmt.Printf("tipset:%s\n", string(data))
	for _, block := range tipset.Blocks() {
		bmsgs, _ := fullNodeAPI.ChainGetBlockMessages(context.Background(), block.Cid())
		data, _ := json.Marshal(bmsgs)
		fmt.Printf("blockmessages:%s\n", string(data))
	}
	addrs, err := fullNodeAPI.StateListMiners(context.Background(), tipset.Key())
	for _, addr := range addrs {
		power, _ := fullNodeAPI.StateMinerPower(context.Background(), addr, tipset.Key())
		fmt.Printf("power:%d\n", power)
		info, _ := fullNodeAPI.StateMinerInfo(context.Background(), addr, tipset.Key())
		data, _ := json.Marshal(&info)
		fmt.Printf("info:%s\n", string(data))
		sectors, _ := fullNodeAPI.StateMinerSectors(context.Background(), addr, nil, false, tipset.Key())
		data, _ = json.Marshal(sectors)
		fmt.Printf("sectors:%s\n", string(data))
		baseInfo, _ := fullNodeAPI.MinerGetBaseInfo(context.Background(), addr, 1, tipset.Key())
		data, _ = json.Marshal(baseInfo)
		fmt.Printf("baseinfo:%s\n", string(data))
		messages, _ := fullNodeAPI.StateListMessages(context.Background(), nil, tipset.Key(), 1)
		data, _ = json.Marshal(messages)
		fmt.Printf("messages:%s\n", string(data))
	}

}
func TestStateListActors(t *testing.T) {
	tipset, err := fullNodeAPI.ChainHead(context.Background())
	if err != nil {
		panic(err)
	}
	actors, err := fullNodeAPI.StateListActors(context.Background(), tipset.Key())
	fmt.Println(actors)
}
