package userSafetyApi

import (
	. "EFunc/utils"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"github.com/dgrijalva/jwt-go"
	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"
	"server/app/controller/userSafetyApi/response"
	"server/app/global"
	"server/app/logic/common/cloudStorage"
	"server/app/logic/common/jsEngine"
	"server/app/logic/webUser/appInfoWebUser"
	"server/app/models/constant"
	dbm "server/app/models/db"
	"server/app/service"
	"strings"
	"time"
)

// 1.0.325+版本添加可用
func UserApi_云存储_取文件上传授权(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}

	// {"Api":"GetUploadToken","Path":"8987657"}
	path := strings.TrimSpace(局_ctx.Q请求明文.Get("Path").String())

	if path == "" || strings.Index(path, ".") == -1 || W文本_取右边(path, 1) == "/" {
		response.FailMsg(c, constant.Status_操作失败, "暂不支持该文件类型")
		return
	}
	取文件上传授权, err := cloudStorage.L_云存储.Q取文件上传授权(c, path)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, err.Error())
		return
	}

	response.OkData(c, 取文件上传授权)
	return
}
func UserApi_云函数执行(c *gin.Context) {
	defer func() {
		if err2 := recover(); err2 != nil {
			局_GoJa错误, ok := err2.(*goja.Exception)
			if ok {
				response.FailMsg(c, constant.Status_操作失败, "异常:可能JS函数传参或返回值类型错误,具体:"+局_GoJa错误.String())
			} else {
				response.FailMsg(c, constant.Status_操作失败, "异常:可能JS函数传参或返回值类型错误,具体:js引擎未返回报错信息")
			}
			return
		}
	}()

	局_ctx := 取上下文(c)

	// {"Api":"RunJS","Parameter":"{'a':1}","JsName":"获取用户相关信息","IsGlobal":false,"Time":1684497856,"Status":30873}
	var 局_JSid = 0
	if 局_ctx.Q请求明文.Get("IsGlobal").Bool() {
		局_JSid = service.NewPublicJs(c, global.GVA_DB).Name取Id([]int{service.Js类型_公共函数}, 局_ctx.Q请求明文.Get("JsName").String())
	} else {
		局_JSid = service.NewPublicJs(c, global.GVA_DB).Name取Id([]int{局_ctx.AppInfo.AppId}, 局_ctx.Q请求明文.Get("JsName").String())
	}
	if 局_JSid == 0 {
		response.FailMsg(c, constant.Status_操作失败, "JS公共函数不存在")
		return
	}
	局_耗时 := time.Now().UnixMilli()

	var 局_PublicJs dbm.DB_PublicJs
	var err error
	局_PublicJs, err = service.NewPublicJs(c, global.GVA_DB).Q取值2(局_JSid)

	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "JS公共函数不存在")
		return
	}
	if 局_PublicJs.IsVip > 0 && !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	if W文件_是否存在(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value) {
		局_PublicJs.Value = string(W文件_读入文件(global.GVA_CONFIG.Q取运行目录 + 局_PublicJs.Value))
	} else {
		response.FailMsg(c, constant.Status_操作失败, "js文件读取失败可能被删除")
		return
	}

	局_云函数型参数 := ""
	if 局_ctx.Q请求明文.Get("Parameter").IsMap() {
		局_云函数型参数 = 局_ctx.Q请求明文.Get("Parameter").String()
	} else {
		局_云函数型参数 = 局_ctx.Q请求明文.Get("Parameter").String()
	}
	vm := jsEngine.J脚本引擎_初始化用户(c, &局_ctx.AppInfo, &局_ctx.Z在线信息, &局_PublicJs)
	_, err = vm.RunString(局_PublicJs.Value)
	if 局_详细错误, ok := err.(*goja.Exception); ok {
		response.FailMsg(c, constant.Status_操作失败, "JS代码运行失败:"+局_详细错误.String())
		return
	}
	var 局_待执行js函数名 func(string) interface{}
	ret := vm.Get(局_PublicJs.Name)
	if ret == nil {
		response.FailMsg(c, constant.Status_操作失败, "Js中没有["+局_PublicJs.Name+"()]函数")
		return
	}
	err = vm.ExportTo(ret, &局_待执行js函数名)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "Js绑定函数到变量失败")
		return
	}
	局_return := 局_待执行js函数名(局_云函数型参数)
	response.OkData(c, gin.H{"Return": 局_return, "Time": time.Now().UnixMilli() - 局_耗时})
	return
}

