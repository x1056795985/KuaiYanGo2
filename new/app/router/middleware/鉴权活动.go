package middleware

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common/response"
	controller "server/new/app/controller/webUser"
	"server/new/app/models/constant"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

func isPromotionActive(promotionType int) gin.HandlerFunc {
	return func(c *gin.Context) {
		var err error
		var info = struct {
			appInfo             DB.DB_AppInfo
			likeInfo            DB.DB_LinksToken
			AppPromotionConfigs []dbm.DB_AppPromotionConfig
			AppPromotionConfig  dbm.DB_AppPromotionConfig
		}{}
		controller.Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
		tx := *global.GVA_DB

		info.AppPromotionConfigs, err = service.NewAppPromotionConfig(c, &tx).Infos(
			map[string]interface{}{
				"appId":         info.appInfo.AppId,
				"promotionType": promotionType,
			})
		if err != nil && err.Error() != "record not found" {
			response.FailWithMessage(c, err.Error())
			c.Abort()
			return
		}

		局_当前时间戳 := time.Now().Unix()
		for i := range info.AppPromotionConfigs {
			if info.AppPromotionConfigs[i].StartTime < 局_当前时间戳 && info.AppPromotionConfigs[i].EndTime > 局_当前时间戳 {
				info.AppPromotionConfig = info.AppPromotionConfigs[i]
				break
			}
		}
		if info.AppPromotionConfig.Id == 0 {
			局_活动类型 := ""
			switch promotionType {
			case constant.H活动类型_cps:
				局_活动类型 = "CPS"
			case constant.H活动类型_签到:
				局_活动类型 = "签到"
			}
			response.FailWithMessage(c, "未开启"+局_活动类型+"活动")
			c.Abort()
			return
		}

		c.Set("AppPromotionConfig", info.AppPromotionConfig)
		c.Next()
	}
}

// 使用方式
func Is存在活动_cps() gin.HandlerFunc {
	return isPromotionActive(constant.H活动类型_cps)
}

func Is存在活动_签到() gin.HandlerFunc {
	return isPromotionActive(constant.H活动类型_签到)
}
