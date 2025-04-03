package middleware

import (
	"github.com/gin-gonic/gin"
	"server/new/app/logic/common/agent"
	"server/structs/Http/response"
)

func Is代理鉴权(权限名称 []int) gin.HandlerFunc {
	return func(c *gin.Context) {
		map_id_权限名称 := agent.L_agent.Q取全部代理功能ID_MAP(c)
		for i := range 权限名称 {
			if !agent.L_agent.Id功能权限检测(c, c.GetInt("Uid"), 权限名称[i]) {
				response.FailWithMessage("无["+map_id_权限名称[权限名称[i]]+"]权限,请联系上级代理授权.", c)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
