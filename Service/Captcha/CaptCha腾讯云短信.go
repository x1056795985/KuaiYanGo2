package Captcha

import (
	"encoding/json"
	系统错误 "errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
	"net/url"
	"server/global"
	"strings"
)

func Sms_当前选择发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	switch global.GVA_CONFIG.D短信平台配置.D当前选择 {
	case 0, 1:
		return TX云_sms发送短信验证码(模板变量, 接收短信手机号)
	case 2:
		return D短信宝_sms发送短信验证码(模板变量, 接收短信手机号)
	default:
		return 系统错误.New("短信平台配置.当前选择配置无效")
	}
}

func TX云_sms发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	SecretId := global.GVA_CONFIG.D短信平台配置.TX云短信Sms.SECRET_ID
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效SECRET_ID")
	}

	SecretKey := global.GVA_CONFIG.D短信平台配置.TX云短信Sms.SECRET_KEY
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效SECRET_KEY")
	}
	短信应用ID := global.GVA_CONFIG.D短信平台配置.TX云短信Sms.D短信应用ID
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效短信应用ID")
	}
	短信签名 := global.GVA_CONFIG.D短信平台配置.TX云短信Sms.D短信签名
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效短信签名")
	}
	正文模板id := global.GVA_CONFIG.D短信平台配置.TX云短信Sms.Z正文模板ID
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效正文模板id")
	}

	/* 必要步骤：
	 * 实例化一个认证对象，入参需要传入腾讯云账户密钥对secretId，secretKey。
	 * 这里采用的是从环境变量读取的方式，需要在环境变量中先设置这两个值。
	 * 你也可以直接在代码中写死密钥对，但是小心不要将代码复制、上传或者分享给他人，
	 * 以免泄露密钥对危及你的财产安全。
	 * SecretId、SecretKey 查询: https://console.cloud.tencent.com/cam/capi */
	credential := common.NewCredential(
		SecretId,
		SecretKey,
	)
	/* 非必要步骤:
	 * 实例化一个客户端配置对象，可以指定超时时间等配置 */
	cpf := profile.NewClientProfile()

	/* SDK默认使用POST方法。
	 * 如果你一定要使用GET方法，可以在这里设置。GET方法无法处理一些较大的请求 */
	cpf.HttpProfile.ReqMethod = "POST"

	/* SDK有默认的超时时间，非必要请不要进行调整
	 * 如有需要请在代码中查阅以获取最新的默认值 */
	// cpf.HttpProfile.ReqTimeout = 5

	/* 指定接入地域域名，默认就近地域接入域名为 sms.tencentcloudapi.com ，也支持指定地域域名访问，例如广州地域的域名为 sms.ap-guangzhou.tencentcloudapi.com */
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"

	/* SDK默认用TC3-HMAC-SHA256进行签名，非必要请不要修改这个字段 */
	cpf.SignMethod = "HmacSHA1"

	/* 实例化要请求产品(以sms为例)的client对象
	 * 第二个参数是地域信息，可以直接填写字符串ap-guangzhou，支持的地域列表参考 https://cloud.tencent.com/document/api/382/52071#.E5.9C.B0.E5.9F.9F.E5.88.97.E8.A1.A8 */
	client, _ := sms.NewClient(credential, "ap-guangzhou", cpf)

	/* 实例化一个请求对象，根据调用的接口和实际情况，可以进一步设置请求参数
	 * 你可以直接查询SDK源码确定接口有哪些属性可以设置
	 * 属性可能是基本类型，也可能引用了另一个数据结构
	 * 推荐使用IDE进行开发，可以方便的 跳转查阅各个接口和数据结构的文档说明 */
	request := sms.NewSendSmsRequest()

	/* 基本类型的设置:
	 * SDK采用的是指针风格指定参数，即使对于基本类型你也需要用指针来对参数赋值。
	 * SDK提供对基本类型的指针引用封装函数
	 * 帮助链接：
	 * 短信控制台: https://console.cloud.tencent.com/smsv2
	 * 腾讯云短信小助手: https://cloud.tencent.com/document/product/382/3773#.E6.8A.80.E6.9C.AF.E4.BA.A4.E6.B5.81 */

	/* 短信应用ID: 短信SdkAppId在 [短信控制台] 添加应用后生成的实际SdkAppId，示例如1400006666 */
	// 应用 Id 可前往 [短信控制台](https://console.cloud.tencent.com/smsv2/app-manage) 查看
	request.SmsSdkAppId = common.StringPtr(短信应用ID)

	/* 短信签名内容: 使用 UTF-8 编码，必须填写已审核通过的签名 */
	// 签名信息可前往 [国内短信](https://console.cloud.tencent.com/smsv2/csms-sign) 或 [国际/港澳台短信](https://console.cloud.tencent.com/smsv2/isms-sign) 的签名管理查看
	request.SignName = common.StringPtr(短信签名)

	/* 模板 Id: 必须填写已审核通过的模板 Id */
	// 模板 Id 可前往 [国内短信](https://console.cloud.tencent.com/smsv2/csms-template) 或 [国际/港澳台短信](https://console.cloud.tencent.com/smsv2/isms-template) 的正文模板管理查看
	request.TemplateId = common.StringPtr(正文模板id)
	/* 模板参数: 模板参数的个数需要与 TemplateId 对应模板的变量个数保持一致，若无模板参数，则设置为空*/
	request.TemplateParamSet = common.StringPtrs(模板变量)

	/* 下发手机号码，采用 E.164 标准，+[国家或地区码][手机号]
	 * 示例如：+8613711112222， 其中前面有一个+号 ，86为国家码，13711112222为手机号，最多不要超过200个手机号*/
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + 接收短信手机号})

	/* 用户的 session 内容（无需要可忽略）: 可以携带用户侧 Id 等上下文信息，server 会原样返回 */
	request.SessionContext = common.StringPtr("")

	/* 短信码号扩展号（无需要可忽略）: 默认未开通，如需开通请联系 [腾讯云短信小助手] */
	request.ExtendCode = common.StringPtr("")

	/* 国内短信无需填写该项；国际/港澳台短信已申请独立 SenderId 需要填写该字段，默认使用公共 SenderId，无需填写该字段。注：月度使用量达到指定量级可申请独立 SenderId 使用，详情请联系 [腾讯云短信小助手](https://cloud.tencent.com/document/product/382/3773#.E6.8A.80.E6.9C.AF.E4.BA.A4.E6.B5.81)。 */
	request.SenderId = common.StringPtr("")

	// 通过client对象调用想要访问的接口，需要传入请求对象
	response, err := client.SendSms(request)
	// 处理异常
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("腾讯云SMS 返回API错误: %s", err)
		return 系统错误.New(err.Error())

	}
	// 非SDK异常，直接失败。实际代码中可以加入其他的处理。
	if err != nil {
		return 系统错误.New(err.Error())
	}

	if *response.Response.SendStatusSet[0].Code != "Ok" {
		b, _ := json.Marshal(response.Response)
		// 打印返回的json字符串
		fmt.Printf("腾讯云SMS 返回API错误: %s", b)
		return 系统错误.New(*response.Response.SendStatusSet[0].Code)
	}
	return nil
	/* 当出现以下错误码时，快速解决方案参考
	 * [FailedOperation.SignatureIncorrectOrUnapproved](https://cloud.tencent.com/document/product/382/9558#.E7.9F.AD.E4.BF.A1.E5.8F.91.E9.80.81.E6.8F.90.E7.A4.BA.EF.BC.9Afailedoperation.signatureincorrectorunapproved-.E5.A6.82.E4.BD.95.E5.A4.84.E7.90.86.EF.BC.9F)
	 * [FailedOperation.TemplateIncorrectOrUnapproved](https://cloud.tencent.com/document/product/382/9558#.E7.9F.AD.E4.BF.A1.E5.8F.91.E9.80.81.E6.8F.90.E7.A4.BA.EF.BC.9Afailedoperation.templateincorrectorunapproved-.E5.A6.82.E4.BD.95.E5.A4.84.E7.90.86.EF.BC.9F)
	 * [UnauthorizedOperation.SmsSdkAppIdVerifyFail](https://cloud.tencent.com/document/product/382/9558#.E7.9F.AD.E4.BF.A1.E5.8F.91.E9.80.81.E6.8F.90.E7.A4.BA.EF.BC.9Aunauthorizedoperation.smssdkappidverifyfail-.E5.A6.82.E4.BD.95.E5.A4.84.E7.90.86.EF.BC.9F)
	 * [UnsupportedOperation.ContainDomesticAndInternationalPhoneNumber](https://cloud.tencent.com/document/product/382/9558#.E7.9F.AD.E4.BF.A1.E5.8F.91.E9.80.81.E6.8F.90.E7.A4.BA.EF.BC.9Aunsupportedoperation.containdomesticandinternationalphonenumber-.E5.A6.82.E4.BD.95.E5.A4.84.E7.90.86.EF.BC.9F)
	 * 更多错误，可咨询[腾讯云助手](https://tccc.qcloud.com/web/im/index.html#/chat?webAppId=8fa15978f85cb41f7e2ea36920cb3ae1&title=Sms)
	 */
}

