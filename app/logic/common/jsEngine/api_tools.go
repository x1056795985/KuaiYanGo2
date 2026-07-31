package jsEngine

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/imroc/req/v3"

	"server/app/global"
	"server/app/logic/common/VMP"
	"server/app/logic/common/captcha"
	"server/app/logic/common/cloudStorage"
	"server/app/models/common"
)

const 脚本引擎_默认HTTP超时秒数 = 15

func 脚本引擎_延时(milliseconds int64) bool {
	time.Sleep(time.Duration(milliseconds) * time.Millisecond)
	return true
}

func 脚本引擎_网页访问Get(rawURL, headers, cookies string, timeoutSeconds int, proxyURL string) 脚本引擎_Http响应 {
	return 脚本引擎_执行HTTP请求(http.MethodGet, rawURL, "", headers, cookies, timeoutSeconds, proxyURL)
}

func 脚本引擎_网页访问Post(rawURL, body, headers, cookies string, timeoutSeconds int, proxyURL string) 脚本引擎_Http响应 {
	return 脚本引擎_执行HTTP请求(http.MethodPost, rawURL, body, headers, cookies, timeoutSeconds, proxyURL)
}

func 脚本引擎_执行HTTP请求(method, rawURL, body, headers, cookies string, timeoutSeconds int, proxyURL string) 脚本引擎_Http响应 {
	if timeoutSeconds <= 0 {
		timeoutSeconds = 脚本引擎_默认HTTP超时秒数
	}
	局_客户端 := req.C().
		EnableInsecureSkipVerify().
		SetTimeout(time.Duration(timeoutSeconds) * time.Second).
		EnableForceHTTP1().
		SetRedirectPolicy(req.NoRedirectPolicy())
	if proxyURL != "" {
		if _, 局_错误 := url.ParseRequestURI(proxyURL); 局_错误 != nil {
			return 脚本引擎_Http响应{Body: "代理地址错误: " + 局_错误.Error()}
		}
		局_客户端.SetProxyURL(proxyURL)
	}
	局_请求 := 局_客户端.R().SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 Chrome/114 Safari/537.36")
	脚本引擎_应用请求头(局_请求, headers)
	if cookies != "" {
		局_请求.SetHeader("Cookie", cookies)
	}
	if method == http.MethodPost {
		if json.Valid([]byte(body)) {
			局_请求.SetHeader("Content-Type", "application/json")
			局_请求.SetHeader("Accept", "application/json, text/plain, */*")
		}
		局_请求.SetBody(body)
	}

	var 局_响应 *req.Response
	var 局_错误 error
	if method == http.MethodPost {
		局_响应, 局_错误 = 局_请求.Post(rawURL)
	} else {
		局_响应, 局_错误 = 局_请求.Get(rawURL)
	}
	if 局_错误 != nil {
		return 脚本引擎_Http响应{Body: 局_错误.Error()}
	}
	局_响应内容 := 局_响应.Bytes()
	return 脚本引擎_Http响应{
		StatusCode: 局_响应.StatusCode,
		Headers:    局_响应.HeaderToString(),
		Cookies:    脚本引擎_合并Cookie(cookies, 局_响应.Cookies()),
		Body:       string(局_响应内容),
		Base64Body: base64.StdEncoding.EncodeToString(局_响应内容),
	}
}

func 脚本引擎_应用请求头(request *req.Request, encoded string) {
	for _, 局_请求头行 := range strings.FieldsFunc(encoded, func(字符 rune) bool { return 字符 == '\r' || 字符 == '\n' }) {
		局_名称, 局_值, 局_存在 := strings.Cut(局_请求头行, ":")
		if 局_存在 && strings.TrimSpace(局_名称) != "" {
			request.SetHeader(strings.TrimSpace(局_名称), strings.TrimSpace(局_值))
		}
	}
}

func 脚本引擎_合并Cookie(existing string, received []*http.Cookie) string {
	局_Cookie表 := make(map[string]string, len(received)+4)
	for _, 局_项目 := range strings.Split(existing, ";") {
		局_名称, 局_值, 局_存在 := strings.Cut(strings.TrimSpace(局_项目), "=")
		if 局_存在 && 局_名称 != "" {
			局_Cookie表[局_名称] = 局_值
		}
	}
	for _, 局_Cookie := range received {
		局_Cookie表[局_Cookie.Name] = 局_Cookie.Value
	}
	局_键数组 := make([]string, 0, len(局_Cookie表))
	for 局_键 := range 局_Cookie表 {
		局_键数组 = append(局_键数组, 局_键)
	}
	sort.Strings(局_键数组)
	var 局_结果 strings.Builder
	for _, 局_键 := range 局_键数组 {
		局_结果.WriteString(局_键)
		局_结果.WriteByte('=')
		局_结果.WriteString(局_Cookie表[局_键])
		局_结果.WriteByte(';')
	}
	return 局_结果.String()
}

