package api

import (
	"fmt"
	"time"

	"github.com/EpiK-Protocol/epik-explorer-backend/epik"
	"github.com/EpiK-Protocol/epik-explorer-backend/storage"
	"github.com/EpiK-Protocol/epik-explorer-backend/utils"
	"github.com/gin-gonic/gin"
)

//SetAdminAPI ...
func SetAdminAPI(e *gin.Engine) {
	g := e.Group("admin", tokenConfirm)
	e.POST("admin/login", adminLogin)
	e.POST("admin/regist", adminRegist)
	g.GET("miner/list", adminMinerList)
	g.POST("miner/confirm", adminMinerConfirm)
	g.POST("miner/reject", adminMinerReject)
	g.GET("profit/list", adminProfitList)
	g.POST("profit/delete", adminProfitDelete)
	g.POST("profit/done", adminProfitDone)
	g.POST("profit/cleanpending", adminProfitCleanPending)
	g.POST("profit/caculate", adminProfitCaculate)
}

const passwordKey = "epik-explorer"

func encryptPassword(password string) (encrypt string) {
	return utils.HMACSHA256(password, passwordKey)
}

func adminRegist(c *gin.Context) {
	req := &struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	admin := &epik.Admin{
		UserName: req.UserName,
		Password: encryptPassword(req.Password),
	}
	err := admin.Create(storage.DB)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
}

func adminLogin(c *gin.Context) {
	req := &struct {
		UserName string `json:"user_name"`
		Password string `json:"password"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	admin, err := epik.GetAdminByUserName(req.UserName, storage.DB)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	if encryptPassword(req.Password) != admin.Password {
		responseJSON(c, errUnauthorized)
		return
	}
	token := epik.CreateToken(int64(admin.ID), "admin", time.Hour)
	responseJSON(c, errOK, "token", token)
}

func adminMinerList(c *gin.Context) {
	page := ParsePage(c)
	status := c.Query("status")
	weixin := c.Query("weixin")
	id := c.Query("id")
	o := storage.DB
	miners := []*epik.Miner{}
	var total int64
	o = o.Model(epik.Miner{})
	if !isEmpty(status) {
		o = o.Where("status = ?", status)
	}
	if !isEmpty(weixin) {
		o = o.Where("wei_xin = ?", weixin)
	}
	if !isEmpty(id) {
		o = o.Where("id = ?", id)
	}
	err := o.Count(&total).Limit(page.Size).Offset(page.Offset).Find(&miners).Error
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK, "list", miners, "total", total)
}

func adminMinerConfirm(c *gin.Context) {
	req := &struct {
		MinerID string `json:"miner_id"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	miner, err := epik.GetMiner(storage.DB, req.MinerID)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	miner.Status = epik.MinerStatusConfirmed
	err = miner.Update(storage.DB, "status")
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
	go func() {
		if miner.Airdrop <= 0 {
			txHash, err := SendEPK(epikWallet, miner.EpikAddress, "201.00")
			if err == nil {
				miner.Airdrop = 201
				fmt.Println(txHash)
				miner.Update(storage.DB, "airdrop")
			} else {
				fmt.Printf("send epk:%s\n", err)
			}
		}
	}()

}

func adminMinerReject(c *gin.Context) {
	req := &struct {
		MinerID string `json:"miner_id"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	miner, err := epik.GetMiner(storage.DB, req.MinerID)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	miner.Status = epik.MinerStatusRejected
	err = miner.Update(storage.DB, "status")
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
}

func adminProfitList(c *gin.Context) {
	page := ParsePage(c)
	status := c.Query("status")
	o := storage.DB
	records := []*epik.ProfitRecord{}
	var total int64
	o = o.Model(epik.ProfitRecord{})
	if !isEmpty(status) {
		o = o.Where("status = ?", status)
	}
	err := o.Model(epik.ProfitRecord{}).Count(&total).Limit(page.Size).Offset(page.Offset).Find(&records).Error
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK, "list", records, "total", total)
}

func adminProfitDelete(c *gin.Context) {
	req := &struct {
		RecordID int64 `json:"record_id"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	record, err := epik.GetProfitRecord(storage.DB, req.RecordID)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	err = record.Delete(storage.DB)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
}

func adminProfitCleanPending(c *gin.Context) {
	o := storage.DB
	err := o.Exec("DELETE FROM profit_record WHERE status = ?", epik.MinerStatusPending).Error
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
}

func adminProfitDone(c *gin.Context) {
	req := &struct {
		RecordID int64 `json:"record_id"`
	}{}
	if err := c.ShouldBindJSON(req); err != nil {
		responseJSON(c, clientError(err))
		return
	}
	record, err := epik.GetProfitRecord(storage.DB, req.RecordID)
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	record.Status = epik.MinerStatusConfirmed
	err = record.Update(storage.DB, "status")
	if err != nil {
		responseJSON(c, serverError(err))
		return
	}
	responseJSON(c, errOK)
}

func adminProfitCaculate(c *gin.Context) {
	responseJSON(c, errOK)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println(err)
			}
		}()
		err := GenTestnetMinerBonusByPledge()
		if err != nil {
			fmt.Println(err)
		}
	}()
}
