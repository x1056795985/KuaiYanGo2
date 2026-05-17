package controller

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
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
	if 局_AppID == 0 || !Ser_AppInfo.AppId是否存在(局_AppID) {
		response.FailWithMessage("应用不存在", c)
		return
	}

	response.OkWithDetailed(Ser_AppInfo.App取App最新下载地址Json(局_AppID), "获取成功", c)
	return
}
