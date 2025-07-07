package controller

import (
	. "EFunc/utils"
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/service"
	DB "server/structs/db"
)

func Y用户数据信息还原(c *gin.Context, 在线信息 *DB.DB_LinksToken, AppInfo *DB.DB_AppInfo) {
	局_临时通用, _ := c.Get("DB_LinksToken")
	*在线信息 = 局_临时通用.(DB.DB_LinksToken)
	if AppInfo != nil {
		db := *global.GVA_DB
		局_临时通用, _ = service.NewAppInfo(c, &db).Info(D到整数(在线信息.Tab))
		*AppInfo = 局_临时通用.(DB.DB_AppInfo)
	}
}
