package epik

import (
	"gorm.io/gorm"
)

//Admin ...
type Admin struct {
	UserName string
	Password string
	gorm.Model
}

//AdminLog ...
type AdminLog struct {
	AdminID int64
	Action  string
	gorm.Model
}

//GetAdmin ...
func GetAdmin(id int64, o *gorm.DB) (admin *Admin, err error) {
	admin = &Admin{}
	err = o.Model(Admin{}).First(admin, id).Error
	return
}

//GetAdminByUserName ...
func GetAdminByUserName(userName string, o *gorm.DB) (admin *Admin, err error) {
	admin = &Admin{}
	err = o.Model(Admin{}).Where("user_name = ?", userName).First(admin).Error
	return
}

//Create ...
func (admin *Admin) Create(o *gorm.DB) (err error) {
	return o.Create(admin).Error
}

//Delete ...
func (admin *Admin) Delete(o *gorm.DB) (err error) {
	err = o.Delete(admin).Error
	return
}