// 本命令由自动生成，请配合[ go get -u gitee.com/anyueyinluo/Efunc ]库使用。
func D短信宝_sms发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	局_短信宝 := global.GVA_CONFIG.D短信平台配置.Sms短信宝

	if 局_短信宝.User == "" {
		return 系统错误.New("Sms短信宝用户名配置无效")
	}
	if 局_短信宝.ApiKey == "" {
		return 系统错误.New("Sms短信宝ApiKey配置无效")
	}
	if 局_短信宝.F发送内容 == "" || strings.Index(局_短信宝.F发送内容, "{Code}") == -1 {
		return 系统错误.New("Sms短信宝F发送内容必须包含验证码占位符 {Code}")
	}
	for _, 值 := range 模板变量 {
		局_短信宝.F发送内容 = strings.Replace(局_短信宝.F发送内容, "{Code}", 值, 1)
	}

	局_网址 := `https://api.smsbao.com/sms`
	Http请求 := req.R()
	局_网址 += `?u={u}&p={p}&g={g}&m={m}&c={c}`
	Http请求.SetPathParam(`u`, 局_短信宝.User)
	Http请求.SetPathParam(`p`, 局_短信宝.ApiKey)
	Http请求.SetPathParam(`g`, 局_短信宝.C产品Id)
	Http请求.SetPathParam(`m`, 接收短信手机号)
	Http请求.SetPathParam(`c`, url.QueryEscape(局_短信宝.F发送内容))

	var 局_请求结果 *req.Response
	var err error
	for i := 0; i < 3; i++ { // 重试三次防止意外
		局_请求结果, err = Http请求.Get(局_网址)
		if len(局_请求结果.Bytes()) > 0 || err != nil {
			break
		}
	}
	var 局_返回 string
	局_返回 = 局_请求结果.String()

	switch 局_返回 {
	case "0":
		return nil
	case "30":
		return 系统错误.New("短信宝Api错误")
	case "40":
		return 系统错误.New("短信宝账号不存在")
	case "41":
		return 系统错误.New("短信宝余额不足")
	case "43":
		return 系统错误.New("短信宝IP地址限制")
	case "50":
		return 系统错误.New("短信宝内容含有敏感词")
	case "51":
		return 系统错误.New("短信宝手机号码不正确")
	}

	return 系统错误.New("未知错误:" + 局_返回)
}
