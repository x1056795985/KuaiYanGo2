package webApi

import (
	"github.com/gin-gonic/gin"
	controller "server/new/app/controller/webApi"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebApiRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	webApiRouter := router.Group("webApi")
	//回调支付
	局_NotifyController := controller.NewPayNotifyController()
	{
		//http://anyueyinluo.w1.luyouxia.net/payNotify/240426223543000001
		webApiRouter.POST("/payNotify/:order", 局_NotifyController.PayNotify)   //通用支付回调
		webApiRouter.POST("/payNotify2/:order", 局_NotifyController.PayNotify2) //通用退款回调
	}

	//兼容旧版小叮当 因为服务器已配置
	webApiRouter = router.Group("WebApi")
	{
		webApiRouter.POST("/PayXiaoDingDangNotify", 局_NotifyController.PayNotify) //小叮当回调
	}
}
