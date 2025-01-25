package webApi2

import (
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	controller "server/new/app/controller/webApi"
)

type 路由信息 struct {
	Z中文名  string
	Z指向函数 func(*gin.Context)
}

var J集_UserAPi路由2 = map[string]路由信息{}

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
	//上面的不用令牌就可以访问
	//下边的需要令牌才可以访问
	webApiRouter = router.Group("WebApi")
	webApiRouter.Use(middleware.IsWebApiHost())
	webApiRouter.Use(middleware.IsTokenWebApi()) ///鉴权中间件 检查 token  单独优先处理
	//云存储
	局_CloudStorageController := controller.NewCloudStorageController()
	{

		J集_UserAPi路由2["/CloudStorage/GetUploadToken"] = 路由信息{Z中文名: "云存储_取文件上传授权", Z指向函数: 局_CloudStorageController.GetUploadToken}
		J集_UserAPi路由2["/CloudStorage/GetDownloadUrl"] = 路由信息{Z中文名: "云存储_取外链", Z指向函数: 局_CloudStorageController.GetDownloadUrl}
	}

	webApiRouter = router.Group("WebApi")
	webApiRouter.Use(middleware.IsWebApiHost())
	webApiRouter.Use(middleware.IsTokenWebApi()) ///鉴权中间件 检查 token  单独优先处理
	{
		for 键名, 键值 := range J集_UserAPi路由2 {
			webApiRouter.GET(键名, 键值.Z指向函数)
			webApiRouter.POST(键名, 键值.Z指向函数)
		}
	}

}
