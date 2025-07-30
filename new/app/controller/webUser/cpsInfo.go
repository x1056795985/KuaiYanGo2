package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
	"time"
)

type Cps struct {
	Common.Common
}

func NewCpsController() *Cps {
	return &Cps{}
}

func (C *Cps) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo             DB.DB_AppInfo
		likeInfo            DB.DB_LinksToken
		cps                 dbm.DB_CpsInfo
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	//获取该应用是否已开启了cps 可能会有多个符合时间的配置信息 只获取第一个
	info.AppPromotionConfigs, err = service.NewAppPromotionConfig(c, &tx).Infos(
		map[string]interface{}{
			"appId":         info.appInfo.AppId,
			"promotionType": 1,
		})
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
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
		response.FailWithMessage(c, "未开启CPS活动")
		return
	}
	info.cps, err = service.NewCpsInfo(c, &tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"appPromotionConfig": info.AppPromotionConfig,
		"cps":                info.cps,
	})
	return
}
