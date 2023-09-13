package main

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"os"
	"runtime/debug"
	"server/Service/Ser_Init"
	"server/api/middleware"
	"server/core"
	"server/global"
)

/*设置包自动管理*/
//go:generate go env -w GO111MODULE=on

/*设置包管理下载地址*/
//go:generate go env -w GOPROXY=https://goproxy.cn,direct

/*设置包自动整理信息到go.mod内*/
//go:generate go mod tidy

/*设置包自动根据go.mod下载*/
//go:generate go mod download

func main() {
	defer func() {
		if err := recover(); err != nil {
			局_上报错误 := fmt.Sprintln("全局捕获错误:\n", err, "\n堆栈信息:\n", string(debug.Stack()))
			fmt.Println("发生致命错误:", 局_上报错误)
			global.Q快验.Z置新用户消息(2, 局_上报错误)

		}
	}()

	global.GVA_Viper = core.InitViper() //初始化配置读写器 和全局配置结构变量GVA_config
	middleware.G更新哈希APi名称(global.GVA_CONFIG.X系统设置.Y用户API加密盐)
	global.GVA_LOG = core.InitZap()          // 初始化zap日志记录器
	zap.ReplaceGlobals(global.GVA_LOG)       //替换系统的log记录器 为zap的全局日志记录器 方便统一管理
	global.GVA_DB = Ser_Init.InitGormMysql() // gorm连接数据库  Gorm参考资料https://www.cnblogs.com/davis12/p/16365213.html

	if global.GVA_DB != nil { //如果数据库不为空
		Ser_Init.InitDbTables() // 如果数据库连接成功就初始化表  //不在这里了,只能由 InitMysql 初始化

		// 程序结束前关闭数据库链接
		db, _ := global.GVA_DB.DB()
		defer db.Close() //延迟关闭程序结束前关闭表
	} else {
		global.GVA_LOG.Info(fmt.Sprintf("数据库连接失败,等待输入数据库信息"))
	}
	core.InitCron定时任务()

	core.RunWindowsServer() //启动web服务器  先启动 不然无法自验证

}

func STOP() {
	global.Cron定时任务.Cron.Stop()
	//先关闭端口 解除占用
	global.GVA_Gin.Shutdown(context.Background()) //这句话可以停止侦听关闭端口
	// 退出当前进程
	os.Exit(0)

}
