package controller

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"server/app/controller/Common"
	wangGuan "server/app/logic/wangGuan"
	"server/app/global"
	"server/app/models/dbm"
	"server/app/models/old/response"
	"server/app/models/request"
	"server/app/service"
	"strings"
)

// Gateway 网关转发管理(扩展->网关转发)
type Gateway struct {
	Common.Common
}

func NewGatewayController() *Gateway {
	return &Gateway{}
}

// 请求结构体
type 请求_GatewayGetList struct {
	request.List
}

type 请求_GatewayId struct {
	Id int `json:"Id" binding:"required,min=1"`
}

type 请求_GatewayIds struct {
	Ids []int `json:"Ids" binding:"required,min=1"`
}

type 请求_GatewayNew struct {
	Name       string `json:"Name" binding:"required,min=1,max=100"`
	Url        string `json:"Url" binding:"required,min=1,max=255"`
	Secret     string `json:"Secret"` //可空=自动生成
	TimeoutSec int    `json:"TimeoutSec"`
	Status     int    `json:"Status"`
	Remark     string `json:"Remark" binding:"max=255"`
}

type 请求_GatewaySave struct {
	Id         int    `json:"Id" binding:"required,min=1"`
	Name       string `json:"Name" binding:"required,min=1,max=100"`
	Url        string `json:"Url" binding:"required,min=1,max=255"`
	Secret     string `json:"Secret"` //与原值相同或为空=不修改;修改后旧密钥归档过渡
	TimeoutSec int    `json:"TimeoutSec"`
	Status     int    `json:"Status"`
	Remark     string `json:"Remark" binding:"max=255"`
}

// 响应_网关项 管理员拥有最大权限,Secret直接明文返回
type 响应_网关项 struct {
	Id         int    `json:"Id"`
	Name       string `json:"Name"`
	Url        string `json:"Url"`
	Secret     string `json:"Secret"`
	Status     int    `json:"Status"`
	TimeoutSec int    `json:"TimeoutSec"`
	Remark     string `json:"Remark"`
}

