package appUser

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	dbm "server/app/models/db"
	"server/app/service"
)

// P应用用户_批量新增 在事务中写入账号或卡号及其应用用户。
func P应用用户_批量新增(数据库 *gorm.DB, appId int, appType int, 用户数组 []dbm.DB_User, 卡号数组 []dbm.DB_Ka, 应用用户数组 []dbm.DB_AppUser) error {
	return 数据库.Transaction(func(tx *gorm.DB) error {
		if appType < 3 {
			if 局_错误 := tx.Model(dbm.DB_User{}).CreateInBatches(&用户数组, len(用户数组)).Error; 局_错误 != nil {
				return 局_错误
			}
			for 局_索引, 局_用户 := range 用户数组 {
				应用用户数组[局_索引].Uid = 局_用户.Id
			}
		} else {
			if 局_错误 := tx.Model(dbm.DB_Ka{}).CreateInBatches(&卡号数组, len(卡号数组)).Error; 局_错误 != nil {
				return 局_错误
			}
			for 局_索引, 局_卡号 := range 卡号数组 {
				应用用户数组[局_索引].Uid = 局_卡号.Id
			}
		}
		return tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(appId)).CreateInBatches(&应用用户数组, len(应用用户数组)).Error
	})
}

// B应用用户_保存管理员编辑 在事务中同步应用用户、卡号状态、在线归属和用户配置。
func B应用用户_保存管理员编辑(c *gin.Context, 数据库 *gorm.DB, appId int, 应用用户 dbm.DB_AppUser, 用户配置 []dbm.DB_UserConfig) (旧用户 dbm.DB_AppUser, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		if 局_错误 := tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(appId)).Where("Id = ?", 应用用户.Id).Take(&旧用户).Error; 局_错误 != nil {
			return 局_错误
		}
		if 局_错误 := tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(appId)).Where("Id = ?", 应用用户.Id).Updates(map[string]interface{}{
			"Status": 应用用户.Status, "Key": 应用用户.Key, "VipTime": 应用用户.VipTime,
			"VipNumber": 应用用户.VipNumber, "Note": 应用用户.Note, "MaxOnline": 应用用户.MaxOnline,
			"UserClassId": 应用用户.UserClassId, "AgentUid": 应用用户.AgentUid,
		}).Error; 局_错误 != nil {
			return 局_错误
		}
		if service.NewAppInfo(c, tx).App是否为卡号(appId) {
			if 局_错误 := tx.Model(dbm.DB_Ka{}).Where("Id = ?", 应用用户.Id).Update("Status", 应用用户.Status).Error; 局_错误 != nil {
				return 局_错误
			}
		}
		if 旧用户.AgentUid != 应用用户.AgentUid {
			if 局_错误 := tx.Model(dbm.DB_LinksToken{}).Where("LoginAppid = ?", appId).Where("Uid = ?", 旧用户.Uid).Update("AgentUid", 应用用户.AgentUid).Error; 局_错误 != nil {
				return 局_错误
			}
		}
		for _, 局_配置 := range 用户配置 {
			if 局_错误 := service.NewUserConfig(c, tx).Z置值(appId, 应用用户.Uid, 局_配置.Name, 局_配置.Value); 局_错误 != nil {
				return 局_错误
			}
		}
		return nil
	})
	return
}

// B应用用户_保存代理编辑 在事务中同步应用用户和关联卡号状态。
func B应用用户_保存代理编辑(c *gin.Context, 数据库 *gorm.DB, appId int, id int, status int, key string, appType int) error {
	return 数据库.Transaction(func(tx *gorm.DB) error {
		if _, 局_错误 := service.NewAppUser(c, tx, appId).Update(id, map[string]interface{}{"Status": status, "Key": key}); 局_错误 != nil {
			return 局_错误
		}
		if appType == 2 || appType == 3 {
			_, 局_错误 := service.NewKa(c, tx).Update(id, map[string]interface{}{"Status": status})
			return 局_错误
		}
		return nil
	})
}
