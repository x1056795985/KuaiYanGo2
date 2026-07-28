package service

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	dbm "server/new/app/models/db"
	"time"
)

type UserConfig struct {
	db *gorm.DB
	c  *gin.Context
}

// NewUserConfig 创建 UserConfig 实例
func NewUserConfig(c *gin.Context, db *gorm.DB) *UserConfig {
	return &UserConfig{
		db: db,
		c:  c,
	}
}

// Name是否存在 用户配置Name是否存在
func (s *UserConfig) Name是否存在(AppId, Uid int, Name string) bool {
	var Count int64
	s.db.Model(dbm.DB_UserConfig{}).Select("1").Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).Take(&Count)
	return Count > 0

}

// Infos 按条件查询多条用户配置
func (s *UserConfig) Infos(where map[string]interface{}) (info []dbm.DB_UserConfig, err error) {
	tx := s.db.Model(dbm.DB_UserConfig{}).Where(where).Find(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// Info2 按条件查询单条用户配置
func (s *UserConfig) Info2(where map[string]interface{}) (info dbm.DB_UserConfig, err error) {
	tx := s.db.Model(dbm.DB_UserConfig{}).Where(where).First(&info)
	if tx.Error != nil {
		err = tx.Error
	}
	return
}

// Update 按条件更新用户配置
func (s *UserConfig) Update(where map[string]interface{}, 数据 map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_UserConfig{}).Where(where).Updates(&数据)
	return tx.RowsAffected, tx.Error
}

// Create 创建用户配置
func (s *UserConfig) Create(info dbm.DB_UserConfig) (row int64, err error) {
	tx := s.db.Model(dbm.DB_UserConfig{}).Create(&info)
	return tx.RowsAffected, tx.Error
}

// Delete2 按条件删除用户配置
func (s *UserConfig) Delete2(where map[string]interface{}) (row int64, err error) {
	tx := s.db.Model(dbm.DB_UserConfig{}).Where(where).Delete("")
	return tx.RowsAffected, tx.Error
}

// Q取值 按AppId和Uid取用户配置值
func (s *UserConfig) Q取值(AppId, Uid int, Name string) string {
	var value string = ""
	s.db.Model(dbm.DB_UserConfig{}).Select("Value").Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).First(&value)
	return value
}

// Q取值2 按AppId和Uid取用户配置完整信息
func (s *UserConfig) Q取值2(AppId, Uid int, Name string) (dbm.DB_UserConfig, error) {
	var value dbm.DB_UserConfig
	err := s.db.Model(dbm.DB_UserConfig{}).Where("AppId=?", AppId).Where("Uid=?", Uid).Where("Name=?", Name).First(&value).Error
	return value, err
}

// Z置值 设置用户配置值(不存在则创建)
func (s *UserConfig) Z置值(Appid, Uid int, Name string, Value string) error {
	db := s.db.Model(dbm.DB_UserConfig{})
	var err error
	if s.Name是否存在(Appid, Uid, Name) {
		updates := map[string]interface{}{
			"Value":      Value,
			"UpdateTime": time.Now().Unix(),
		}
		err = db.Where("AppId=?", Appid).Where("Uid=?", Uid).Where("Name=?", Name).Updates(updates).Error
	} else {
		var 局_User = ""
		appInfo := AppInfo{db: s.db, c: s.c}
		if appInfo.App是否为卡号(Appid) {
			ka := Ka{db: s.db, c: s.c}
			局_User = ka.Id取卡号(Uid)
		} else {
			u := User{db: s.db, c: s.c}
			局_User = u.Id取User(Uid)
		}
		var 局_用户配置 = dbm.DB_UserConfig{AppId: Appid, Uid: Uid, Name: Name, Value: Value, Time: time.Now().Unix(), UpdateTime: time.Now().Unix(), User: 局_User}
		err = db.Create(&局_用户配置).Error
	}

	return err
}

// Z置值2 批量修改用户配置值(按AppId和Name)
func (s *UserConfig) Z置值2(PublicData dbm.DB_UserConfig) error {
	return s.db.Model(dbm.DB_UserConfig{}).Select("Value", "IsVip", "Note").Omit("Type", "AppId", "Name").Where("AppId=?", PublicData.AppId).Where("Name=?", PublicData.Name).Updates(PublicData).Error
}

// C创建 创建用户配置
func (s *UserConfig) C创建(PublicData dbm.DB_UserConfig) error {
	err := s.db.Model(dbm.DB_UserConfig{}).Create(&PublicData).Error
	return err
}

// P批量取值 按Appid批量取用户配置
func (s *UserConfig) P批量取值(Appid int) []dbm.DB_UserConfig {
	var value []dbm.DB_UserConfig
	s.db.Model(dbm.DB_UserConfig{}).Where("AppId=?", Appid).Find(&value)
	return value
}

// P批量置值 批量保存用户配置
func (s *UserConfig) P批量置值(DB_PublicData []dbm.DB_UserConfig) error {

	return s.db.Model(dbm.DB_UserConfig{}).Save(DB_PublicData).Error
}

// P批量置值2 批量置值(按Appid, Uid数组, Name)
func (s *UserConfig) P批量置值2(Appid int, Uid []int, Name string, Value string) error {
	if Value == "" {
		return s.db.Model(dbm.DB_UserConfig{}).Where("AppId=?", Appid).Where("Uid IN ?", Uid).Where("Name=?", Name).Delete("").Error
	}

	var 局_数据 []dbm.DB_UserConfig
	局_数据 = make([]dbm.DB_UserConfig, len(Uid))
	for i, v := range Uid {
		局_数据[i].AppId = Appid
		局_数据[i].Uid = v
		局_数据[i].Name = Name
		局_数据[i].Value = Value
	}

	return s.db.Model(dbm.DB_UserConfig{}).Save(局_数据).Error
}
