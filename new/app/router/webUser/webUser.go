package webUser

import (
	"github.com/gin-gonic/gin"
	"net/http"
	controller "server/new/app/controller/webUser"
	mid2 "server/new/app/router/middleware"
	"server/new/app/web/webUser"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebUserRouter(router *gin.RouterGroup) {

	//打包静态VueAdmin文件============================
	html := webUser.NewHtmlHandler()
	Router根Admin := router.Group("user/:appId") //127.0.0.1:18080/  这个后面第一个不需要 / 符号
	Router根Admin.GET("/", html.Index)
	//http://127.0.0.1:18888/user/10001/assets/index-BUyaaghm.css
	Router根Admin.GET("assets/*filepath", func(c *gin.Context) {
		c.FileFromFS("assets/"+c.Param("filepath"), http.FS(webUser.Assets))
	})
	// http://127.0.0.1:18888/user/10001/static/shilu-login/2.png
	Router根Admin.GET("static/*filepath", func(c *gin.Context) {
		c.FileFromFS("static/"+c.Param("filepath"), http.FS(webUser.Static))
	})

	// 跨域，如需跨域可以打开下面的注释
	//global.GVA_Viper.GetString("管理入口")
	根路由 := router.Group("userApi")
	// 需要鉴权才能访问的接口
	根路由.Use(mid2.IsDbConn())
	根路由.Use(mid2.IsTokenWebUser())
	// 无需鉴权就可以访问的接口 通过中间件 白名单控制
	局_AppInfo := controller.NewAppInfoController()
	{
		根路由.POST("app/getAppBaseInfo", 局_AppInfo.GetAppBaseInfo)
		根路由.POST("app/getAppGongGao", 局_AppInfo.GetAppGongGao)
	}
	局_Base := controller.NewBaseController()
	{
		根路由.POST("base/loginUserOrKa", 局_Base.LoginUserOrKa)
		根路由.POST("base/Captcha2", 局_Base.Captcha2)
	}
	根路由.GET("base/loginKey", 局_Base.LoginKey) //这个是get请求,单独处理

	局_user := controller.NewUserController()
	{
		根路由.POST("user/newUserInfo", 局_user.NewUserInfo)
		根路由.POST("user/getPwSendSms", 局_user.GetPwSendSms)
		根路由.POST("user/getInfo", 局_user.GetInfo)
		根路由.POST("user/smsCodeSetPassWord", 局_user.SmsCodeSetPassWord)
		根路由.POST("user/logout", 局_user.Logout)
		根路由.POST("user/setBaseInfo", 局_user.SetBaseInfo)
		根路由.POST("user/sendSms", 局_user.SendSms)
	}

	局_appUser := controller.NewAppUserController()
	{
		根路由.POST("appUser/getAppUserInfo", 局_appUser.GetAppUserInfo)
	}

	局_ka := controller.NewKaController()
	{
		根路由.POST("ka/useKa", 局_ka.UseKa)
	}

	局_pay := controller.NewPayController()
	{
		根路由.POST("pay/getPayStatus", 局_pay.GetPayStatus)
		根路由.POST("pay/getPayOrderStatus", 局_pay.GetPayOrderStatus)
		根路由.POST("pay/getPayKaList", 局_pay.GetPayKaList)
		根路由.POST("pay/payKaUsa", 局_pay.PayKaUsa)
	}

	局_AppPromotionConfig := controller.NewAppPromotionConfigController()
	{
		根路由.POST("appPromotionConfig/getList", 局_AppPromotionConfig.GetList)
	}

	局_cpsInfo := controller.NewCpsInfoController()
	{
		根路由.Group("", mid2.Is存在活动_cps()).POST("cps/info", 局_cpsInfo.Info)
	}
	局_shortUrl := controller.NewShortUrlController()
	router.GET("/c/:shortUrl", 局_shortUrl.Jump) //以c为二级目录区分短链模块
	{
		根路由.POST("shortUrl/info", 局_shortUrl.Info)
		根路由.POST("shortUrl/create", 局_shortUrl.Create)
	}

	局_cpsInvitingRelation := controller.NewCpsInvitingRelationController()
	{
		根路由.POST("cpsInvitingRelation/set", 局_cpsInvitingRelation.Set)
		根路由.POST("cpsInvitingRelation/get", 局_cpsInvitingRelation.Get)
		根路由.POST("cpsInvitingRelation/getInvitingList", 局_cpsInvitingRelation.GetInvitingList)
	}

	局_cpsUser := controller.NewCpsUserController()
	{
		根路由.Group("", mid2.Is存在活动_cps()).POST("cpsUser/info", 局_cpsUser.Info)
	}
	局_cpsPayOrder := controller.NewCpsPayOrderController()
	{
		根路由.Group("", mid2.Is存在活动_cps()).POST("cpsPayOrder/list", 局_cpsPayOrder.List)
	}

	局_CheckInUser := controller.NewCheckInUserController()
	{
		根路由.Group("", mid2.Is存在活动_签到()).POST("checkInUser/info", 局_CheckInUser.Info)
	}
	局_CheckInLog := controller.NewCheckInLogController()
	{
		根路由.Group("", mid2.Is存在活动_签到()).POST("checkInLog/create", 局_CheckInLog.Create)
	}
	局_CheckInScoreLog := controller.NewCheckInScoreLogController()
	{
		根路由.Group("", mid2.Is存在活动_签到()).POST("checkInScoreLog/getList", 局_CheckInScoreLog.GetList)
	}

	局_checkInInfo := controller.NewCheckInInfoController()
	{
		根路由.Group("", mid2.Is存在活动_签到()).POST("checkInInfo/info", 局_checkInInfo.Info)
	}
}
