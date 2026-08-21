package controller

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"server/app/controller/Common"
	"server/app/global"
	"server/app/logic/admin/L_chart"
	"server/app/models/old/response"
	"server/app/monitoring"
	"server/app/utils"
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
		global.GVA_LOG.Println("获取失败!", err)
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
		global.GVA_LOG.Println("重启失败!", err)
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
	response.OkWithDetailed(L_chart.Get在线用户统计(c), "获取成功", c)
}

// ChartLinksUserIPCity 在线用户IP地图分布统计
func (p *Panel) ChartLinksUserIPCity(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get在线用户Ip地图分布统计(c), "获取成功", c)
}

// ChartLinksUserLoginTime 统计用户日活月活
func (p *Panel) ChartLinksUserLoginTime(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get统计用户日活月活(c), "获取成功", c)
}

// ChartEveryHourLinksCount 统计分时段在线总数
func (p *Panel) ChartEveryHourLinksCount(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get统计分时段在线总数(c), "获取成功", c)
}

// ChartAppUserClass 应用用户类型统计
func (p *Panel) ChartAppUserClass(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get应用用户类型统计(c), "获取成功", c)
}

// ChartUser 用户账号登录注册统计
func (p *Panel) ChartUser(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get用户账号登录注册统计(c), "获取成功", c)
}

// ChartRMBAddSub 余额充值消费统计
func (p *Panel) ChartRMBAddSub(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get余额充值消费统计(c), "获取成功", c)
}

// ChartVipNumberAddSub 积分点数消费统计
func (p *Panel) ChartVipNumberAddSub(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get积分点数消费统计(c), "获取成功", c)
}

// ChartAppUser 应用用户统计
func (p *Panel) ChartAppUser(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get应用用户统计(c), "获取成功", c)
}

// ChartAppKa 卡号列表统计应用卡可用已用
func (p *Panel) ChartAppKa(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get卡号列表统计应用卡可用已用(c), "获取成功", c)
}

// ChartAppKaClass 卡号列表统计应用卡类可用已用
func (p *Panel) ChartAppKaClass(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get卡号列表统计应用卡类可用已用(c), "获取成功", c)
}

// ChartKaRegister 卡号列表统计制卡
func (p *Panel) ChartKaRegister(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get卡号列表统计制卡(c), "获取成功", c)
}

// ChartKaMonthSummary 卡号月度汇总(本月/上月制卡数量和使用数量)
func (p *Panel) ChartKaMonthSummary(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get卡号月度汇总(c), "获取成功", c)
}

// ChartDashboardSummary 仪表台汇总(卡号总数/未使用 + 本月/上月充值成功总额)
func (p *Panel) ChartDashboardSummary(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get仪表台汇总(c), "获取成功", c)
}

// ChartAppUserRegister 应用用户账号注册统计
func (p *Panel) ChartAppUserRegister(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get应用用户账号注册统计(c), "获取成功", c)
}

// ChartAgentLevel 代理组织架构图
func (p *Panel) ChartAgentLevel(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get代理组织架构图(c, -1), "获取成功", c)
}

// ChartTidTaskData 任务池任务Id分析
func (p *Panel) ChartTidTaskData(c *gin.Context) {
	response.OkWithDetailed(L_chart.Get任务池任务Id分析(c), "获取成功", c)
}

type 请求_监控画像 struct {
	M名称   string `json:"name"`
	D调试等级 int    `json:"debug"`
	G执行GC bool   `json:"gc"`
}

type 请求_监控设置 struct {
	K开启Block画像 bool `json:"enableBlockProfile"`
	BBlock画像速率 int  `json:"blockProfileRate"`
	K开启Mutex画像 bool `json:"enableMutexProfile"`
	MMutex画像比例 int  `json:"mutexFraction"`
}

type 请求_CPU画像 struct {
	S秒数 int `json:"seconds"`
}

// GetMonitorOverview 获取监控总览
func (j *Panel) Q监控总览(c *gin.Context) {
	局_总览, err := monitoring.Q监控.Q监控总览()
	if err != nil {
		global.GVA_LOG.Println("获取监控总览失败!", err)
		response.FailWithMessage("获取监控总览失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(局_总览, "获取成功", c)
}

// GetMonitorProfileText 获取 pprof 文本
func (j *Panel) Q监控画像文本(c *gin.Context) {
	var req 请求_监控画像
	_ = c.ShouldBindJSON(&req)

	data, err := monitoring.Q监控.Q画像文本(req.M名称, req.D调试等级, req.G执行GC)
	if err != nil {
		response.FailWithMessage("读取pprof失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}

// UpdateMonitorSettings 更新 pprof 采样设置
func (j *Panel) G监控设置(c *gin.Context) {
	var req 请求_监控设置
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误:"+err.Error(), c)
		return
	}

	snapshot := monitoring.Q监控.G更新Pprof设置(
		req.K开启Block画像,
		req.BBlock画像速率,
		req.K开启Mutex画像,
		req.MMutex画像比例,
	)
	response.OkWithDetailed(snapshot, "设置成功", c)
}

// CaptureCPUProfile 手动抓取 CPU Profile
func (j *Panel) C抓取CPU画像(c *gin.Context) {
	var req 请求_CPU画像
	_ = c.ShouldBindJSON(&req)

	result, err := monitoring.Q监控.C抓取CPU画像(req.S秒数)
	if err != nil {
		response.FailWithMessage("抓取CPU Profile失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(result, "抓取成功", c)
}

// GetMonitorProcessTop 获取进程 CPU / 内存占用排行
func (j *Panel) Q监控进程排行(c *gin.Context) {
	局_结果, err := monitoring.Q监控.Q进程排行()
	if err != nil {
		response.FailWithMessage("获取进程排行失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(局_结果, "获取成功", c)
}

// DownloadMonitorProfile 下载非 CPU pprof 文件
func (j *Panel) Q下载监控画像(c *gin.Context) {
	var req 请求_监控画像
	_ = c.ShouldBindJSON(&req)

	data, filename, err := monitoring.Q监控.Q下载画像(req.M名称, req.G执行GC)
	if err != nil {
		response.FailWithMessage("下载pprof失败:"+err.Error(), c)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// DownloadLastCPUProfile 下载最近一次 CPU Profile
func (j *Panel) Q下载最近CPU画像(c *gin.Context) {
	data, filename, err := monitoring.Q监控.Q最近CPU画像()
	if err != nil {
		response.FailWithMessage("下载CPU Profile失败:"+err.Error(), c)
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// GetMonitorMySQLDiag 执行 MySQL 诊断（进程列表 / 慢日志开关 / SQL 摘要 Top10）
func (j *Panel) Q监控MySQL诊断(c *gin.Context) {
	局_结果, err := monitoring.Q监控.QMySQL诊断()
	if err != nil {
		response.FailWithMessage("MySQL 诊断失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(局_结果, "获取成功", c)
}
