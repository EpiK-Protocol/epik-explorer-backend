package epik

import "gorm.io/gorm"

//RegisterModels ...
func RegisterModels(o *gorm.DB) {
	o.AutoMigrate(
		Miner{},
		ProfitRecord{},
		Admin{},
		AdminLog{},
	)
}
