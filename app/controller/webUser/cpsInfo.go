package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/controller/Common/response"
	"server/app/global"
	dbm "server/app/models/db"
	"server/app/service"
)

type CpsInfo struct {
	Common.Common
}

func NewCpsInfoController() *CpsInfo {
	return &CpsInfo{}
}

func (C *CpsInfo) Info(c *gin.Context) {
	var err error
	var info = struct {
		appInfo             dbm.DB_AppInfo
		likeInfo            dbm.DB_LinksToken
		CpsInfo             dbm.DB_CpsInfo
		AppPromotionConfigs []dbm.DB_AppPromotionConfig
		AppPromotionConfig  dbm.DB_AppPromotionConfig
	}{}
	Y用户数据信息还原(c, &info.likeInfo, &info.appInfo)
	tx := *global.GVA_DB
	if 临时, ok := c.Get("AppPromotionConfig"); ok {
		info.AppPromotionConfig = 临时.(dbm.DB_AppPromotionConfig)
	} else {
		response.FailWithMessage(c, "未开启cps活动")
		return
	}

	info.CpsInfo, err = service.NewCpsInfo(c, &tx).Info(info.AppPromotionConfig.TypeAssociatedId)
	if err != nil && err.Error() != "record not found" {
		response.FailWithMessage(c, err.Error())
		return
	}
	//把info.CpsInfo 结构体转为 gin.H 并添加新字段 BindGiveKaClassName
	局_临时 := struct {
		dbm.DB_CpsInfo
		BindGiveKaClassName string `json:"bindGiveKaClassName"`
	}{info.CpsInfo, ""}
	KaClass, err := service.NewKaClass(c, &tx).Info(info.CpsInfo.BindGiveKaClassId)
	if err == nil {
		局_临时.BindGiveKaClassName = KaClass.Name
	}

	response.OkWithData(c, gin.H{
		"appPromotionConfig": info.AppPromotionConfig,
		"cps":                局_临时,
	})
	return
}
