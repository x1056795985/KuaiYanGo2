package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
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
		appInfo        DB.DB_AppInfo
		appInfoWebUser dbm.DB_AppInfoWebUser
		appInfoUser    dbm.DB_AppInfoWebUser
	}{}
	var err error
	tx := *global.GVA_DB

	if info.appInfoWebUser, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "未开启网页用户中心")
		return
	}
	if info.appInfoWebUser.Status != 1 {
		response.FailWithMessage(c, "未开启网页用户中心")
		return
	}
	if info.appInfo, err = service.NewAppInfo(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "AppId不存在")
		return
	}
	if info.appInfoUser, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId); err != nil {
		info.appInfoUser = dbm.DB_AppInfoWebUser{
			Status: 2,
		}
	}
	局_最新版本 := "1.0.0"
	局_可用版本 := W文本_分割文本(info.appInfo.AppVer, "\n")
	if len(局_可用版本) > 0 {
		局_最新版本 = 局_可用版本[0]
	}
	//如果下载地址url不是json则直接填写, 如果是json,则获取 data 第一个成员的 url地址
	info.appInfo.UrlDownload = info.appInfoUser.UrlDownload

	局_downloadUrl := Ser_AppInfo.App下载更新地址变量处理(info.appInfo)

	data := gin.H{
		"appId":            info.appInfo.AppId,
		"appType":          info.appInfo.AppType,
		"appName":          info.appInfo.AppName,
		"appWeb":           info.appInfo.AppWeb,
		"UrlHome":          info.appInfo.UrlHome,
		"status":           info.appInfo.Status,
		"appStatusMessage": info.appInfo.AppStatusMessage,
		"webUserStatus":    info.appInfoUser.Status,
		"appVer":           局_最新版本,
		"downloadUrl":      局_downloadUrl,
		"qrcodeUrl":        T图片_生成二维码base64(局_downloadUrl),
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
		appInfo        DB.DB_AppInfo
		appInfoWebUser dbm.DB_AppInfoWebUser
	}{}
	var err error
	tx := *global.GVA_DB
	if info.appInfoWebUser, err = service.NewAppInfoWebUser(c, &tx).Info(请求.AppId); err != nil {
		response.FailWithMessage(c, "未开启网页用户中心")
		return
	}
	if info.appInfoWebUser.Status != 1 {
		response.FailWithMessage(c, "未开启网页用户中心")
		return
	}
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
