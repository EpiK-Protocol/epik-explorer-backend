package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/etc"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"
	"github.com/dgraph-io/badger/v2"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/EpiK-Protocol/go-epik/api/client"
	"github.com/EpiK-Protocol/go-epik/chain/types"
	"github.com/EpiK-Protocol/go-epik/lib/sigs"
	_ "github.com/EpiK-Protocol/go-epik/lib/sigs/bls"
	_ "github.com/EpiK-Protocol/go-epik/lib/sigs/secp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/specs-actors/actors/abi"
	epikcrypto "github.com/filecoin-project/specs-actors/actors/crypto"
	"github.com/gin-gonic/gin"
	"github.com/gofrs/uuid"
)

//SetTestNetAPI ...
func SetTestNetAPI(e *gin.Engine) {
	e.POST("/testnet/signup", testnetSignUp)
	e.GET("/testnet/home", testnetHome)
	e.GET("/testnet/profit", testnetProfit)
}

func testnetSignUp(c *gin.Context) {
	req := &struct {
		Weixin         string `json:"weixin"`
		EpikAddress    string `json:"epik_address"`
		Erc20Address   string `json:"erc20_address"`
		EpikSignature  string `json:"epik_signature"`
		Erc20Signature string `json:"erc20_signature"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, errClientError)
		return
	}
	if isEmpty(req.Weixin, req.EpikAddress, req.Erc20Address, req.EpikSignature, req.Erc20Signature) {
		responseJSON(c, errClientError)
		return
	}
	hash := sha256.Sum256([]byte(req.Weixin))

	erc20Addr := common.HexToAddress(req.Erc20Address)

	erc20signature, err := hex.DecodeString(req.Erc20Signature)
	if err != nil {
		responseJSON(c, clientError(fmt.Errorf("erc20 signature error:%s", err.Error())))
		return
	}

	ercpubkey, err := crypto.SigToPub(hash[:], erc20signature)
	if err != nil {
		responseJSON(c, clientError(fmt.Errorf("erc20 signature faild:%s", err.Error())))
		return
	}
	erc20signaddr := crypto.PubkeyToAddress(*ercpubkey)
	if erc20signaddr != erc20Addr {
		responseJSON(c, clientError(fmt.Errorf("erc20 signature faild:%s", erc20signaddr.Hex())))
		return
	}

	epikAddr, err := address.NewFromString(req.EpikAddress)
	if err != nil {
		responseJSON(c, clientError(fmt.Errorf("epik address error:%s", err.Error())))
		return
	}
	epikSignature, err := hex.DecodeString(req.EpikSignature)
	if err != nil {
		responseJSON(c, clientError(fmt.Errorf("epik signature error:%s", err.Error())))
		return
	}

	epiksig := &epikcrypto.Signature{}
	epiksig.UnmarshalBinary(epikSignature)
	err = sigs.Verify(epiksig, epikAddr, hash[:])
	if err != nil {
		responseJSON(c, clientError(fmt.Errorf("epik signature faild:%s", err.Error())))
		return
	}

	id := uuid.NewV5(uuid.NamespaceDNS, req.Weixin)
	miner := &epik.Miner{
		ID:           id.String(),
		WeiXin:       req.Weixin,
		EpikAddress:  req.EpikAddress,
		Erc20Address: req.Erc20Address,
		CreatedAt:    time.Now(),
		Status:       epik.MinerStatusPending,
	}
	err = miner.Create(storage.DB)
	if err != nil {
		responseJSON(c, serverError(fmt.Errorf("报名失败，微信或钱包账号重复")))
		return
	}
	responseJSON(c, errOK, "id", miner.ID)
}

func testnetHome(c *gin.Context) {
	address := c.Query("address")
	testnet := &epik.TestNet{}
	err := storage.TestNetKV.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte("TESTNETCACHE"))
		if err != nil {
			return err
		}
		err = item.Value(func(value []byte) (err error) {
			return json.Unmarshal(value, testnet)
		})
		return err
	})

	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	if isEmpty(address) {
		responseJSON(c, errOK, "testnet", testnet)
	} else {
		miner, err := epik.GetMinerByERC20Address(storage.DB, address)
		if err != nil {
			responseJSON(c, errOK, "testnet", testnet)
		} else {
			responseJSON(c, errOK, "testnet", testnet, "id", miner.ID, "status", miner.Status)
		}
	}
}

func testnetProfit(c *gin.Context) {
	userID := c.Query("id")
	o := storage.DB
	profits := []*epik.ProfitRecord{}
	var tepk, erc20epk float64
	err := o.Model(epik.ProfitRecord{}).Where("miner_id = ?  AND status= ?", userID, epik.MinerStatusConfirmed).Order("id DESC").Find(&profits).Error
	if err != nil {
		responseJSON(c, errServerError)
		return
	}
	o.Raw("SELECT SUM(tepk) FROM profit_record WHERE miner_id = ? AND status= ?;", userID, epik.MinerStatusConfirmed).Scan(&tepk)
	o.Raw("SELECT SUM(erc20_epk) FROM profit_record WHERE miner_id = ?  AND status= ?;", userID, epik.MinerStatusConfirmed).Scan(&erc20epk)
	responseJSON(c, errOK, "tepk", tepk, "erc20_epk", erc20epk, "list", profits)
}

//RefreshTestNet ...
func RefreshTestNet(o *gorm.DB) (err error) {
	fmt.Println("refresh testnet")
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	testnet := &epik.TestNet{}
	testnet.TotalSupply = 2_500_000
	o.Raw("SELECT SUM(erc20_epk) FROM profit_record WHERE status = 'confirmed';").Scan(&testnet.Issuance)
	o.Model(&epik.Miner{}).Where("status = ? AND profit > 0", epik.MinerStatusConfirmed).Order("profit DESC").Limit(100).Find(&(testnet.TopList))
	data, _ := json.Marshal(testnet)
	storage.TestNetKV.Update(func(txn *badger.Txn) (err error) {
		return txn.Set([]byte("TESTNETCACHE"), data)
	})
	return
}

//GenTestnetMinerBonusByBlock ...
func GenTestnetMinerBonusByBlock() (err error) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
		}
	}()
	o := storage.DB
	var pending int64
	o.Model(epik.ProfitRecord{}).Where("status = ?", epik.MinerStatusPending).Count(&pending)
	if pending > 0 {
		return nil
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
	minerWork := map[string]int64{}
	now := time.Now()
	end := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, now.Location())
	start := end.Add(-time.Hour * 24)
	tipset := head
	var blockCount int64
	height := tipset.Height()
Reconnect:
	fullAPI, _, err = client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		time.Sleep(time.Second)
		goto Reconnect
	}
	for {
		fmt.Print(height, ",")
		tipset, err = fullAPI.ChainGetTipSetByHeight(context.Background(), height, head.Key())
		if err != nil {
			goto Reconnect
		}
		if tipset.MinTimestamp() > uint64(end.Unix()) {
			continue
		}
		if tipset.MinTimestamp() < uint64(start.Unix()) {
			break
		}
		if height <= 0 {
			break
		}
		for _, block := range tipset.Blocks() {
			minerWork[block.Miner.String()]++
			blockCount++
		}
		height--
	}
	fmt.Printf("find blocks:%d\n", blockCount)
	fmt.Println(minerWork)
	confirmMiners := []*epik.Miner{}
	o.Model(epik.Miner{}).Where("status = ?", epik.MinerStatusConfirmed).Find(&confirmMiners)
	confirmMinerMap := map[string]*epik.Miner{}
	for _, miner := range confirmMiners {
		confirmMinerMap[miner.EpikAddress] = miner
	}
	var totalBlock int64
	confirmWorks := map[*epik.Miner]int64{}
	for addr, count := range minerWork {
		if confirmMinerMap[addr] != nil {
			confirmWorks[confirmMinerMap[addr]] = count
			totalBlock += count
		}
	}
	if totalBlock <= 0 {
		return
	}
	bonus := etc.Config.EPIKERC20.DailyBonus / float64(totalBlock)
	o.Begin()
	for miner, count := range confirmWorks {
		record := &epik.ProfitRecord{
			MinerID:   miner.ID,
			ERC20EPK:  bonus * float64(count),
			CreatedAt: time.Now(),
			Status:    epik.MinerStatusPending,
		}
		err = record.Create(o)
		if err != nil {
			o.Rollback()
			return err
		}
	}
	return o.Commit().Error
}

const godAddress = "t3sbsmbkmvu7kpc5adrcaubyuna33t3gl2medbcjo65ttcuppb45srztz52lgy4iqrf3nwle2okvfhlf5xddgq"

//GenTestnetMinerBonusByEPK ...
func GenTestnetMinerBonusByEPK() (err error) {
	lastHeight := abi.ChainEpoch(0)
	toHeight := abi.ChainEpoch(0)
	storage.TestNetKV.View(func(txn *badger.Txn) (err error) {
		item, err := txn.Get([]byte(fmt.Sprintf("TESTNETMINERBONUSLASTHEIGHT")))
		if err == nil {
			item.Value(func(value []byte) error {
				lastHeight = abi.ChainEpoch(utils.ParseInt64(string(value)))
				return nil
			})
		}
		return
	})
	god, err := address.NewFromString(godAddress)
	if err != nil {
		return err
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
	if head.Height() < toHeight {
		err = fmt.Errorf("toHeight > chainHead")
		return
	}
	head, err = fullAPI.ChainGetTipSetByHeight(context.Background(), toHeight, head.Key())
	if err != nil {
		return
	}
	fmt.Println("head:", head)
	minerIncome := map[*epik.Miner]decimal.Decimal{}
	total := decimal.NewFromInt(0)
	fmt.Println("lastheight:", lastHeight, "   toHeight:", toHeight)
Reconnect:
	fullAPI, _, err = client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		time.Sleep(time.Second)
		goto Reconnect
	}
	fmt.Println(god)
	allMiner, err := fullAPI.StateListMiners(context.Background(), head.Key())
	if err != nil {
		fmt.Println(err)
		goto Reconnect
	}
	fmt.Println(allMiner)
	o := storage.DB
	confirmMiners := []*epik.Miner{}
	o.Model(epik.Miner{}).Where("status = ?", epik.MinerStatusConfirmed).Find(&confirmMiners)
	confirmMinerMap := map[string]*epik.Miner{} //全部过审
	bindMinerMap := map[string]*epik.Miner{}    //已绑定MinerID
	for _, miner := range confirmMiners {
		confirmMinerMap[miner.EpikAddress] = miner
		if miner.MinerID != "" {
			bindMinerMap[miner.MinerID] = miner
		}
	}
	// fmt.Println(confirmMinerMap)
	fmt.Println(bindMinerMap)
	//绑定矿工
	for _, miner := range allMiner {
		if bindMinerMap[miner.String()] == nil { //未绑定
			bindCids, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: miner}, head.Key(), 0)
			if err != nil {
				fmt.Println(err)
				goto Reconnect
			}
			if len(bindCids) > 0 {
				msg, err := fullAPI.ChainGetMessage(context.Background(), bindCids[0])
				if err != nil {
					fmt.Println(err)
					goto Reconnect
				}
				fmt.Println("bind:", msg)
				unbindMiner := confirmMinerMap[msg.From.String()]
				if unbindMiner != nil { //找到可以绑定的矿工，计算有效收益
					unbindMiner.MinerID = miner.String()
					bindMinerMap[miner.String()] = unbindMiner
					err = unbindMiner.Update(o, "miner_id")
					if err != nil {
						fmt.Println(err)
						goto Reconnect
					}
				}
			}
		}
	}
	dec200, _ := decimal.NewFromString("200000000000000000000")
	fmt.Println("miner all bind")
	for _, miner := range bindMinerMap {
		if minerIncome[miner].Cmp(decimal.Zero) > 0 {
			continue
		}
		minerAddress, err := address.NewFromString(miner.MinerID)
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		cids, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: minerAddress}, head.Key(), lastHeight)
		if err != nil {
			fmt.Println(err)
			goto Reconnect
		}
		fmt.Println("miner", miner.MinerID, "msg size:", len(cids))
		income := decimal.Zero
		for i, cid := range cids {
			msg, err := fullAPI.ChainGetMessage(context.Background(), cid)
			if err != nil {
				fmt.Println(err)
				goto Reconnect
			}
			value, err := decimal.NewFromString(msg.Value.String())
			if value.Cmp(dec200) == 0 {
				continue
			}
			income = income.Add(value)
			fmt.Printf("\r%d:%s:%s", i, value, income)
		}
		fmt.Println("\nincome:", income)
		minerIncome[miner] = income
	}
	fmt.Println(minerIncome)
	total = decimal.Zero
	for _, income := range minerIncome {
		total = total.Add(income)
	}
	fmt.Println("total", total)
	if total.Cmp(decimal.Zero) <= 0 {
		return
	}
	fmt.Println("daily bonus:", etc.Config.EPIKERC20.DailyBonus)

	wei, _ := decimal.NewFromString("1000000000000000000")
	bonusUnit := decimal.NewFromFloat(etc.Config.EPIKERC20.DailyBonus).Mul(wei).Div(total)
	fmt.Println("bonus unit:", bonusUnit)
	o = o.Begin()
	for miner, tepk := range minerIncome {
		if tepk.Equal(decimal.Zero) {
			continue
		}
		record := &epik.ProfitRecord{
			MinerID:   miner.ID,
			CreatedAt: time.Now(),
			Status:    epik.MinerStatusPending,
		}
		record.TEPK, _ = tepk.Div(wei).Float64()
		record.ERC20EPK, _ = tepk.Div(wei).Mul(bonusUnit).Float64()
		fmt.Println(record)
		err = record.Create(o)
		if err != nil {
			o.Rollback()
			return err
		}
	}
	return o.Commit().Error
}

//GenTestnetMinerBonusByPledge ...
func GenTestnetMinerBonusByPledge() (err error) {
	httpHeader := http.Header{}
	httpHeader.Set("Authorization", fmt.Sprintf("Bearer %s", etc.Config.EPIK.RPCToken))
	fullAPI, _, err := client.NewFullNodeRPC(etc.Config.EPIK.RPCHost, httpHeader)
	if err != nil {
		return
	}
Reconnect:
	head, err := fullAPI.ChainHead(context.Background())
	if err != nil {
		return
	}
	allMiner, err := fullAPI.StateListMiners(context.Background(), head.Key())
	if err != nil {
		fmt.Println(err)
		goto Reconnect
	}
	fmt.Println(allMiner)
	o := storage.DB
	confirmMiners := []*epik.Miner{}
	o.Model(epik.Miner{}).Where("status = ?", epik.MinerStatusConfirmed).Find(&confirmMiners)
	confirmMinerMap := map[string]*epik.Miner{} //全部过审
	bindMinerMap := map[string]*epik.Miner{}    //已绑定MinerID
	for _, miner := range confirmMiners {
		confirmMinerMap[miner.EpikAddress] = miner
		if miner.MinerID != "" {
			bindMinerMap[miner.MinerID] = miner
		}
	}
	// fmt.Println(confirmMinerMap)
	fmt.Println(bindMinerMap)
	//绑定矿工
	for _, miner := range allMiner {
		if bindMinerMap[miner.String()] == nil { //未绑定
			bindCids, err := fullAPI.StateListMessages(context.Background(), &types.Message{To: miner}, head.Key(), 0)
			if err != nil {
				fmt.Println(err)
				goto Reconnect
			}
			if len(bindCids) > 0 {
				msg, err := fullAPI.ChainGetMessage(context.Background(), bindCids[0])
				if err != nil {
					fmt.Println(err)
					goto Reconnect
				}
				fmt.Println("bind:", msg)
				unbindMiner := confirmMinerMap[msg.From.String()]
				if unbindMiner != nil { //找到可以绑定的矿工，计算有效收益
					unbindMiner.MinerID = miner.String()
					bindMinerMap[miner.String()] = unbindMiner
					err = unbindMiner.Update(o, "miner_id")
					if err != nil {
						fmt.Println(err)
						goto Reconnect
					}
				}
			}
		}
	}
	fmt.Println("miner all bind")
	bonus := []*epik.Miner{}
	for _, addr := range allMiner {
		balance, _ := fullAPI.StateMinerAvailableBalance(context.Background(), addr, head.Key())
		miner := &epik.Miner{}
		err := o.Model(epik.Miner{}).Where("miner_id = ?", addr.String()).First(miner).Error
		if err == nil && balance.String() != "0" {
			fmt.Printf("{\nID:%s\nMinerID:%s\nAddress:%s\n,PledgeEPK:%s\n}\n", miner.ID, miner.MinerID, miner.EpikAddress, balance.String())
			bonus = append(bonus, miner)
		}
	}
	if len(bonus) <= 0 {
		return
	}
	bonusUnit := etc.Config.EPIKERC20.DailyBonus / float64(len(bonus))
	for _, miner := range bonus {
		record := &epik.ProfitRecord{
			MinerID:   miner.ID,
			TEPK:      0,
			ERC20EPK:  bonusUnit,
			Status:    epik.MinerStatusPending,
			CreatedAt: time.Now(),
		}
		record.Create(storage.DB)
	}
	return
}

// //PushMinerERC20Bonus ...
// func PushMinerERC20Bonus() (err error) {
// 	o := storage.DB
// 	for {
// 		err = o.Begin().Error
// 		if err != nil {
// 			return
// 		}
// 		record := &epik.ProfitRecord{}
// 		err = o.Model(epik.ProfitRecord{}).Where("status = ?", epik.MinerStatusPending).First(record).Error
// 		if err == gorm.ErrRecordNotFound {
// 			o.Rollback()
// 			return nil
// 		}
// 		if err != nil {
// 			o.Rollback()
// 			return
// 		}
// 		miner := &epik.Miner{}
// 		miner, err = epik.GetMiner(o, record.MinerID)
// 		if err != nil {
// 			o.Rollback()
// 			return
// 		}
// 		var txHash string
// 		txHash, err = TransferToken(miner.Erc20Address, "EPK", fmt.Sprintf("%f", record.ERC20EPK))
// 		if err != nil {
// 			o.Rollback()
// 			return
// 		}
// 		miner.Profit += record.ERC20EPK
// 		record.Hash = txHash
// 		record.Status = epik.MinerStatusConfirmed
// 		err = miner.Update(o, "profit")
// 		if err != nil {
// 			o.Rollback()
// 			return
// 		}
// 		err = record.Update(o, "hash", "status")
// 		if err != nil {
// 			o.Rollback()
// 			return
// 		}
// 	}
// 	return
// }

type currencyType string

const (
	USDT currencyType = "USDT"
	EPK  currencyType = "EPK"
)

var contractAddress = map[currencyType]string{
	USDT: "0xdac17f958d2ee523a2206206994597c13d831ec7",
	EPK:  "0xDaF88906aC1DE12bA2b1D2f7bfC94E9638Ac40c4",
}

// //TransferToken ...
// func TransferToken(to string, currency string, amount string) (txHash string, err error) {

// 	client, err := ethclient.DialContext(context.Background(), etc.Config.EPIKERC20.RPCHost)
// 	if err != nil {
// 		return
// 	}
// 	defer client.Close()
// 	privateKey, err := crypto.HexToECDSA(etc.Config.EPIKERC20.MainPrivateKey)
// 	if err != nil {
// 		return "", err
// 	}
// 	pub := *&privateKey.PublicKey
// 	fromAddr := crypto.PubkeyToAddress(pub)
// 	toAddr := common.HexToAddress(to)
// 	switch currencyType(currency) {
// 	case USDT:
// 		contract := common.HexToAddress(contractAddress[USDT])
// 		usdtToken, err := usdt.NewUsdt(contract, client)
// 		if err != nil {
// 			return "", err
// 		}
// 		opts := &bind.CallOpts{}
// 		bal, err := usdtToken.BalanceOf(opts, fromAddr)
// 		if err != nil {
// 			return "", err
// 		}
// 		dec, err := usdtToken.Decimals(opts)
// 		amountDec, err := decimal.NewFromString(amount)
// 		if err != nil {
// 			return "", err
// 		}

// 		amountWei := amountDec.Mul(decimal.NewFromFloat(math.Pow10(int(dec.Int64()))))
// 		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
// 			return "", fmt.Errorf("Out of Balance")
// 		}
// 		gasPrice, err := client.SuggestGasPrice(context.Background())
// 		if err != nil {
// 			return "", err
// 		}
// 		nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
// 		if err != nil {
// 			return "", err
// 		}

// 		auth := bind.NewKeyedTransactor(privateKey)
// 		auth.Nonce = big.NewInt(int64(nonce))
// 		auth.GasLimit = uint64(100000)
// 		auth.GasPrice = gasPrice

// 		tx, err := usdtToken.Transfer(auth, toAddr, amountWei.BigInt())
// 		if err != nil {
// 			return "", err
// 		}
// 		txHash = tx.Hash().String()
// 	case EPK:
// 		contract := common.HexToAddress(contractAddress[EPK])
// 		epkToken, err := epk.NewEpk(contract, client)
// 		if err != nil {
// 			return "", err
// 		}
// 		opts := &bind.CallOpts{}
// 		bal, err := epkToken.BalanceOf(opts, fromAddr)
// 		if err != nil {
// 			return "", err
// 		}
// 		dec, err := epkToken.Decimals(opts)
// 		amountDec, err := decimal.NewFromString(amount)
// 		if err != nil {
// 			return "", err
// 		}

// 		amountWei := amountDec.Mul(decimal.NewFromFloat(math.Pow10(int(dec))))
// 		if amountWei.Cmp(decimal.NewFromBigInt(bal, 10)) > 0 {
// 			return "", fmt.Errorf("Out of Balance")
// 		}
// 		gasPrice, err := client.SuggestGasPrice(context.Background())
// 		if err != nil {
// 			return "", err
// 		}
// 		nonce, err := client.PendingNonceAt(context.Background(), fromAddr)
// 		if err != nil {
// 			return "", err
// 		}

// 		auth := bind.NewKeyedTransactor(privateKey)
// 		auth.Nonce = big.NewInt(int64(nonce))
// 		auth.GasLimit = uint64(100000)
// 		auth.GasPrice = gasPrice

// 		tx, err := epkToken.Transfer(auth, toAddr, amountWei.BigInt())
// 		if err != nil {
// 			return "", err
// 		}
// 		txHash = tx.Hash().String()
// 	default:
// 		return "", fmt.Errorf("Currency  Unsuppoted")
// 	}
// 	return
// }
