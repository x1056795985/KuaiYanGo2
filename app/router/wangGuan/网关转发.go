package wangGuan

// 网关转发路由组:鉴权 → 签发JWT → 透明转发(HTTP/WS)
// 全程零日志零限制;错误统一空body+状态码+x-wg-error头
// 客户端请求:/wangGuan/*  头(或query)带 Token + WangGuanId
// 注:对外URL路径与代码标识符统一为 wangGuan

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	wangGuan "server/app/logic/wangGuan"
)

type AllRouter struct{}

func (r *AllRouter) InitWangGuanRouter(router *gin.RouterGroup) {
	//gin1.9: 空路径或"/"路径与catch-all共存注册会panic,精确路径必须注册在分组之外
	//裸路径(/wangGuan)精确注册,避免301重定向丢失POST body与请求头
	router.Any("wangGuan", W网关处理)
	局_分组 := router.Group("wangGuan")
	局_分组.Any("/*path", W网关处理)
}

// W网关处理 网关唯一入口:鉴权失败不转发直接返回,成功则透传
func W网关处理(c *gin.Context) {
	//取WangGuanId:header优先,query兜底(浏览器WebSocket场景)
	局_网关id文本 := strings.TrimSpace(c.GetHeader("WangGuanId"))
	if 局_网关id文本 == "" {
		局_网关id文本 = strings.TrimSpace(c.Query("WangGuanId"))
	}
	局_网关id, 局_err := strconv.Atoi(局_网关id文本)
	if 局_err != nil || 局_网关id <= 0 {
		wangGuan.Q响应网关错误(c.Writer, http.StatusNotFound, "网关不存在")
		return
	}
	局_网关配置, 局_存在 := wangGuan.L_网关.Q取网关(局_网关id)
	if !局_存在 {
		wangGuan.Q响应网关错误(c.Writer, http.StatusNotFound, "网关不存在")
		return
	}
	if 局_网关配置.Status != 1 {
		wangGuan.Q响应网关错误(c.Writer, http.StatusForbidden, "网关已禁用")
		return
	}

	//取Token:header优先,query兜底
	局_token := strings.TrimSpace(c.GetHeader("Token"))
	if 局_token == "" {
		局_token = strings.TrimSpace(c.Query("Token"))
	}
	局_鉴权信息, 局_鉴权错误 := wangGuan.L_网关.Q鉴权Token(局_token)
	if 局_鉴权错误 != nil {
		wangGuan.Q响应网关错误(c.Writer, http.StatusUnauthorized, 局_鉴权错误.Error())
		return
	}

	局_ip := wangGuan.L_网关.Q取真实ip(c)
	局_jwt, 局_签发错误 := wangGuan.L_网关.Q签发JWT(局_网关配置, 局_鉴权信息, 局_ip)
	if 局_签发错误 != nil {
		wangGuan.Q响应网关错误(c.Writer, http.StatusInternalServerError, 局_签发错误.Error())
		return
	}

	//透明转发:HTTP与WebSocket同一入口, ReverseProxy自动处理WS升级透传
	wangGuan.L_网关.Q执行转发(c, 局_网关配置, 局_jwt, 局_ip)
}
