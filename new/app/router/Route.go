package router

import (
	"github.com/gin-gonic/gin"
	"server/new/app/router/admin"
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

	return gin
}
