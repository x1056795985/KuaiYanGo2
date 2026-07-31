package ka

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	dbm "server/app/models/db"
	"server/app/service"
)

// B卡号_保存管理员编辑 在事务中同步卡号及卡号模式应用用户状态。
func B卡号_保存管理员编辑(c *gin.Context, 数据库 *gorm.DB, 卡号 dbm.DB_Ka) (旧卡号 dbm.DB_Ka, err error) {
	err = 数据库.Transaction(func(tx *gorm.DB) error {
		var 局_错误 error
		旧卡号, 局_错误 = service.NewKa(c, tx).Id取详情(卡号.Id)
		if 局_错误 != nil {
			return 局_错误
		}
		局_更新 := map[string]interface{}{
			"Status": 卡号.Status, "Num": 卡号.Num, "AdminNote": 卡号.AdminNote, "AgentNote": 卡号.AgentNote,
			"VipTime": 卡号.VipTime, "InviteCount": 卡号.InviteCount, "RMb": 卡号.RMb,
			"VipNumber": 卡号.VipNumber, "UserClassId": 卡号.UserClassId,
			"NoUserClass": 卡号.NoUserClass, "MaxOnline": 卡号.MaxOnline,
		}
		if 局_错误 = tx.Model(dbm.DB_Ka{}).Where("Id = ?", 旧卡号.Id).Updates(&局_更新).Error; 局_错误 != nil {
			return 局_错误
		}
		if service.NewAppInfo(c, tx).App是否为卡号(旧卡号.AppId) {
			局_错误 = tx.Model(dbm.DB_AppUser{}).Table("db_AppUser_"+strconv.Itoa(旧卡号.AppId)).Where("Id = ?", 旧卡号.Id).Update("Status", 卡号.Status).Error
		}
		return 局_错误
	})
	return
}
