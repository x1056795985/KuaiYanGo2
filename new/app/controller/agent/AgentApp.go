package controller

import (
	"github.com/gin-gonic/gin"
)

type AgentApp struct{}

func NewAgentAppController() *AgentApp {
	return &AgentApp{}
}

// GetAppIdNameList 取代理可操作应用列表
func (A *AgentApp) GetAppIdNameList(c *gin.Context) {
	var 局_Ka AgentKa
	局_Ka.GetAppIdNameList(c)
}

type Agent应用键值对 struct {
	AppId   int    `json:"appId"`
	AppName string `json:"appName"`
}

type Agent应用列表响应 struct {
	Map   map[string]string `json:"map"`
	Array []Agent应用键值对      `json:"array"`
}
