package service

import (
	"errors"
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

// 增
func (s *AppUser) Create(info *DB.DB_AppUser) (row int64, err error) {
	//创建会自动重新赋值info.AppId为新插入的数据AppId
	tx := s.db.Model(DB.DB_AppUser{}).Create(info)
	return tx.RowsAffected, tx.Error
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
func (s *AppUser) Update2(where map[string]interface{}, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_" + strconv.Itoa(s.appid)).Where(where).Updates(&数据)
	return tx.RowsAffected, tx.Error
}
func (s *AppUser) UpdateUid(Uid int, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Uid = ?", Uid).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// Id点数增减 可能减少到0以下 ,增加无限制
func (s *AppUser) Id点数增减_批量(Id []int, 增减值 int64, is增加 bool) (err error) {
	//因为无符号 转换正负数 比较乱容易精度错误,所以 增加一个 Is增加 形参 判断是增加还是减少
	if len(Id) == 0 {
		return errors.New("Id数组不能为空")
	}
	if 增减值 == 0 {
		//增减0 直接成功
		return nil
	}
	sql := "VipTime - ?"
	if is增加 {
		sql = "VipTime + ?"
	}
	err = s.db.Model(DB.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(s.appid)).Where("Id IN ?", Id).Update("VipTime", gorm.Expr(sql, 增减值)).Error
	return err

}