func 脚本引擎_执行SQL查询(query string, parameters []any) 脚本引擎_Api结果 {
	局_结果数组 := make([]map[string]any, 0)
	if 局_错误 := global.GVA_DB.Raw(query, parameters...).Scan(&局_结果数组).Error; 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	局_编码结果, 局_错误 := json.Marshal(局_结果数组)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_Api结果{IsOk: true, Err: string(局_编码结果), Data: 局_结果数组}
}

func 脚本引擎_执行SQL功能(query string, parameters []any) 脚本引擎_Api结果 {
	局_结果 := global.GVA_DB.Exec(query, parameters...)
	if 局_结果.Error != nil {
		return 脚本引擎_失败(局_结果.Error)
	}
	return 脚本引擎_成功(strconv.FormatInt(局_结果.RowsAffected, 10), nil)
}

func 脚本引擎_取缓存(name string) string {
	if name == "" || global.H缓存 == nil {
		return ""
	}
	局_值, 局_存在 := global.H缓存.Get(脚本引擎_缓存键前缀 + name)
	if !局_存在 {
		return ""
	}
	局_文本, _ := 局_值.(string)
	return 局_文本
}

func 脚本引擎_置缓存(name, value string, lifetimeSeconds int) bool {
	if name == "" || global.H缓存 == nil {
		return false
	}
	局_键 := 脚本引擎_缓存键前缀 + name
	if value == "" {
		global.H缓存.Delete(局_键)
	} else {
		global.H缓存.Set(局_键, value, time.Duration(lifetimeSeconds)*time.Second)
	}
	return true
}

func 脚本引擎_Jwt生成(jsonData, secret string) 脚本引擎_Api结果 {
	if jsonData == "" {
		return 脚本引擎_失败消息("签名数据不正确")
	}
	if secret == "" {
		return 脚本引擎_失败消息("签名密钥不正确")
	}
	局_声明 := jwt.MapClaims{}
	if 局_错误 := json.Unmarshal([]byte(jsonData), &局_声明); 局_错误 != nil {
		return 脚本引擎_失败消息("JSON数据异常: " + 局_错误.Error())
	}
	局_令牌, 局_错误 := jwt.NewWithClaims(jwt.SigningMethodHS256, 局_声明).SignedString([]byte(secret))
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", 局_令牌)
}

func 脚本引擎_云存储取外链(path string, validSeconds int64) 脚本引擎_Api结果 {
	局_值, 局_错误 := cloudStorage.L_云存储.Q取外链地址(脚本引擎_后台上下文(), path, validSeconds)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", 局_值)
}

func 脚本引擎_云存储取文件上传授权(path string) 脚本引擎_Api结果 {
	局_值, 局_错误 := cloudStorage.L_云存储.Q取文件上传授权(脚本引擎_后台上下文(), path)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", 局_值)
}

func 脚本引擎_云存储_文件信息(path string) 脚本引擎_Api结果 {
	局_值, 局_错误 := cloudStorage.L_云存储.Q取文件信息(脚本引擎_后台上下文(), path)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", 局_值)
}

func 脚本引擎_VMP计算授权码(bits int, privateKeyBase64, modulusBase64, productCodeBase64, paramsJSON string) 脚本引擎_Api结果 {
	var 局_参数 struct {
		common.VmpParams
		UserDataBase64 string `json:"UserDataBas64"`
	}
	if 局_错误 := json.Unmarshal([]byte(paramsJSON), &局_参数); 局_错误 != nil {
		return 脚本引擎_失败消息("授权信息json格式错误")
	}
	if 局_参数.UserDataBase64 != "" {
		局_解码数据, 局_错误 := base64.StdEncoding.DecodeString(局_参数.UserDataBase64)
		if 局_错误 != nil {
			return 脚本引擎_失败消息("UserDataBas64格式错误")
		}
		局_参数.UserData = 局_解码数据
	}
	局_值, 局_错误 := VMP.L_VMP.J计算授权码(nil, common.VmpRsa{
		Rsa位数: bits, RsaBase64私钥: privateKeyBase64,
		RsaBase64模数: modulusBase64, Base64产品代码: productCodeBase64,
	}, 局_参数.VmpParams)
	if 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("", 局_值)
}

func 脚本引擎_短信发送(templateVariables []string, phone string) 脚本引擎_Api结果 {
	if 局_错误 := captcha.SendSMS(templateVariables, phone); 局_错误 != nil {
		return 脚本引擎_失败(局_错误)
	}
	return 脚本引擎_成功("成功", nil)
}
