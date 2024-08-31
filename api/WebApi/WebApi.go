package WebApi

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/service"
	DB "server/structs/db"
)

var J集_UserAPi路由 = map[string]路由信息{
	"TaskPoolGetTask":       {"任务处理获取", R任务池_任务处理获取},
	"TaskPoolSetTask":       {"任务处理返回", R任务池_任务处理返回},
	"TaskPoolNewData":       {"任务池_任务创建", R任务池_任务创建},
	"TaskPoolGetData":       {"任务池_任务查询", R任务池_任务查询},
	"GetAppUpDataJson":      {"取App最新下载地址", Q取App最新下载地址},
	"GetKaInfo":             {"取卡号详细信息", Get卡号详细信息},
	"NewKa":                 {"新制卡号", New制新卡},
	"RunJs":                 {"运行公共js函数", RunJs},
	"Pay/GetPayOrderStatus": {"取支付订单状态", Q取支付订单状态},
	"GetPublicData":         {"取公共变量", Q取公共变量},
	"GetPublicDataLen":      {"取公共变量行数", Q取队列长度},
	"SetPublicData":         {"置公共变量", Z置公共变量},
}

type 路由信息 struct {
	Z中文名  string
	Z指向函数 func(*gin.Context)
}

func Y用户数据信息还原(c *gin.Context, AppInfo *DB.DB_AppInfo, 在线信息 *DB.DB_LinksToken) {
	db := *global.GVA_DB
	*AppInfo, _ = service.NewAppInfo(c, &db).Info(3)
	局_临时通用, _ := c.Get("DB_LinksToken")
	*在线信息 = 局_临时通用.(DB.DB_LinksToken)
	return
}
