package agent

import (
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	controller "server/new/app/controller/agent"
)

type AllRouter struct {
}

func (r *AllRouter) InitAgentRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	adminRouter := router.Group("agent")
	adminRouter.Use(middleware.IsTokenAgent()) //加载中间件

	局_Setting := controller.NewSettingController()
	{
		adminRouter.POST("setting/getInfoPay", 局_Setting.GetPayInfo)
		adminRouter.POST("setting/setInfoPay", 局_Setting.SetPayInfo)
		adminRouter.POST("setting/setBaseInfo", 局_Setting.SetBaseInfo)
		adminRouter.POST("setting/getBaseInfo", 局_Setting.GetBaseInfo)
	}
}
