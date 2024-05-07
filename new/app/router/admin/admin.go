package admin

import (
	"github.com/gin-gonic/gin"
	"server/api/middleware"
	controller "server/new/app/controller/admin"
)

type AllRouter struct {
}

func (r *AllRouter) InitAdminRouter(router *gin.RouterGroup) {
	// 跨域，如需跨域可以打开下面的注释
	adminRouter := router.Group("Admin")
	adminRouter.Use(middleware.IsTokenAdmin()) //加载中间件

	局_Setting := controller.NewSettingController()
	{
		adminRouter.POST("setting/info", 局_Setting.Info)
	}

	//用户日志
	局_LogUserMsg := controller.NewLogUserMsgController()
	{
		adminRouter.POST("LogUserMsg/DeleteDuplicateMsg", 局_LogUserMsg.S删除重复消息)
	}
	//黑名单管理
	局_Blacklist := controller.NewBlacklistController()
	{
		adminRouter.POST("Blacklist/GetList", 局_Blacklist.GetList)
		adminRouter.POST("Blacklist/Create", 局_Blacklist.Create)
		adminRouter.POST("Blacklist/Delete", 局_Blacklist.Delete)
		adminRouter.POST("Blacklist/Update", 局_Blacklist.Update)
		adminRouter.POST("Blacklist/Info", 局_Blacklist.Info)
		adminRouter.POST("Blacklist/DeleteBatch", 局_Blacklist.DeleteBatch) //批量删除 1全部

	}
	//定时任务管理
	局_Cron := controller.NewCronController()
	{
		adminRouter.POST("Cron/GetList", 局_Cron.GetList)
		adminRouter.POST("Cron/Create", 局_Cron.Create)
		adminRouter.POST("Cron/Delete", 局_Cron.Delete)
		adminRouter.POST("Cron/Update", 局_Cron.Update)
		adminRouter.POST("Cron/Info", 局_Cron.Info)
		adminRouter.POST("Cron/DeleteBatch", 局_Cron.DeleteBatch)   //批量删除 1全部
		adminRouter.POST("Cron/UpdateStatus", 局_Cron.UpdateStatus) //更新状态
		adminRouter.POST("Cron/TestRunId", 局_Cron.Z执行)             //测试运行一次
	}

	//定时任务日志
	局_CronLog := controller.NewCronLogController()
	{
		adminRouter.POST("LogCronTask/GetList", 局_CronLog.GetList)
		adminRouter.POST("LogCronTask/Delete", 局_CronLog.Delete)
		adminRouter.POST("LogCronTask/DeleteBatch", 局_CronLog.DeleteBatch) //批量删除 1全部

	}
	//仪表盘
	局_chart := controller.NewChartController()
	{
		adminRouter.POST("Panel/getConsumptionRanking", 局_chart.Q取余额消费排行榜)
		adminRouter.POST("Panel/getRmbIncreaseRanking", 局_chart.Q取余额增长排行榜)
		adminRouter.POST("Panel/getNumberIncreaseRanking", 局_chart.Q取积分消费排行榜)
		adminRouter.POST("Panel/gaodeWeather", 局_chart.G高德取天气)
	}

}
