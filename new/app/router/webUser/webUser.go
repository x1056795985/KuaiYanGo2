package webUser

import (
	"github.com/gin-gonic/gin"
	"net/http"
	controller "server/new/app/controller/webUser"
	"server/new/app/router/middleware"
	"server/new/app/web/webUser"
	VueWebUserAssets "server/new/app/web/webUser/assets"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebUserRouter(router *gin.RouterGroup) {

	//打包静态VueAdmin文件============================
	html := webUser.NewHtmlHandler()
	Router根Admin := router.Group("user/:appId") //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	Router根Admin.GET("/", html.Index)
	Router根Admin2 := router.Group("user/0") //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	Router根Admin2.StaticFS("/assets", http.FS(VueWebUserAssets.Assets))

	// 跨域，如需跨域可以打开下面的注释
	//global.GVA_Viper.GetString("管理入口")
	adminRouter := router.Group("userApi")
	// 需要鉴权才能访问的接口
	adminRouter.Use(middleware.IsDbConn())
	adminRouter.Use(middleware.IsTokenWebUser())
	// 无需鉴权就可以访问的接口
	局_AppInfo := controller.NewAppInfoController()
	{
		adminRouter.POST("app/getAppBaseInfo", 局_AppInfo.GetAppBaseInfo)
		adminRouter.POST("app/getAppGongGao", 局_AppInfo.GetAppGongGao)
	}
	局_Base := controller.NewBaseController()
	{
		adminRouter.POST("base/loginUserOrKa", 局_Base.LoginUserOrKa)
	}
	adminRouter.GET("base/loginKey", 局_Base.LoginKey) //这个是get请求,单独处理

	局_user := controller.NewUserController()
	{
		adminRouter.POST("user/newUserInfo", 局_user.NewUserInfo)
		adminRouter.POST("user/getPwSendSms", 局_user.GetPwSendSms)
		adminRouter.POST("user/getInfo", 局_user.GetInfo)
		adminRouter.POST("user/smsCodeSetPassWord", 局_user.SmsCodeSetPassWord)
		adminRouter.POST("user/logout", 局_user.Logout)
		adminRouter.POST("user/setBaseInfo", 局_user.SetBaseInfo)
		adminRouter.POST("user/sendSms", 局_user.SendSms)
	}

	局_appUser := controller.NewAppUserController()
	{
		adminRouter.POST("appUser/getAppUserInfo", 局_appUser.GetAppUserInfo)
	}

	局_ka := controller.NewKaController()
	{
		adminRouter.POST("ka/useKa", 局_ka.UseKa)
	}

	局_pay := controller.NewPayController()
	{
		adminRouter.POST("pay/getPayStatus", 局_pay.GetPayStatus)
		adminRouter.POST("pay/getPayOrderStatus", 局_pay.GetPayOrderStatus)
		adminRouter.POST("pay/getPayKaList", 局_pay.GetPayKaList)
		adminRouter.POST("pay/payKaUsa", 局_pay.PayKaUsa)
	}

	局_cps := controller.NewCpsController()
	{
		adminRouter.POST("cps/setVisitRelation", 局_cps.SetVisitRelation)
	}

}
