package Panel

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"server/global"
	"server/structs/Http/response"
	"server/utils"
	"time"
)

type Api struct{}

// ReloadSystem
// @Tags      System
// @Summary   重启系统
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{msg=string}  "重启系统"
// @Router    /system/reloadSystem [post]
func (s *Api) ReloadSystem(c *gin.Context) {
	err := utils.Reload()
	//err := utils.Reload热重启(global.GVA_CONFIG.Port)

	if err != nil {
		global.GVA_LOG.Error("重启失败!", zap.Error(err))
		response.FailWithMessage("重启失败:"+err.Error(), c)
		return
	}

	response.OkWithMessage("重启系统成功,请稍等大约十秒左右", c)
}

// GetServerInfo
// @Tags      System
// @Summary   获取服务器信息
// @Security  ApiKeyAuth
// @Produce   application/json
// @Success   200  {object}  response.Response{data=map[string]interface{},msg=string}  "获取服务器信息"
// @Router    /system/getServerInfo [post]
func (s *Api) GetServerInfo(c *gin.Context) {

	server, err := GetServerInfo()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"server": server}, "获取成功", c)
}

func GetServerInfo() (server *utils.Server, err error) {
	var s utils.Server
	s.Os = utils.InitOS()
	if s.Cpu, err = utils.InitCPU(); err != nil {
		global.GVA_LOG.Error("func utils.InitCPU() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Ram, err = utils.InitRAM(); err != nil {
		global.GVA_LOG.Error("func utils.InitRAM() Failed", zap.String("err", err.Error()))
		return &s, err
	}
	if s.Disk, err = utils.InitDisk(); err != nil {
		global.GVA_LOG.Error("func utils.InitDisk() Failed", zap.String("err", err.Error()))
		return &s, err
	}

	return &s, nil
}

// 停止系统
func (s *Api) StopSystem(c *gin.Context) {

	response.FailWithMessage("已操作停止系统,再见", c)
	go func() {
		time.Sleep(2 * time.Second) //延迟2秒在关闭主程序,让这个请求返回,给前端一个反馈
		fmt.Printf("已操作停止系统,再见")
		//先关闭端口 解除占用
		global.GVA_Gin.Shutdown(context.Background()) //这句话可以停止侦听关闭端口
		// 退出当前进程
		os.Exit(0)
	}()
	return

}
