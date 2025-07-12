package middleware

import (
	"github.com/gin-gonic/gin"
	"server/global"
	"server/new/app/controller/Common/response"
)

// 检测数据库连接
func IsDbConn() gin.HandlerFunc {
	return func(c *gin.Context) {
		if global.GVA_DB == nil {
			response.FailWithMessage(c, "数据库未连接,请联系管理员")
			c.Abort()
			return
		}
		c.Next()
	}
}
