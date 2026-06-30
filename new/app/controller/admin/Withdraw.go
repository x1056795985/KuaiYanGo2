package controller

import (
	"encoding/base64"
	"html"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/skip2/go-qrcode"
	"server/global"
	"server/new/app/controller/Common"
	"server/new/app/models/request"
	. "server/new/app/models/response"
	"server/new/app/service"
	"server/structs/Http/response"
)

type Withdraw struct {
	Common.Common
}

func NewWithdrawController() *Withdraw {
	return &Withdraw{}
}

func (C *Withdraw) GetConfig(c *gin.Context) {
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	response.OkWithDetailed(S.GetConfig(&tx), "获取成功", c)
}

func (C *Withdraw) SaveConfig(c *gin.Context) {
	var req service.WithdrawConfig
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	if err := S.SaveConfig(&tx, req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("保存成功", c)
}

func (C *Withdraw) List(c *gin.Context) {
	var req service.WithdrawListRequest
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	count, list, err := S.List(&tx, req, 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: list, Count: count}, "获取成功", c)
}

func (C *Withdraw) Detail(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	data, err := S.Detail(&tx, req.Id, 0)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(data, "获取成功", c)
}

func (C *Withdraw) AuditPass(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	if err := S.AuditPass(req.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("审核通过", c)
}

func (C *Withdraw) Reject(c *gin.Context) {
	var req struct {
		Id     int    `json:"id" binding:"required,min=1"`
		Reason string `json:"reason" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	if err := S.Reject(req.Id, req.Reason, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("驳回成功", c)
}

func (C *Withdraw) MarkPaid(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	if err := S.MarkPaid(req.Id, c.GetInt("Uid"), c.GetString("User"), c.ClientIP()); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("已标记付款", c)
}

func (C *Withdraw) Delete(c *gin.Context) {
	var req service.WithdrawDeleteRequest
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	count, err := S.Delete(&tx, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithMessage("删除成功,数量"+strconv.FormatInt(count, 10), c)
}

func (C *Withdraw) UploadVoucher(c *gin.Context) {
	id, _ := strconv.Atoi(c.PostForm("id"))
	if id <= 0 {
		response.FailWithMessage("提现单id错误", c)
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("请选择付款凭证", c)
		return
	}
	var S = service.S_RmbWithdraw{}
	path, err := S.UploadVoucher(id, file, c.GetInt("Uid"), c.GetString("User"), c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(gin.H{"path": path}, "上传成功", c)
}

func (C *Withdraw) Image(c *gin.Context) {
	var req struct {
		Path string `json:"path" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	info, err := S.GetAdminImage(&tx, req.Path)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	c.File(info.AbsPath)
}

func (C *Withdraw) CreateVoucherToken(c *gin.Context) {
	var req request.Id2
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	info, err := S.CreateVoucherToken(req.Id, c.GetInt("Uid"), c.GetString("User"))
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	uploadUrl := "/" + strings.Trim(global.GVA_Viper.GetString("管理入口"), "/") + "/withdraw/uploadVoucherByToken?token=" + info.Token
	fullUploadUrl := requestOrigin(c) + uploadUrl
	response.OkWithDetailed(gin.H{
		"token":         info.Token,
		"expireTime":    info.ExpireTime,
		"uploadUrl":     uploadUrl,
		"fullUploadUrl": fullUploadUrl,
		"qrcodeBase64":  makeQrBase64(fullUploadUrl),
	}, "创建成功", c)
}

func (C *Withdraw) UploadVoucherByTokenPage(c *gin.Context) {
	token := c.Query("token")
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
<input type="hidden" name="token" value="`+html.EscapeString(token)+`">
<input type="file" name="file" accept="image/*" required>
<button type="submit">提交凭证</button>
</form>
<div class="msg">请选择手机相册中的付款截图，提交成功后电脑端会自动刷新凭证。</div>
</div></div></body></html>`)
}

func (C *Withdraw) UploadVoucherByToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		token = c.PostForm("token")
	}
	file, err := c.FormFile("file")
	if token == "" || err != nil {
		response.FailWithMessage("token或文件不能为空", c)
		return
	}
	var S = service.S_RmbWithdraw{}
	path, err := S.UploadVoucherByToken(token, file, c.ClientIP())
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	if strings.Contains(c.GetHeader("Accept"), "text/html") && c.PostForm("token") != "" {
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, "<!doctype html><meta charset=\"utf-8\"><meta name=\"viewport\" content=\"width=device-width,initial-scale=1\"><body style=\"font-family:sans-serif;padding:24px;text-align:center\"><h2>上传成功</h2><p>可以返回电脑端继续处理。</p></body>")
		return
	}
	response.OkWithDetailed(gin.H{"path": path}, "上传成功", c)
}

func (C *Withdraw) GetUploadVoucherByTokenStatus(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	info, ok := S.GetVoucherTokenStatus(req.Token)
	response.OkWithDetailed(gin.H{"exists": ok, "info": info}, "获取成功", c)
}

func requestOrigin(c *gin.Context) string {
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
		if c.Request.TLS != nil {
			proto = "https"
		}
	}
	host := c.GetHeader("X-Forwarded-Host")
	if host == "" {
		host = c.Request.Host
	}
	return proto + "://" + host
}

func makeQrBase64(content string) string {
	png, err := qrcode.Encode(content, qrcode.Medium, 220)
	if err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(png)
}

func (C *Withdraw) Logs(c *gin.Context) {
	var req service.WithdrawListRequest
	if !C.ToJSON(c, &req) {
		return
	}
	var S = service.S_RmbWithdraw{}
	tx := *global.GVA_DB
	count, list, err := S.Logs(&tx, req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}
	response.OkWithDetailed(GetList2{List: list, Count: count}, "获取成功", c)
}
