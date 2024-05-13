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
	agentRouter := router.Group("agent")
	agentRouter.Use(middleware.IsTokenAgent()) //加载中间件

	局_Setting := controller.NewSettingController()
	{
		agentRouter.POST("setting/getInfoPay", 局_Setting.GetPayInfo)
		agentRouter.POST("setting/setInfoPay", 局_Setting.SetPayInfo)
		agentRouter.POST("setting/setBaseInfo", 局_Setting.SetBaseInfo)
		agentRouter.POST("setting/getBaseInfo", 局_Setting.GetBaseInfo)
	}
	局_AppUser := controller.NewAppUserController()
	{
		agentRouter.POST("AppUser/GetList", 局_AppUser.GetList)        // 获取列表
		agentRouter.POST("AppUser/GetInfo", 局_AppUser.GetAppUserInfo) // 获取详细信息
		agentRouter.POST("AppUser/SetStatus", 局_AppUser.Set修改状态)      // 修改状态
	}
	局_UserClass := controller.NewUserClassController()
	{
		agentRouter.POST("UserClass/GetIdNameList", 局_UserClass.GetIdNameList) // 获取列表
	}
}
