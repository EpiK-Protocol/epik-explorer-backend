package epik

//EPeer ...
type EPeer struct {
	PeerID string
	IP     string
}

//EPeerPoint ...
type EPeerPoint struct {
	LocationCN string
	LocationEN string
	Longitude  float64
	Latitude   float64
	Peers      []*EPeer
}

//EBaseInformation ...
type EBaseInformation struct {
	TipsetHeight      int64   `json:"tipset_height"`
	BlockReward       float64 `json:"block_reward"`
	AvgMessageSize    float64 `json:"avg_message_size"`
	AvgGasPrice       float64 `json:"avg_gas_price"`
	AvgMessagesTipset float64 `json:"avg_messages_tipset"`
	PledgeCollateral  string  `json:"pledge_collateral"`
	HeadUpdate        int64   `json:"head_update"`
	Outstanding       float64 `json:"outstanding"`
	AvgBlockTipset    float64 `json:"avg_block_tipset"`
}

//ETipset ...
type ETipset struct {
	CID       string
	MinerID   string
	Height    int64
	BlockTime int64
	TickCount int64
}

//ETipsetOwn ...
type ETipsetOwn struct {
	Tipset         []*ETipset
	Height         int64
	MinTicketBlock string
}

//EWonList ...
type EWonList struct {
	Miner       string
	Percent     float64
	TickPercent float64
}

//EPowerGraph ...
type EPowerGraph struct {
	Time                 int64
	Power                int64
	QualityAdjustedPower int64
}

//EMinerPowerAtTime ...
type EMinerPowerAtTime struct {
	AtTime      int64
	MinerStates struct {
		Address                string
		Power                  int64
		QualityAdjPower        int64
		PowerPercent           float64
		QualityAdjPowerPercent float64
		PeerID                 string
	}
}

//EBlock ...
type EBlock struct {
	BlockHeader struct {
		Miner                 string
		Tickets               string
		EPostProofLen         int64 `json:"e_post_proof_len"`
		Parents               []string
		ParentWeight          int64
		Height                int64
		ParentStateRoot       string
		ParentMessageReceipts string
		Messages              string
		Timestamp             int64
		BLSAggregate          struct {
			Type int64
			Data string
		} `json:"bls_aggregate"`
		BlockSig struct {
			Type int64
			Data string
		}
	}
	CID     string
	Size    int64
	MsgCIDs []string `json:"msg_cids"`
	Reward  float64
}

//EMessage ...
type EMessage struct {
	Msg struct {
		From     string
		To       string
		Nonce    int64
		Value    float64
		Gasprice float64
		gaslimit float64
		Method   int64
		Params   string
	}
	CID        string
	SignCID    string
	Size       int64
	Msgcreate  int64
	Height     int64
	BlockCIDs  []string `json:"block_cids"`
	ExitCode   string
	Return     string
	GasUsed    int64
	MethodName string
}
