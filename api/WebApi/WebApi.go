package WebApi

import (
	"github.com/gin-gonic/gin"
)

var J集_UserAPi路由 = map[string]路由信息{
	"TaskPoolGetTask":  {"任务处理获取", R任务池_任务处理获取},
	"TaskPoolSetTask":  {"任务处理返回", R任务池_任务处理返回},
	"GetAppUpDataJson": {"取App最新下载地址", Q取App最新下载地址},
	"GetKaInfo":        {"取卡号详细信息", Get卡号详细信息},
	"NewKa":            {"新制卡号", New制新卡},
	"RunJs":            {"运行公共js函数", RunJs},
}

type 路由信息 struct {
	Z中文名  string
	Z指向函数 func(*gin.Context)
}
