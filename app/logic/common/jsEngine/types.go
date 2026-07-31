package jsEngine

import "github.com/gin-gonic/gin"

const 脚本引擎_缓存键前缀 = "gghsjs_"

type 脚本引擎_Api结果 struct {
	IsOk bool   `json:"IsOk"`
	Err  string `json:"Err"`
	Data any    `json:"Data"`
}

type 脚本引擎_Http响应 struct {
	StatusCode int    `json:"StatusCode"`
	Headers    string `json:"Headers"`
	Cookies    string `json:"Cookies"`
	Body       string `json:"Body"`
	Base64Body string `json:"base64Body"`
}

func 脚本引擎_后台上下文() *gin.Context {
	return &gin.Context{}
}

func 脚本引擎_成功(message string, data any) 脚本引擎_Api结果 {
	return 脚本引擎_Api结果{IsOk: true, Err: message, Data: data}
}

func 脚本引擎_失败(err error) 脚本引擎_Api结果 {
	if err == nil {
		return 脚本引擎_Api结果{IsOk: false}
	}
	return 脚本引擎_Api结果{IsOk: false, Err: err.Error()}
}

func 脚本引擎_失败消息(message string) 脚本引擎_Api结果 {
	return 脚本引擎_Api结果{IsOk: false, Err: message}
}
