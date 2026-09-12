package controller

import (
	"encoding/base64"
	"html"
	"net/http"
	"server/app/global"
	rmbWithdrawLogic "server/app/logic/common/rmbWithdraw"
	"server/app/models/old/response"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"server/app/controller/Common"
	"server/app/models/request"
	. "server/app/models/response"
	"server/app/service"
)

type Withdraw struct {
	Common.Common
}

func NewWithdrawController() *Withdraw {
	return &Withdraw{}
}

func (j *Withdraw) GetConfig(c *gin.Context) {
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	response.OkWithDetailed(局_提现服务.Q取配置(局_数据库), "获取成功", c)
}

func (j *Withdraw) SaveConfig(c *gin.Context) {
	var 局_请求 service.T提现_配置
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	if err := 局_提现服务.B保存配置(局_数据库, 局_请求); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

func (j *Withdraw) List(c *gin.Context) {
	var 局_请求 service.T提现_列表请求
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数量, 局_列表, err := 局_提现服务.L列表(局_数据库, 局_请求, 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: 局_列表, Count: 局_数量}, "获取成功", c)
}

func (j *Withdraw) Detail(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数据, err := 局_提现服务.X详情(局_数据库, 局_请求.Id, 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(局_数据, "获取成功", c)
}

func (j *Withdraw) AuditPass(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if err := rmbWithdrawLogic.T提现_审核通过(global.Get局db(), 局_请求.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("审核通过", c)
}

func (j *Withdraw) Reject(c *gin.Context) {
	var 局_请求 struct {
		Id     int    `json:"id" binding:"required,min=1"`
		Reason string `json:"reason" binding:"required"`
	}
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if err := rmbWithdrawLogic.T提现_驳回(global.Get局db(), 局_请求.Id, 局_请求.Reason, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("驳回成功", c)
}

func (j *Withdraw) MarkPaid(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	if err := rmbWithdrawLogic.T提现_标记已付款(global.Get局db(), 局_请求.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已标记付款", c)
}

func (j *Withdraw) Delete(c *gin.Context) {
	var 局_请求 service.T提现_删除请求
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数量, err := 局_提现服务.S删除(局_数据库, 局_请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(局_数量, 10), c)
}

func (j *Withdraw) UploadVoucher(c *gin.Context) {
	局_提现Id, _ := strconv.Atoi(c.PostForm("id"))
	if 局_提现Id <= 0 {
		response.FailWithMessage("提现单id错误", c)
		return
	}
	局_文件, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择付款凭证", c)
		return
	}
	局_路径, err := rmbWithdrawLogic.T提现_上传凭证(global.Get局db(), 局_提现Id, 局_文件, c.GetInt("Uid"), c.GetString("User"), c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"path": 局_路径}, "上传成功", c)
}

func (j *Withdraw) Image(c *gin.Context) {
	var 局_请求 struct {
		Path string `json:"path" binding:"required"`
	}
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_图片信息, err := 局_提现服务.Q取管理图片(局_数据库, 局_请求.Path)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.File(局_图片信息.AbsPath)
}

func (j *Withdraw) CreateVoucherToken(c *gin.Context) {
	var 局_请求 request.Id2
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_令牌信息, err := 局_提现服务.C创建凭证令牌(局_请求.Id, c.GetInt("Uid"), c.GetString("User"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	局_上传地址 := "/" + strings.Trim(global.GVA_Viper.GetString("管理入口"), "/") + "/withdraw/uploadVoucherByToken?token=" + 局_令牌信息.Token
	局_完整上传地址 := 请求_来源(c) + 局_上传地址
	response.OkWithDetailed(gin.H{
		"token":         局_令牌信息.Token,
		"expireTime":    局_令牌信息.ExpireTime,
		"uploadUrl":     局_上传地址,
		"fullUploadUrl": 局_完整上传地址,
		"qrcodeBase64":  二维码_生成base64(局_完整上传地址),
	}, "创建成功", c)
}

func (j *Withdraw) UploadVoucherByTokenPage(c *gin.Context) {
	局_令牌 := c.Query("token")
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, `<!doctype html>
<html lang="zh-CN">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>上传付款凭证</title>
<style>
body{margin:0;background:#f4f6f8;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#202124}
.box{max-width:480px;margin:0 auto;padding:24px 18px}
.card{background:#fff;border:1px solid #e5e7eb;border-radius:10px;padding:18px;box-shadow:0 6px 20px rgba(0,0,0,.06)}
h1{font-size:20px;margin:0 0 14px}
input,button{width:100%;box-sizing:border-box}
input{padding:12px;border:1px solid #dcdfe6;border-radius:6px;background:#fff}
button{margin-top:14px;border:0;border-radius:6px;padding:12px;background:#1677ff;color:#fff;font-size:16px}
.msg{margin-top:12px;font-size:14px;color:#666;line-height:1.6}
</style>
</head>
<body><div class="box"><div class="card">
<h1>上传付款凭证</h1>
<form method="post" enctype="multipart/form-data">
<input type="hidden" name="token" value="`+html.EscapeString(局_令牌)+`">
<input type="file" name="file" accept="image/*" required>
<button type="submit">提交凭证</button>
</form>
<div class="msg">请选择手机相册中的付款截图，提交成功后电脑端会自动刷新凭证。</div>
</div></div></body></html>`)
}

func (j *Withdraw) UploadVoucherByToken(c *gin.Context) {
	局_令牌 := c.Query("token")
	if 局_令牌 == "" {
		局_令牌 = c.PostForm("token")
	}
	局_文件, err := c.FormFile("file")
	if 局_令牌 == "" || err != nil {
		response.FailWithMessage("token或文件不能为空", c)
		return
	}
	局_路径, err := rmbWithdrawLogic.T提现_使用令牌上传凭证(global.Get局db(), 局_令牌, 局_文件, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if strings.Contains(c.GetHeader("Accept"), "text/html") && c.PostForm("token") != "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<!doctype html><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><body style=\"font-family:sans-serif;padding:24px;text-align:center\"><h2>上传成功</h2><p>可以返回电脑端继续处理。</p></body>")
		return
	}
	response.OkWithDetailed(gin.H{"path": 局_路径}, "上传成功", c)
}

func (j *Withdraw) GetUploadVoucherByTokenStatus(c *gin.Context) {
	var 局_请求 struct {
		Token string `json:"token" binding:"required"`
	}
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_令牌信息, 局_有效 := 局_提现服务.Q取凭证令牌状态(局_请求.Token)
	response.OkWithDetailed(gin.H{"exists": 局_有效, "info": 局_令牌信息}, "获取成功", c)
}

func 请求_来源(c *gin.Context) string {
	局_协议 := c.GetHeader("X-Forwarded-Proto")
	if 局_协议 == "" {
		局_协议 = "http"
		if c.Request.TLS != nil {
			局_协议 = "https"
		}
	}
	局_主机 := c.GetHeader("X-Forwarded-Host")
	if 局_主机 == "" {
		局_主机 = c.Request.Host
	}
	return 局_协议 + "://" + 局_主机
}

func 二维码_生成base64(内容 string) string {
	局_图片, err := qrcode.Encode(内容, qrcode.Medium, 220)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(局_图片)
}

func (j *Withdraw) Logs(c *gin.Context) {
	var 局_请求 service.T提现_列表请求
	if !j.ToJSON(c, &局_请求) {
		return
	}
	局_提现服务 := service.S_RmbWithdraw{}
	局_数据库 := global.Get局db()
	局_数量, 局_列表, err := 局_提现服务.R日志列表(局_数据库, 局_请求)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: 局_列表, Count: 局_数量}, "获取成功", c)
}
