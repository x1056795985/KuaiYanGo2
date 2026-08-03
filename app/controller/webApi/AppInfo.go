package controller

import (
	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	"server/app/global"
	"server/app/logic/common/appInfo"
	dbm "server/app/models/db"
	"server/app/models/old/response"
	"server/app/service"
	"strconv"
)

type AppInfoWebApi struct {
	Common.Common
}

func NewAppInfoWebApiController() *AppInfoWebApi {
	return &AppInfoWebApi{}
}

// Q取App最新下载地址 取App最新下载地址
func (A *AppInfoWebApi) GetAppUpDataJson(c *gin.Context) {
	局_AppID, _ := strconv.Atoi(c.DefaultQuery("AppId", ""))
	if 局_AppID == 0 {
		response.FailWithMessage("应用不存在", c)
		return
	}
	db := *global.GVA_DB
	var 局_appInfo dbm.DB_AppInfo
	var err error
	局_appInfo, err = service.NewAppInfo(c, &db).Info(局_AppID)
	if err != nil {
		response.FailWithMessage("应用不存在", c)
		return
	}

	response.OkWithDetailed(appInfo.App下载更新地址变量处理(局_appInfo), "获取成功", c)
	return
}
