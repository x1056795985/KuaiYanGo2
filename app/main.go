package app

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songzhibin97/gkit/cache/local_cache"

	"server/app/bootstrap"
	"server/app/global"
	_ "server/app/logic"
	kuaiYanLogic "server/app/logic/admin/kuaiYan"
	"server/app/logic/common/cron"
	"server/app/logic/common/cron/functions"
	"server/app/logic/common/setting"
	"server/app/logic/webSocket"
	"server/app/router"
	"server/app/router/userSafetyApi"
)

// Z主程序_运行 初始化应用基础设施并启动 HTTP 服务。
func Z主程序_运行() {
	defer 主程序_捕获致命错误()

	global.GVA_Viper = bootstrap.P配置_初始化()
	主程序_初始化缓存()
	局_关闭数据库 := 主程序_初始化数据库()
	defer 局_关闭数据库()
	主程序_初始化定时任务()
	bootstrap.InitWebSocket()
	bootstrap.F服务_运行(router.InitRouters())
}

// Z主程序_停止 停止定时任务和 HTTP 服务。
func Z主程序_停止() {
	cron.Q全_定时任务.Cron.Stop()
	if global.GVA_Gin != nil {
		_ = global.GVA_Gin.Shutdown(context.Background())
	}
}

func 主程序_捕获致命错误() {
	if 局_异常 := recover(); 局_异常 != nil {
		局_上报错误 := fmt.Sprintln("全局捕获错误:\n", 局_异常, "\n堆栈信息:\n", string(debug.Stack()))
		fmt.Println("发生致命错误:", 局_上报错误)
		global.Q快验.Z置新用户消息(2, 局_上报错误)
	}
}

func 主程序_初始化缓存() {
	global.H缓存 = local_cache.NewCache(local_cache.SetDefaultExpire(24 * time.Hour))
}

func 主程序_初始化数据库() func() {
	global.GVA_DB, _ = bootstrap.InitGormMysql()
	if global.GVA_DB == nil {
		global.GVA_LOG.Println("数据库连接失败,等待输入数据库信息")
		return func() {}
	}

	局_上下文 := &gin.Context{}
	bootstrap.InitDbTables(局_上下文)
	局_数据库连接, 局_错误 := global.GVA_DB.DB()
	if 局_错误 != nil {
		global.GVA_LOG.Println("获取数据库连接失败:" + 局_错误.Error())
		return func() {}
	}
	userSafetyApi.J集_UserAPi路由_加密.G更新md5APi名称(setting.Q系统设置().Y用户API加密盐)
	webSocket.L_webSocket.D断开所有连接()
	return func() {
		if 局_关闭错误 := 局_数据库连接.Close(); 局_关闭错误 != nil {
			global.GVA_LOG.Println("关闭数据库连接失败:" + 局_关闭错误.Error())
		}
	}
}

func 主程序_初始化定时任务() {
	cron.Q全_定时任务 = cron.D定时任务{}
	cron.Q全_定时任务.Init()

	局_错误 := cron.Q全_定时任务.T添加本机任务("快验心跳", "0 */5 * * * *", kuaiYanLogic.K快验_心跳)
	if 局_错误 != nil {
		global.GVA_LOG.Println("添加快验心跳定时任务失败:" + 局_错误.Error())
	}
	局_错误 = cron.Q全_定时任务.T添加本机任务("刷新数据库定时任务", "0 */1 * * * *", functions.S刷新数据库定时任务2)
	if 局_错误 != nil {
		global.GVA_LOG.Println("添加刷新数据库定时任务失败:" + 局_错误.Error())
	}
	_ = functions.S刷新数据库定时任务(true)
	functions.D定时任务_统计初始化日活月活(&gin.Context{})
	cron.Q全_定时任务.Cron.Start()
}
