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
		appInfo     DB.DB_AppInfo
		appInfoUser dbm.DB_AppInfoWebUser
	}{}
	var err error
	tx := *global.GVA_DB

	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")
		return
	}
	if info.appInfoUser, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId); err != nil {
		info.appInfoUser = dbm.DB_AppInfoWebUser{
			Status: 2,
		}
	}
	data := gin.H{
		"appId":            info.appInfo.AppId,
		"appType":          info.appInfo.AppType,
		"appName":          info.appInfo.AppName,
		"appWeb":           info.appInfo.AppWeb,
		"status":           info.appInfo.Status,
		"appStatusMessage": info.appInfo.AppStatusMessage,
		"webUserStatus":    info.appInfoUser.Status,
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
