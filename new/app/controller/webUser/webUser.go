package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/new/app/global"
	dbm "server/new/app/models/db"
	"server/new/app/models/old/response"
	"server/new/app/service"
)

func Y用户数据信息还原(c *gin.Context, 在线信息 *dbm.DB_LinksToken, AppInfo *dbm.DB_AppInfo) {
	局_临时通用, _ := c.Get("局_在线信息")
	*在线信息 = 局_临时通用.(dbm.DB_LinksToken)
	if AppInfo != nil {
		db := *global.GVA_DB
		局_临时通用, _ = service.NewAppInfo(c, &db).Info(D到整数(在线信息.AppIdEx))
		*AppInfo = 局_临时通用.(dbm.DB_AppInfo)
	}
}
func Y限账号模式应用(c *gin.Context, AppInfo *dbm.DB_AppInfo) bool {
	if AppInfo.AppType == 1 || AppInfo.AppType == 2 {
		return true
	}
	response.FailWithMessage("仅限账号模式应用调用", c)
	return false
}
