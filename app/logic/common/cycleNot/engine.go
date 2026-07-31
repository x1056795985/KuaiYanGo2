// Package cycleNot 提供脚本引擎初始化函数注入，避免包循环依赖。
package cycleNot

import (
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"

	dbm "server/app/models/db"
)

// J脚本引擎_初始化函数 定义脚本运行时初始化函数签名。
type J脚本引擎_初始化函数 func(
	c *gin.Context,
	appInfo *dbm.DB_AppInfo,
	online *dbm.DB_LinksToken,
	publicJS *dbm.DB_PublicJs,
) *goja.Runtime

// Q全_脚本引擎初始化 保存跨包注入的脚本引擎初始化实现。
var Q全_脚本引擎初始化 J脚本引擎_初始化函数

// J脚本引擎_设置初始化函数 注入脚本引擎初始化实现。
func J脚本引擎_设置初始化函数(initializer J脚本引擎_初始化函数) {
	Q全_脚本引擎初始化 = initializer
}
