package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	DB "server/structs/db"
	"strconv"
)

type AppUser struct {
	db    *gorm.DB
	c     *gin.Context
	appid int
}

// NewAppUser 创建 AppUser 实例
func NewAppUser(c *gin.Context, db *gorm.DB, appId int) *AppUser {
	return &AppUser{
		db:    db,
		appid: appId,
		c:     c,
	}
}

func (s *AppUser) Info(id int) (info DB.DB_AppUser, err error) {
	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id = ?", id).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}
func (s *AppUser) InfoUid(Uid int) (info DB.DB_AppUser, err error) {

	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Uid = ?", Uid).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

func (s *AppUser) Update(Id int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id = ?", Id).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
func (s *AppUser) UpdateUid(Uid int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Uid = ?", Uid).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
