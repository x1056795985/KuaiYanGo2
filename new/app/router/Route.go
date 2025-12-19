package router

import (
	"github.com/gin-gonic/gin"
	"server/new/app/router/admin"
	"server/new/app/router/agent"
	"server/new/app/router/webApi2"
	"server/new/app/router/webSocket"
	"server/new/app/router/webUser"
)

// 总路由
var AllRouter = new(RouterGroup)

type RouterGroup struct {
}

func RouterInit(gin *gin.RouterGroup) *gin.RouterGroup {

	Router := gin //返回路由实例
	// 跨域，如需跨域可以打开下面的注释
	router := admin.AllRouter{}
	router.InitAdminRouter(Router) //初始化管理员路由

	routerAgent := agent.AllRouter{}
	routerAgent.InitAgentRouter(Router) //初始化Agent路由

	routerWebApi := webApi2.AllRouter{}
	routerWebApi.InitWebApiRouter(Router) //初始化WEBAPi路由

	routerWebUser := webUser.AllRouter{}
	routerWebUser.InitWebUserRouter(Router) //初始化WEBUser路由

	routerWebSocket := webSocket.AllRouter{}
	routerWebSocket.InitWebSocketRouter(Router) //初始化WebSocket路由
	return gin
}
