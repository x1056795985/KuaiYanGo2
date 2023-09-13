package middleware

import (
	"github.com/gin-gonic/gin"
	"server/structs/Http/response"
)

// 部分URL演示模式需要拦截,不然有捣乱的
func Is演示模式url拦截() gin.HandlerFunc {
	return func(c *gin.Context) {
		response.FailWithMessage("演示状态不可操作,请配置到服务器", c)
		c.Abort()
		return
	}
}
