package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"os"
	"server/Service/Ser_Chare"
	"server/global"
	"server/new/app/controller/Common"
	"server/structs/Http/response"
	"server/utils"
	"time"
)

type Panel struct {
	Common.Common
}

func NewPanelController() *Panel {
	return &Panel{}
}

// GetServerInfo 获取服务器信息
func (p *Panel) GetServerInfo(c *gin.Context) {
	server, err := getServerInfo()
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败", c)
		return
	}
	response.OkWithDetailed(gin.H{"server": server}, "获取成功", c)
}

func getServerInfo() (server *utils.Server, err error) {
	var s utils.Server
	s.Os = utils.InitOS()
	if s.Cpu, err = utils.InitCPU(); err != nil {
		return &s, err
	}
	if s.Ram, err = utils.InitRAM(); err != nil {
		return &s, err
	}
	if s.Disk, err = utils.InitDisk(); err != nil {
		return &s, err
	}
	return &s, nil
}

// ReloadSystem 重启系统
func (p *Panel) ReloadSystem(c *gin.Context) {
	err := utils.Reload()
	if err != nil {
		global.GVA_LOG.Error("重启失败!", zap.Error(err))
		response.FailWithMessage("重启失败:"+err.Error(), c)
		return
	}
	response.OkWithMessage("重启系统成功,请稍等大约十秒左右", c)
}

// StopSystem 停止系统
func (p *Panel) StopSystem(c *gin.Context) {
	response.FailWithMessage("已操作停止系统,再见", c)
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Printf("已操作停止系统,再见")
		global.GVA_Gin.Shutdown(context.Background())
		os.Exit(0)
	}()
	return
}

// 图表分析页相关方法

// ChartLinksUser 在线统计
func (p *Panel) ChartLinksUser(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get在线用户统计(c), "获取成功", c)
}

// ChartLinksUserIPCity 在线用户IP地图分布统计
func (p *Panel) ChartLinksUserIPCity(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get在线用户Ip地图分布统计(c), "获取成功", c)
}

// ChartLinksUserLoginTime 统计用户日活月活
func (p *Panel) ChartLinksUserLoginTime(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get统计用户日活月活(c), "获取成功", c)
}

// ChartEveryHourLinksCount 统计分时段在线总数
func (p *Panel) ChartEveryHourLinksCount(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get统计分时段在线总数(c), "获取成功", c)
}

// ChartAppUserClass 应用用户类型统计
func (p *Panel) ChartAppUserClass(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户类型统计(c), "获取成功", c)
}

// ChartUser 用户账号登录注册统计
func (p *Panel) ChartUser(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get用户账号登录注册统计(c), "获取成功", c)
}

// ChartRMBAddSub 余额充值消费统计
func (p *Panel) ChartRMBAddSub(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get余额充值消费统计(c), "获取成功", c)
}

// ChartVipNumberAddSub 积分点数消费统计
func (p *Panel) ChartVipNumberAddSub(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get积分点数消费统计(c), "获取成功", c)
}

// ChartAppUser 应用用户统计
func (p *Panel) ChartAppUser(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户统计(c), "获取成功", c)
}

// ChartAppKa 卡号列表统计应用卡可用已用
func (p *Panel) ChartAppKa(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计应用卡可用已用(c), "获取成功", c)
}

// ChartAppKaClass 卡号列表统计应用卡类可用已用
func (p *Panel) ChartAppKaClass(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计应用卡类可用已用(c), "获取成功", c)
}

// ChartKaRegister 卡号列表统计制卡
func (p *Panel) ChartKaRegister(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计制卡(c), "获取成功", c)
}

// ChartAppUserRegister 应用用户账号注册统计
func (p *Panel) ChartAppUserRegister(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户账号注册统计(c), "获取成功", c)
}

// ChartAgentLevel 代理组织架构图
func (p *Panel) ChartAgentLevel(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get代理组织架构图(c, -1), "获取成功", c)
}

// ChartTidTaskData 任务池任务Id分析
func (p *Panel) ChartTidTaskData(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get任务池任务Id分析(c), "获取成功", c)
}
