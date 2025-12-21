// engine.go - JS引擎初始化器，用于解决循环引用问题
package cycleNot

import (
	"github.com/dop251/goja"   // JavaScript运行时引擎
	"github.com/gin-gonic/gin" // HTTP框架，用于处理请求上下文
	DB "server/structs/db"     // 数据库结构体定义
)

// JsEngineInitializer - JS引擎初始化函数类型定义
// 用于初始化JavaScript运行环境，传入请求上下文、应用信息、在线用户信息和公共JS配置
type JS引擎初始化_用户 func(
	c *gin.Context,             // Gin框架的请求上下文
	AppInfo *DB.DB_AppInfo,     // 应用基本信息
	在线信息 *DB.DB_LinksToken, // 用户在线状态信息
	局_PublicJs *DB.DB_PublicJs) *goja.Runtime // 公共JS脚本配置

// GlobalJsEngineInit - 全局JS引擎初始化函数变量
// 用于存储实际的JS引擎初始化实现，避免包间直接依赖
var GlobalJsEngineInit JS引擎初始化_用户

// SetJsEngineInitializer - 设置全局JS引擎初始化函数
// 通过此函数注入具体的JS引擎初始化实现
func SetJsEngineInitializer(init JS引擎初始化_用户) {
	GlobalJsEngineInit = init
}
