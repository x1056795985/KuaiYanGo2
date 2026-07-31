package user

import (
	"strconv"

	"gorm.io/gorm"

	dbm "server/app/models/db"
)

// S用户_删除非代理 在事务中删除账号及所有账号模式应用用户。
func S用户_删除非代理(数据库 *gorm.DB, 用户ids []int) (影响行数 int64, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		局_结果 := tx.Model(dbm.DB_User{}).Where("Id IN ?", 用户ids).Delete("")
		if 局_结果.Error != nil {
			return 局_结果.Error
		}
		影响行数 = 局_结果.RowsAffected
		var 局_appIds []int
		if 局_错误 := tx.Model(dbm.DB_AppInfo{}).Select("AppId").Where("AppType IN ?", []int{1, 2}).Scan(&局_appIds).Error; 局_错误 != nil {
			return 局_错误
		}
		for _, 局_appId := range 局_appIds {
			if 局_错误 := tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(局_appId)).Where("Uid IN ?", 用户ids).Delete("").Error; 局_错误 != nil {
				return 局_错误
			}
		}
		return nil
	})
	return
}
