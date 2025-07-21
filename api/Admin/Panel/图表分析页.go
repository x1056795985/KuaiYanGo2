package Panel

import (
	"github.com/gin-gonic/gin"
	"server/Service/Ser_Chare"
	"server/structs/Http/response"
)

func (s *Api) Get在线统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get在线用户统计(c), "获取成功", c)
}
func (s *Api) Get在线用户Ip地图分布统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get在线用户Ip地图分布统计(c), "获取成功", c)
}
func (s *Api) Get在线用户统计登录活动时间(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get在线用户统计登录活动时间(c), "获取成功", c)
}
func (s *Api) Get统计分时段在线总数(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get统计分时段在线总数(c), "获取成功", c)
}
func (s *Api) Get应用用户类型统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户类型统计(c), "获取成功", c)
}
func (s *Api) Get用户账号登录注册统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get用户账号登录注册统计(c), "获取成功", c)
}
func (s *Api) Get余额充值消费统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get余额充值消费统计(c), "获取成功", c)
}
func (s *Api) Get积分点数消费统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get积分点数消费统计(c), "获取成功", c)
}
func (s *Api) Get应用用户统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户统计(c), "获取成功", c)
}
func (s *Api) Get卡号列表统计应用卡可用已用(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计应用卡可用已用(c), "获取成功", c)
}
func (s *Api) Get卡号列表统计应用卡类可用已用(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计应用卡类可用已用(c), "获取成功", c)
}
func (s *Api) Get卡号列表统计制卡(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get卡号列表统计制卡(c), "获取成功", c)
}
func (s *Api) Get应用用户账号注册统计(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get应用用户账号注册统计(c), "获取成功", c)
}
func (s *Api) Get代理组织架构图(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get代理组织架构图(c, -1), "获取成功", c)
}
func (s *Api) Get任务池任务Id分析(c *gin.Context) {
	response.OkWithDetailed(Ser_Chare.Get任务池任务Id分析(c), "获取成功", c)
}
