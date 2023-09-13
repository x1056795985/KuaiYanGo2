package WebApi

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_AppInfo"
	"server/structs/Http/response"
	"strconv"
)

func Q取App最新下载地址(c *gin.Context) {

	局_AppID, _ := strconv.Atoi(c.DefaultQuery("AppId", ""))
	if 局_AppID == 0 || !Ser_AppInfo.AppId是否存在(局_AppID) {
		response.FailWithMessage("应用不存在", c)
		return
	}

	response.OkWithDetailed(Ser_AppInfo.App取App最新下载地址Json(局_AppID), "获取成功", c)
	return
}
