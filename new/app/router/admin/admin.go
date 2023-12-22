package admin

import (
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	controller "server/new/app/controller/admin"
)

type AllRouter struct {
}

func (r *AllRouter) InitAdminRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	adminRouter := router.Group("Admin")
	adminRouter.Use(middleware.IsTokenAdmin()) //加载中间件

	局_Setting := controller.NewSettingController()
	{
		adminRouter.POST("setting/info", 局_Setting.Info)
	}

	//用户日志
	局_LogUserMsg := controller.NewLogUserMsgController()
	{
		adminRouter.POST("LogUserMsg/DeleteDuplicateMsg", 局_LogUserMsg.S删除重复消息)
	}

}
