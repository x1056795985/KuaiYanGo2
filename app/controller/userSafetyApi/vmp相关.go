package userSafetyApi

import (
	. "EFunc/utils"
	"crypto/x509"
	"encoding/pem"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/encoding/gjson"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/VMP"
	"server/app/logic/common/log"
	"server/app/models/common"
	"server/app/models/constant"
	"strconv"
	"time"
)

func UserApi_VMP计算授权码(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	//{"Api":"VmpComputeAuth","AppId":10001, Hwid:"dada4654","User":"adadasdasd"}
	var VmpRsa common.VmpRsa
	block, _ := pem.Decode([]byte(局_ctx.AppInfo.CryptoKeyPrivate))
	if block == nil {
		response.FailMsg(c, constant.Status_操作失败, "分解私钥失败")
		return
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "无法解析 PKCS1 私钥"+err.Error())
		return
	}

	VmpRsa.Rsa位数 = 1024
	VmpRsa.RsaBase64私钥 = B编码_BASE64编码(VMP.S十进制解码(privateKey.D))
	VmpRsa.RsaBase64模数 = B编码_BASE64编码(VMP.S十进制解码(privateKey.N))

	局_Base64产品代码字节 := Int32ToBytes(int32(局_ctx.Z在线信息.Uid))                                     //共计8个字节,前四个字节为在线用户用户uid 防山寨
	局_Base64产品代码字节 = append(局_Base64产品代码字节, Int32ToBytes(局_ctx.Q请求明文.Get("AppId").Int32())...) //补appid 4个字节 后四个字节为用户appid 防止用户串应用
	VmpRsa.Base64产品代码 = B编码_BASE64编码(局_Base64产品代码字节)

	var 局_授权参数 common.VmpParams
	局_授权参数.UserName = 局_ctx.Q请求明文.Get("User").String()
	//实测只需要授权一天即可,因为授权码使用后,所有功能不在受时间限制 实际还是需要靠心跳控制时分秒 精准度
	//激活码的到期时间只有激活的时候才检测,被保护的函数执行时不检测,所以登陆后立刻调用,当天有效即可
	//但是为了防止遇到极端11:59:59时间登陆的情况,所以有效时间设置为明天

	局_明天time := time.Now().AddDate(0, 0, 1)
	局_授权参数.ExpireDate.Year = 局_明天time.Year()
	局_授权参数.ExpireDate.Month = int(局_明天time.Month())
	局_授权参数.ExpireDate.Day = 局_明天time.Day()
	局_授权参数.MaxBuildDate = common.S时间{
		Year:  time.Now().Year(),
		Month: int(time.Now().Month()),
		Day:   time.Now().Day(),
	}
	局_授权参数.TimeLimit = 1
	局_授权参数.Hwid = 局_ctx.Q请求明文.Get("Hwid").String()

	var 授权码 string
	授权码, err = VMP.L_VMP.J计算授权码(nil, VmpRsa, 局_授权参数)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}
	response.OkData(c, gin.H{"VmpAuth": 授权码})
	return
}
func UserApi_VMP计算授权码防山寨(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	_, ok := global.H缓存.Get("VMP计算code_" + strconv.Itoa(局_ctx.Z在线信息.Id)) //获取
	if ok {                                                              //如果ok说明已经存在这个记录了
		go log.L_log.S写风控日志(c, 局_ctx.Z在线信息.Id, 1, 局_ctx.Z在线信息.User, c.ClientIP(), "用户一次登陆,多次重复计算VMP授权码,可能在尝试转发请求破解")
		response.FailMsg(c, constant.Status_操作失败, "禁止重复计算授权")
		//写风控日志
		return
	}

	//{"Api":"VmpComputeAuthRoot",Hwid:"dada4654" }

	var 局_响应信息 string
	var 局_错误代码 int
	if !global.Q快验.VMP计算授权码(&局_响应信息, 局_ctx.Z在线信息.LoginAppid, 局_ctx.Z在线信息.User, 局_ctx.Q请求明文.Get("Hwid").String()) {
		response.FailMsg(c, 局_错误代码, global.Q快验.Q取错误信息(&局_错误代码))
		return
	}
	局_响应json := gjson.New(局_响应信息) //必定是json 不然中间件就报错参数错误了
	//每个在线id 只允许获取一次
	global.H缓存.Set("VMP计算code_"+strconv.Itoa(局_ctx.Z在线信息.Id), 1, time.Minute*3600)
	response.OkData(c, gin.H{"VmpAuth": 局_响应json.Get("VmpAuth").String()})
	return
}
