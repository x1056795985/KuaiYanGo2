package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	dbm "server/new/app/models/db"
	"server/new/app/service"
	DB "server/structs/db"
)

type CheckInInfo struct {
	Common.Common
}

func NewCheckInInfoController() *CheckInInfo {
	return &CheckInInfo{}
}

func (C *CheckInInfo) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo             DB.DB_AppInfo
		likeInfo            DB.DB_LinksToken
		CheckInInfo         dbm.DB_CheckInInfo
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB

	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启签到活动")
		return
	}

	info.CheckInInfo, err = service.NewCheckInInfo(c, &tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}

	response.OkWithData(c, gin.H{
		"appPromotionConfig": info.AppPromotionConfig,
		"checkIn":            info.CheckInInfo,
	})
	return
}