// GetList 网关分页列表
func (j *Gateway) GetList(c *gin.Context) {
	var 局_请求 请求_GatewayGetList
	if !j.ToJSON(c, &局_请求) {
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	总数, 局_列表, err := S.GetList(局_请求.List)
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	局_结果 := make([]响应_网关项, 0, len(局_列表))
	for _, v := range 局_列表 {
		局_结果 = append(局_结果, Q网关转响应(v))
	}
	response.OkWithDetailed(gin.H{"list": 局_结果, "count": 总数}, "获取成功", c)
}

// Info 单条详情
func (j *Gateway) Info(c *gin.Context) {
	var 局_请求 请求_GatewayId
	if !j.ToJSON(c, &局_请求) {
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	局_info, err := S.Info(局_请求.Id)
	if err != nil {
		response.FailWithMessage("获取失败:"+err.Error(), c)
		return
	}
	response.OkWithDetailed(Q网关转响应(局_info), "获取成功", c)
}

// New 新增网关,Secret自动生成且仅本次返回明文
func (j *Gateway) New(c *gin.Context) {
	var 局_请求 请求_GatewayNew
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if 局_错误 := Q校验网关Url(局_请求.Url); 局_错误 != nil {
		response.FailWithMessage(局_错误.Error(), c)
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	if 局_info, _ := S.InfoName(局_请求.Name); 局_info.Id > 0 {
		response.FailWithMessage("网关名称已存在", c)
		return
	}
	var 局_生成错误 error
	局_secret := strings.TrimSpace(局_请求.Secret)
	if 局_secret == "" {
		局_secret, 局_生成错误 = wangGuan.L_网关.Q生成密钥()
		if 局_生成错误 != nil {
			response.FailWithMessage("生成密钥失败:"+局_生成错误.Error(), c)
			return
		}
	} else if 局_错误 := Q校验Secret(局_secret); 局_错误 != nil {
		response.FailWithMessage(局_错误.Error(), c)
		return
	}
	局_状态 := 局_请求.Status
	if 局_状态 != 1 && 局_状态 != 2 {
		局_状态 = 1
	}
	局_info := dbm.DB_Gateway{
		Name:       局_请求.Name,
		Url:        strings.TrimRight(局_请求.Url, "/"),
		Secret:     局_secret,
		Status:     局_状态,
		TimeoutSec: 局_请求.TimeoutSec,
		Remark:     局_请求.Remark,
	}
	if _, 局_创建错误 := S.Create(&局_info); 局_创建错误 != nil {
		response.FailWithMessage("新增失败:"+局_创建错误.Error(), c)
		return
	}
	_ = wangGuan.G刷新网关缓存()
	response.OkWithMessage("新增成功", c)
}

// SaveInfo 修改网关(密钥字段不接收,只能通过轮换接口变更)
func (j *Gateway) SaveInfo(c *gin.Context) {
	var 局_请求 请求_GatewaySave
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if 局_错误 := Q校验网关Url(局_请求.Url); 局_错误 != nil {
		response.FailWithMessage(局_错误.Error(), c)
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	if 局_info, _ := S.InfoName(局_请求.Name); 局_info.Id > 0 && 局_info.Id != 局_请求.Id {
		response.FailWithMessage("网关名称已存在", c)
		return
	}
	局_状态 := 局_请求.Status
	if 局_状态 != 1 && 局_状态 != 2 {
		局_状态 = 1
	}
	局_info, err := S.Info(局_请求.Id)
	if err != nil {
		response.FailWithMessage("网关不存在", c)
		return
	}
	局_更新 := map[string]interface{}{
		"Name":       局_请求.Name,
		"Url":        strings.TrimRight(局_请求.Url, "/"),
		"Status":     局_状态,
		"TimeoutSec": 局_请求.TimeoutSec,
		"Remark":     局_请求.Remark,
	}
	//Secret 手动修改:直接替换,业务侧自行同步更新验签密钥
	局_新密钥 := strings.TrimSpace(局_请求.Secret)
	if 局_新密钥 != "" && 局_新密钥 != 局_info.Secret {
		if 局_错误 := Q校验Secret(局_新密钥); 局_错误 != nil {
			response.FailWithMessage(局_错误.Error(), c)
			return
		}
		局_更新["Secret"] = 局_新密钥
	}
	_, err = S.Update(局_请求.Id, 局_更新)
	if err != nil {
		response.FailWithMessage("保存失败:"+err.Error(), c)
		return
	}
	_ = wangGuan.G刷新网关缓存()
	response.OkWithMessage("保存成功", c)
}

// Delete 批量删除
func (j *Gateway) Delete(c *gin.Context) {
	var 局_请求 请求_GatewayIds
	if !j.ToJSON(c, &局_请求) {
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	影响行数, err := S.Delete(局_请求.Ids)
	if err != nil {
		response.FailWithMessage("删除失败:"+err.Error(), c)
		return
	}
	_ = wangGuan.G刷新网关缓存()
	response.OkWithMessage(fmt.Sprintf("删除成功,影响%d行", 影响行数), c)
}

// Test 连通性测试:网关侧探测业务地址
func (j *Gateway) Test(c *gin.Context) {
	var 局_请求 请求_GatewayId
	if !j.ToJSON(c, &局_请求) {
		return
	}
	db := *global.GVA_DB
	S := service.NewGateway(c, &db)
	局_info, err := S.Info(局_请求.Id)
	if err != nil {
		response.FailWithMessage("网关不存在", c)
		return
	}
	局_开始 := time.Now()
	局_客户端 := &http.Client{Timeout: 3 * time.Second}
	局_响应, 局_错误 := 局_客户端.Get(strings.TrimRight(局_info.Url, "/") + "/")
	局_耗时 := time.Since(局_开始).Milliseconds()
	if 局_错误 != nil {
		response.OkWithDetailed(gin.H{
			"IsOk":       false,
			"耗时":         局_耗时,
			"Err":        局_错误.Error(),
		}, "业务服务不可达", c)
		return
	}
	_ = 局_响应.Body.Close()
	response.OkWithDetailed(gin.H{
		"IsOk":       true,
		"耗时":         局_耗时,
		"StatusCode": 局_响应.StatusCode,
	}, "业务服务可达", c)
}

// Q网关转响应 表转响应结构
func Q网关转响应(info dbm.DB_Gateway) 响应_网关项 {
	return 响应_网关项{
		Id:         info.Id,
		Name:       info.Name,
		Url:        info.Url,
		Secret:     info.Secret,
		Status:     info.Status,
		TimeoutSec: info.TimeoutSec,
		Remark:     info.Remark,
	}
}

// Q校验Secret 密钥格式校验:8~128位且不含空白
func Q校验Secret(密钥 string) error {
	if len(密钥) < 8 || len(密钥) > 128 {
		return fmt.Errorf("Secret长度必须在8到128位之间")
	}
	if strings.ContainsFunc(密钥, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\r' || r == '\n'
	}) {
		return fmt.Errorf("Secret不能包含空白字符")
	}
	return nil
}

// Q校验网关Url 必须是合法的http/https地址
func Q校验网关Url(地址 string) error {
	地址 = strings.TrimSpace(地址)
	if !strings.HasPrefix(地址, "http://") && !strings.HasPrefix(地址, "https://") {
		return fmt.Errorf("Url必须以http://或https://开头")
	}
	if strings.Contains(地址, " ") {
		return fmt.Errorf("Url不能包含空格")
	}
	return nil
}
