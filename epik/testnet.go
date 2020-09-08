package epik

import (
	"time"

	"gorm.io/gorm"
)

//ProfitRecord ...
type ProfitRecord struct {
	ID        int64       `json:"id"`
	MinerID   string      `json:"miner_id"`
	TEPK      float64     `json:"tepk" gorm:"column:tepk"`
	ERC20EPK  float64     `json:"erc20_epk" gorm:"column:erc20_epk"`
	Hash      string      `json:"hash"`
	Status    MinerStatus `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
}

//TestNet ...
type TestNet struct {
	TotalSupply float64  `json:"total_supply"`
	Issuance    float64  `json:"issuance"`
	TopList     []*Miner `json:"top_list"`
}

type MinerStatus string

const (
	MinerStatusPending   MinerStatus = "pending"
	MinerStatusConfirmed MinerStatus = "confirmed"
	MinerStatusRejected  MinerStatus = "rejected"
)

//Miner ...
type Miner struct {
	ID           string      `json:"id" gorm:"primarykey"`
	WeiXin       string      `json:"wei_xin" gorm:"unique"`
	MinerID      string      `json:"miner_id"`
	EpikAddress  string      `json:"epik_address"`
	Erc20Address string      `json:"erc20_address"`
	Status       MinerStatus `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	Profit       float64     `json:"profit"`
	Airdrop      float64     `json:"airdrop"`
}

//TableName ...
func (Miner) TableName() string {
	return "miner"
}

//TableName ...
func (ProfitRecord) TableName() string {
	return "profit_record"
}

//Create ...
func (record *ProfitRecord) Create(o *gorm.DB) (err error) {
	return o.Create(record).Error
}

//GetProfitRecord ...
func GetProfitRecord(o *gorm.DB, id int64) (record *ProfitRecord, err error) {
	record = &ProfitRecord{}
	err = o.Model(record).First(record, id).Error
	return
}

//Update ...
func (record *ProfitRecord) Update(o *gorm.DB, columns ...string) (err error) {
	if len(columns) > 0 {
		err = o.Model(record).Select(columns).UpdateColumns(record).Error
	} else {
		err = o.Save(record).Error
	}
	return
}

//Delete ...
func (record *ProfitRecord) Delete(o *gorm.DB) (err error) {
	err = o.Delete(record).Error
	return
}

//Create ...
func (miner *Miner) Create(o *gorm.DB) (err error) {
	return o.Create(miner).Error
}

//GetMiner ...
func GetMiner(o *gorm.DB, id string) (miner *Miner, err error) {
	miner = &Miner{}
	err = o.Model(Miner{}).Where("id = ?", id).First(miner).Error
	return
}

//GetMinerByERC20Address ...
func GetMinerByERC20Address(o *gorm.DB, address string) (miner *Miner, err error) {
	miner = &Miner{}
	err = o.Model(Miner{}).Where("erc20_address = ?", address).First(miner).Error
	return
}

//Update ...
func (miner *Miner) Update(o *gorm.DB, columns ...string) (err error) {
	if len(columns) > 0 {
		err = o.Model(miner).Select(columns).UpdateColumns(miner).Error
	} else {
		err = o.Save(miner).Error
	}
	return
}

//Delete ...
func (miner *Miner) Delete(o *gorm.DB) (err error) {
	err = o.Delete(miner).Error
	return
}
