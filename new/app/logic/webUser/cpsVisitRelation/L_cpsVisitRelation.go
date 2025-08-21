package L_CpsVisitRelation

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

var L_CpsVisitRelation appUser

func init() {
	L_CpsVisitRelation = appUser{}

}

type appUser struct {
}

// 四舍五入  索引越小,代理级别越靠下  代理专用
func (j *appUser) S设置邀请人(c *gin.Context, AppId, 邀请人, 被邀请人 int) (err error) {
	var info struct {
		AppInfo DB.DB_AppInfo
		上级      dbm.DB_CpsVisitRelation
		上上级     dbm.DB_CpsVisitRelation
		插入数据    []dbm.DB_CpsVisitRelation
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}
	//查询是否拥有邀请人   如果已设置过,需要删除,因为有唯一索引
	info.上级, info.上上级, err = service.NewCpsVisitRelation(c, tx).Q取归属邀请人(AppId, 被邀请人)
	if err != nil {
		return err
	}
	继续完成逻辑
	info.AppInfo, err = service.NewAppInfo(c, tx).Info(AppId)
	// 卡号模式的   处理同步ka冻结
	err = tx.Transaction(func(tx2 *gorm.DB) error {
		//先修改软件用户
		ret := tx2.Table(表名_AppUser).Where("AgentUid=?", c.GetInt("Uid")).Where("Id IN ? ", id).Update("Status", Status)
		if ret.Error != nil {
			return ret.Error
		}

		if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 && len(id) > 0 {
			// 子查询获取所有软件用户的Uid 在修改卡号 子查询内限制 代理uid
			err = tx2.Model(&DB.DB_Ka{}).Where("Id IN (?)", tx.Table(表名_AppUser).Select("Uid").Where("AgentUid=?", c.GetInt("Uid")).Where("Id IN (?)", id)).Update("Status", Status).Error
		}
		return err
	})

	return
}
