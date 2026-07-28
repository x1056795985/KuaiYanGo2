package controller

import (
	"github.com/gin-gonic/gin"
	"server/new/app/controller/Common"
	"server/new/app/global"
	"server/new/app/models/old/response"
	"server/new/app/service"
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
	if 局_AppID == 0 || !service.NewAppInfo(c, global.GVA_DB).AppId是否存在(局_AppID) {
		response.FailWithMessage("应用不存在", c)
		return
	}

	response.OkWithDetailed(service.NewAppInfo(c, global.GVA_DB).App取App最新下载地址Json(局_AppID), "获取成功", c)
	return
}
