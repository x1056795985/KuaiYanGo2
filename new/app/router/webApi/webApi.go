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
		webApiRouter.POST("/PayHuPiJiaoNotify", 局_NotifyController.PayHuPiJiaoNotify)

		//http://anyueyinluo.w1.luyouxia.net/payNotify/240426223543000001
		webApiRouter.POST("/payNotify/:order", 局_NotifyController.PayNotify) //通用回调
		webApiRouter.POST("/PayXiaoDingDangNotify", 局_NotifyController.PayNotify)
	}
}