// 1.0.310+版本添加可用
func UserApi_取jwtToken(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	db := *global.GVA_DB
	局_AppUser, err := service.NewAppUser(c, &db, 局_ctx.AppInfo.AppId).InfoUid(局_ctx.Z在线信息.Uid)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "读取用户应用信息失败.")
		return
	}
	局_UserClass, _ := service.NewUserClass(c, &db).Info(局_AppUser.UserClassId)
	jwtMap := jwt.MapClaims{}
	_ = json.Unmarshal([]byte(局_ctx.Q请求明文.String()), &jwtMap) //必定是json 不然中间件就报错参数错误了
	//提交的数据都加入到内容里,方便hookAPi

	鉴权密钥 := []byte(局_ctx.AppInfo.CryptoKeyPrivate)
	delete(jwtMap, "Api")
	delete(jwtMap, "Key")
	delete(jwtMap, "Time")
	delete(jwtMap, "Status")

	//这个数据放后面,需要覆盖本地端的数据,防止伪造
	jwtMap["iat"] = time.Now().Unix() // 发布时间
	jwtMap["Uid"] = 局_AppUser.Uid
	jwtMap["User"] = 局_ctx.Z在线信息.User
	jwtMap["Key"] = 局_AppUser.Key
	jwtMap["VipTime"] = 局_AppUser.VipTime
	jwtMap["VipNumber"] = 局_AppUser.VipNumber
	jwtMap["MaxOnline"] = 局_AppUser.MaxOnline
	jwtMap["AgentUid"] = 局_AppUser.AgentUid
	jwtMap["UserClassId"] = 局_AppUser.UserClassId
	jwtMap["UserClassName"] = 局_UserClass.Name
	jwtMap["UserClassMark"] = 局_UserClass.Mark
	jwtMap["UserClassWeight"] = 局_UserClass.Weight
	// 创建一个JWT的Token对象
	block, _ := pem.Decode(鉴权密钥)
	if block == nil {
		response.FailMsg(c, constant.Status_操作失败, "PEM 解析失败")
		return
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "私钥解析失败: "+err.Error())
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwtMap)
	signedToken, err := token.SignedString(privateKey)
	if err != nil {
		response.FailMsg(c, constant.Status_操作失败, "生成JWT失败.")
		return
	}
	response.OkData(c, gin.H{"Jwt": signedToken})
	return
}

/* 已经登陆的状态下,获取登陆web用户中心的登陆短链,实现直接登陆无需密码的登陆
 */
func UserApi_取登陆短链(c *gin.Context) {
	局_ctx := 取上下文(c)
	if !检测用户登录在线正常(&局_ctx.Z在线信息) {
		response.Fail(c, constant.Status_未登录)
		return
	}
	//{"Api":"LoginShortUrl","JumpUrl":"pages/user/home"}
	局_key := W文本_取随机字符串(8)
	//不能使用软件用户信息,必须使用 在线id  因为可能存在, 获取短链后,用户修改密码的情况,这时必须保证短链失效
	global.H缓存.Set(constant.H缓存前缀_LoginURLPrefix+局_key, 局_ctx.Z在线信息.Id, time.Minute*5) //短链有效期5分钟

	局_jump := 局_ctx.Q请求明文.Get("JumpUrl").String() //登陆成功后302跳转路由
	if 局_jump != "" && strings.Index(局_jump, "#/") != -1 {
		局_jump = W文本_取文本右边(局_jump, "#/")
	}
	if 局_jump == "" {
		局_jump = "pages/user/home"
	}

	局_临时地址 := appInfoWebUser.L_appInfoWebUser.Q用户中心域名(c, 局_ctx.AppInfo.AppId) + "/userApi/base/loginKey?k=" + 局_key + "&j=" + 局_jump
	局_ret := make(map[string]string, 2)
	局_ret["webUser"] = 局_临时地址
	局_ret["webUserPng"] = T图片_生成二维码base64(局_临时地址)

	response.OkData(c, 局_ret)
}
