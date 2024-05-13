package appUser

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"server/global"
	"server/new/app/service"
	DB "server/structs/db"
	"strconv"
)

var L_appUser appUser

func init() {
	L_appUser = appUser{}

}

type appUser struct {
}

// 四舍五入  索引越小,代理级别越靠下
func (j *appUser) Z置状态_同步卡号修改(c *gin.Context, AppId int, id []int, Status int) (err error) {

	var 表名_AppUser = "db_AppUser_" + strconv.Itoa(AppId)
	var info struct {
		AppInfo DB.DB_AppInfo
	}
	var tx *gorm.DB
	if tempObj, ok := c.Get("tx"); ok {
		tx = tempObj.(*gorm.DB)
	} else {
		db := *global.GVA_DB
		tx = &db
	}

	info.AppInfo, err = service.NewAppInfo(c, tx).Info(AppId)
	// 卡号模式的   处理同步ka冻结
	err = tx.Transaction(func(tx2 *gorm.DB) error {
		//先修改软件用户
		err = tx2.Table(表名_AppUser).Where("Id IN ? ", id).Update("Status", Status).Error
		if err != nil {
			return err
		}
		if info.AppInfo.AppType == 3 || info.AppInfo.AppType == 4 {
			// 子查询获取所有软件用户的Uid 在修改卡号
			err = tx.Debug().Model(&DB.DB_Ka{}).Where("Id IN (?)", tx.Table(表名_AppUser).Select("Uid").Where("Id IN (?)", id)).Update("Status", Status).Error
		}
		return err
	})

	return
}
