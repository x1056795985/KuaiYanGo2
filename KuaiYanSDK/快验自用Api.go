package KuaiYanSDK

func (k *Api快验_类) K快验Api_发送验证码短信(模板变量 []string, 接收短信手机号 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "KyApiSendSms"
	请求json["Code"] = 模板变量
	请求json["Phone"] = 接收短信手机号

	_, ok := k.通讯(请求json)

	return ok
}

func (k *Api快验_类) K快验Api_极验验证码结果验证(验证码id, 验证码内容 string) bool {
	请求json := make(map[string]interface{}, 10)
	请求json["Api"] = "KyApiJiYanVerifyTicket"
	请求json["CaptchaId"] = 验证码id
	请求json["CaptchaValue"] = 验证码内容

	响应json, ok := k.通讯(请求json)
	if ok {
		return 响应json.GetBool("Data", "Code")
	}
	return true //如果通讯失败,也放行,防止影响客户

}
