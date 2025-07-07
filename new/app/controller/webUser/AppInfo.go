package controller

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/controller/Common/response"
	"server/new/app/service"
	DB "server/structs/db"
)

type AppInfo struct {
	Common.Common
}

func NewAppInfoController() *AppInfo {
	return &AppInfo{}
}

func (C *AppInfo) GetAppBaseInfo(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"AppId" binging:"required,min=10000"` // Appid 必填`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		appInfo DB.DB_AppInfo
	}{}
	var err error
	tx := *global.GVA_DB

	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")
		return
	}
	data := gin.H{
		"AppId":            info.appInfo.AppId,
		"AppType":          info.appInfo.AppType,
		"AppName":          info.appInfo.AppName,
		"AppWeb":           info.appInfo.AppWeb,
		"Status":           info.appInfo.Status,
		"AppStatusMessage": info.appInfo.AppStatusMessage,
	}
	response.OkWithData(c, data)
	return
}

// 修改app排序
func (C *AppInfo) GetAppGongGao(c *gin.Context) {
	var 请求 struct {
		AppId int `json:"AppId" binging:"required,min=10000"` // Appid 必填`
	}
	//解析失败
	if !C.ToJSON(c, &请求) {
		return
	}
	var info = struct {
		appInfo DB.DB_AppInfo
	}{}
	var err error
	tx := *global.GVA_DB

	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")
		return
	}
	data := gin.H{
		"AppGongGao": info.appInfo.AppGongGao,
	}
	response.OkWithData(c, data)
	return
}
