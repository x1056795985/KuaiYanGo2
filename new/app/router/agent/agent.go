package agent

import (
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	"server/global"
	controller "server/new/app/controller/agent"
	mid2 "server/new/app/router/middleware"
	DB "server/structs/db"
)

type AllRouter struct {
}

func (r *AllRouter) InitAgentRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	agentRouter := router.Group(global.GVA_Viper.GetString("代理入口"))
	agentRouter.Use(middleware.IsTokenAgent()) //加载中间件

	局_Setting := controller.NewSettingController()
	{
		agentRouter.POST("Setting/getInfoPay", 局_Setting.GetPayInfo)
		agentRouter.POST("Setting/setInfoPay", 局_Setting.SetPayInfo)
		agentRouter.POST("Setting/setBaseInfo", 局_Setting.SetBaseInfo)
		agentRouter.POST("Setting/getBaseInfo", 局_Setting.GetBaseInfo)
		//代理云配置
		agentRouter.POST("Setting/getAgentUserConfig", 局_Setting.GetAgentUserConfig)
		agentRouter.POST("Setting/delAgentUserConfig", 局_Setting.DelAgentUserConfig)
		agentRouter.POST("Setting/newAgentUserConfig", 局_Setting.NewAgentUserConfig)
		agentRouter.POST("Setting/saveAgentUserConfig", 局_Setting.SaveAgentUserConfig)
	}
	局_AppUser := controller.NewAppUserController()
	{
		agentRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_查看归属软件用户})).POST("AppUser/GetList", 局_AppUser.GetList)        // 获取列表
		agentRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_查看归属软件用户})).POST("AppUser/GetInfo", 局_AppUser.GetAppUserInfo) // 获取详细信息
		agentRouter.POST("AppUser/SetStatus", 局_AppUser.Set修改状态)                                                       // 修改状态
		agentRouter.POST("AppUser/SaveUser", 局_AppUser.Save用户信息)
		agentRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_修改用户密码})).POST("AppUser/SetPassUser", 局_AppUser.Set用户密码)
	}
	// 为需要鉴权的路由单独创建子组

	局_UserClass := controller.NewUserClassController()
	{
		agentRouter.POST("UserClass/GetIdNameList", 局_UserClass.GetIdNameList) // 获取列表
	}

	//代理账号
	局_AgentUser := controller.NewAgentUserController()
	{
		agentRouter.POST("Agent/GetKaSalesStatistics", 局_AgentUser.GetKaSalesStatistics)
	}
	局_KaClass := controller.NewKaClassController()
	{
		agentRouter.POST("KaClass/GetList", 局_KaClass.GetList)
	}

	局_KaClassUpPrice := controller.NewKaClassUpPriceController()
	{
		agentRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_卡类调价})).POST("KaClassUpPrice/Save", 局_KaClassUpPrice.Save)
		agentRouter.Group("", mid2.Is代理鉴权([]int{DB.D代理功能_卡类调价})).POST("KaClassUpPrice/Delete", 局_KaClassUpPrice.Delete)
	}

}
