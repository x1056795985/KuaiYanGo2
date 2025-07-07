package webUser

import (
	"github.com/gin-gonic/gin"
	controller "server/new/app/controller/webUser"
	"server/new/app/router/middleware"
)

type AllRouter struct {
}

func (r *AllRouter) InitWebUserRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	//global.GVA_Viper.GetString("管理入口")
	adminRouter := router.Group("userApi")
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
	局_user := controller.NewUserController()
	{
		adminRouter.POST("user/newUserInfo", 局_user.NewUserInfo)
		adminRouter.POST("user/getPwSendSms", 局_user.GetPwSendSms)
		adminRouter.POST("user/smsCodeSetPassWord", 局_user.SmsCodeSetPassWord)
	}

	// 需要鉴权才能访问的接口
	adminRouter.Use(middleware.IsTokenWebUser())
	局_appUser := controller.NewAppUserController()
	{
		adminRouter.POST("appUser/getAppUserInfo", 局_appUser.GetAppUserInfo)
	}

}
