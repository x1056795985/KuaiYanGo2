package webApi2

import (
	"github.com/gin-gonic/gin"
	controller "server/new/app/controller/webApi"
	"server/new/app/router/middleware"
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
		webApiRouter.GET("/payNotify/:order", 局_NotifyController.PayNotify)    //通用支付回调
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

	// ========== 以下为从 api/WebApi 迁移的路由 ==========

	//任务池
	局_TaskPoolController := controller.NewTaskPoolWebApiController()
	{
		J集_UserAPi路由2["/TaskPoolGetTask"] = 路由信息{Z中文名: "任务处理获取", Z指向函数: 局_TaskPoolController.TaskPoolGetTask}
		J集_UserAPi路由2["/TaskPoolSetTask"] = 路由信息{Z中文名: "任务处理返回", Z指向函数: 局_TaskPoolController.TaskPoolSetTask}
		J集_UserAPi路由2["/TaskPoolNewData"] = 路由信息{Z中文名: "任务池_任务创建", Z指向函数: 局_TaskPoolController.TaskPoolNewData}
		J集_UserAPi路由2["/TaskPoolGetData"] = 路由信息{Z中文名: "任务池_任务查询", Z指向函数: 局_TaskPoolController.TaskPoolGetData}
	}
	//公共函数(RunJs/:JsName 需要单独注册路径参数路由)
	局_RunJs路由 := 局_TaskPoolController.RunJs
	局_RunJs2路由 := 局_TaskPoolController.RunJs2
	_ = 局_RunJs路由
	_ = 局_RunJs2路由
	{
		J集_UserAPi路由2["/RunJs"] = 路由信息{Z中文名: "运行公共函数", Z指向函数: 局_RunJs路由}
	}
	//卡号相关
	局_KaController := controller.NewKaWebApiController()
	{
		J集_UserAPi路由2["/GetKaInfo"] = 路由信息{Z中文名: "取卡号详细信息", Z指向函数: 局_KaController.GetKaInfo}
		J集_UserAPi路由2["/NewKa"] = 路由信息{Z中文名: "新制卡号", Z指向函数: 局_KaController.NewKa}
	}
	//支付订单
	局_PayController := controller.NewPayWebApiController()
	{
		J集_UserAPi路由2["/Pay/GetPayOrderStatus"] = 路由信息{Z中文名: "取支付订单状态", Z指向函数: 局_PayController.GetPayOrderStatus}
	}
	//公共变量
	局_PublicDataController := controller.NewPublicDataWebApiController()
	{
		J集_UserAPi路由2["/GetPublicData"] = 路由信息{Z中文名: "取公共变量", Z指向函数: 局_PublicDataController.GetPublicData}
		J集_UserAPi路由2["/GetPublicDataLen"] = 路由信息{Z中文名: "取公共变量行数", Z指向函数: 局_PublicDataController.GetPublicDataLen}
		J集_UserAPi路由2["/SetPublicData"] = 路由信息{Z中文名: "置公共变量", Z指向函数: 局_PublicDataController.SetPublicData}
	}
	//应用信息
	局_AppInfoController := controller.NewAppInfoWebApiController()
	{
		J集_UserAPi路由2["/GetAppUpDataJson"] = 路由信息{Z中文名: "取App最新下载地址", Z指向函数: 局_AppInfoController.GetAppUpDataJson}
	}

	webApiRouter = router.Group("WebApi")
	webApiRouter.Use(middleware.IsWebApiHost())
	webApiRouter.Use(middleware.IsTokenWebApi()) ///鉴权中间件 检查 token  单独优先处理
	{
		for 键名, 键值 := range J集_UserAPi路由2 {
			webApiRouter.GET(键名, 键值.Z指向函数)
			webApiRouter.POST(键名, 键值.Z指向函数)
		}
		// RunJs/:JsName 路径参数路由需要单独注册
		webApiRouter.GET("/RunJs/:JsName", 局_RunJs2路由)
		webApiRouter.POST("/RunJs/:JsName", 局_RunJs2路由)
	}

}
