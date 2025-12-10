package cps

import (
	"github.com/gin-gonic/gin"
	"server/global"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

var L_cps cps

func init() {
	L_cps = cps{}

}

type cps struct {
}

// 要考虑线程安全,使用数据库 订单号唯一索引锁实现线程安全
func (j *cps) Q开启中cps活动(c *gin.Context, AppId int) (cpsInfo dbm.DB_CpsInfo) {
	var err error
	var info struct {
		邀请人信息              DB.DB_User
		AppPromotionConfig dbm.DB_AppPromotionConfig
	}

	db := *global.GVA_DB

	//获取该应用是否已开启了cps 可能会有多个符合时间的配置信息 只获取第一个
	数组_AppPromotionConfigs, err := service.NewAppPromotionConfig(c, &db).Infos(
		map[string]interface{}{
			"appId":         AppId,
			"promotionType": 1,
		})
	if err != nil && err.Error() != "record not found" {
		//未配置活动
		return
	}
	局_当前时间戳 := time.Now().Unix()
	for i := range 数组_AppPromotionConfigs {
		if 数组_AppPromotionConfigs[i].StartTime < 局_当前时间戳 && 数组_AppPromotionConfigs[i].EndTime > 局_当前时间戳 {
			info.AppPromotionConfig = 数组_AppPromotionConfigs[i]
			break
		}
	}
	if info.AppPromotionConfig.Id == 0 {
		//无符合时间的活动
		return
	}
	cpsInfo, err = service.NewCpsInfo(c, &db).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil {
		return
	}
	return
}
