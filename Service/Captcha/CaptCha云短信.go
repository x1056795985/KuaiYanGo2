package Captcha

import (
	"EFunc/utils"
	系统错误 "errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/qiniu/go-sdk/v7/auth"
	sms_七牛云 "github.com/qiniu/go-sdk/v7/sms"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	sms_tx "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111" // 引入sms
	"server/global"
	"server/new/app/logic/common/setting"
	"strings"
)

func Sms_当前选择发送短信验证码(模板变量 []string, 接收短信手机号 string) error {

	局_临时 := setting.Q短信平台配置().D当前选择
	switch 局_临时 {
	case 0, 1:
		return TX云_sms发送短信验证码(模板变量, 接收短信手机号)
	case 2:
		return D短信宝_sms发送短信验证码(模板变量, 接收短信手机号)
	case 3:
		return Q七牛云_sms发送短信验证码(模板变量, 接收短信手机号)
	case 4:
		return K快验_sms发送短信验证码(模板变量, 接收短信手机号)
	default:
		return 系统错误.New("短信平台配置.当前选择配置无效")
	}
}

func TX云_sms发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	局_配置 := setting.Q短信平台配置()
	SecretId := 局_配置.TX云短信Sms.SECRET_ID
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效SECRET_ID")
	}

	SecretKey := 局_配置.TX云短信Sms.SECRET_KEY
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效SECRET_KEY")
	}
	短信应用ID := 局_配置.TX云短信Sms.D短信应用ID
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效短信应用ID")
	}
	短信签名 := 局_配置.TX云短信Sms.D短信签名
	if SecretId == "" {
		return 系统错误.New("TX短信配置无效短信签名")
	}
	正文模板id := 局_配置.TX云短信Sms.Z正文模板ID
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
	client, _ := sms_tx.NewClient(credential, "ap-guangzhou", cpf)

	/* 实例化一个请求对象，根据调用的接口和实际情况，可以进一步设置请求参数
	 * 你可以直接查询SDK源码确定接口有哪些属性可以设置
	 * 属性可能是基本类型，也可能引用了另一个数据结构
	 * 推荐使用IDE进行开发，可以方便的 跳转查阅各个接口和数据结构的文档说明 */
	request := sms_tx.NewSendSmsRequest()

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
	if *response.Response.SendStatusSet[0].Code == "Ok" {
		return nil
	}

	var 全部错误码中文 = `FailedOperation.ContainSensitiveWord	短信内容中含有敏感词，请联系 腾讯云短信小助手。
FailedOperation.FailResolvePacket	请求包解析失败，通常情况下是由于没有遵守 API 接口说明规范导致的，请参考 请求包体解析1004错误详解。
FailedOperation.ForbidAddMarketingTemplates	个人用户不能申请营销短信。
FailedOperation.InsufficientBalanceInSmsPackage	套餐包余量不足，请 购买套餐包。
FailedOperation.JsonParseFail	解析请求包体时候失败。
FailedOperation.MarketingSendTimeConstraint	营销短信发送时间限制，为避免骚扰用户，营销短信只允许在8点到22点发送。
FailedOperation.MissingSignature	没有申请签名之前，无法申请模板，请根据 创建签名 申请完成之后再次申请。
FailedOperation.MissingSignatureList	无法识别签名，请确认是否已有签名通过申请，一般是签名未通过申请，可以查看 签名审核。
FailedOperation.MissingSignatureToModify	此签名 ID 未提交申请或不存在，不能进行修改操作，请检查您的 SignId 是否填写正确。
FailedOperation.MissingTemplateList	无法识别模板，请确认是否已有模板通过申请，一般是模板未通过申请，可以查看 模板审核。
FailedOperation.MissingTemplateToModify	此模板 ID 未提交申请或不存在，不能进行修改操作，请检查您的 TemplateId是否填写正确。
FailedOperation.NotEnterpriseCertification	非企业认证无法使用签名及模板相关接口，您可以 变更实名认证模式，变更为企业认证用户后，约1小时左右生效。
FailedOperation.OtherError	其他错误，一般是由于参数携带不符合要求导致，请参考API接口说明，如有需要请联系 腾讯云短信小助手。
FailedOperation.ParametersOtherError	未知错误，如有需要请联系 腾讯云短信小助手。
FailedOperation.PhoneNumberInBlacklist	手机号在免打扰名单库中，通常是用户退订或者命中运营商免打扰名单导致的，可联系 腾讯云短信小助手 解决。
FailedOperation.PhoneNumberParseFail	号码解析失败，请检查号码是否符合 E.164 标准。
FailedOperation.ProhibitSubAccountUse	非主账号无法使用拉取模板列表功能。您可以使用主账号下云 API 密钥来调用接口。
FailedOperation.SignIdNotExist	签名 ID 不存在。
FailedOperation.SignNumberLimit	签名个数达到最大值。
FailedOperation.SignatureIncorrectOrUnapproved	签名未审批或格式错误。（1）可登录 短信控制台，核查签名是否已审批并且审批通过；（2）核查是否符合格式规范，签名只能由中英文、数字组成，要求2 - 12个字，若存在疑问可联系 腾讯云短信小助手。
FailedOperation.TemplateAlreadyPassedCheck	此模板已经通过审核，无法再次进行修改。
FailedOperation.TemplateIdNotExist	模板 ID 不存在。
FailedOperation.TemplateIncorrectOrUnapproved	模板未审批或内容不匹配。（1）可登录 短信控制台，核查模板是否已审批并审批通过；（2）核查是否符合 格式规范，若存在疑问可联系 腾讯云短信小助手。
FailedOperation.TemplateNumberLimit	模板个数达到最大值。
FailedOperation.TemplateParamSetNotMatchApprovedTemplate	请求内容与审核通过的模板内容不匹配。请检查请求中模板参数的个数是否与申请的模板一致。若存在疑问可联系 腾讯云短信小助手。
FailedOperation.TemplateUnapprovedOrNotExist	模板未审批或不存在。可登录 短信控制台，核查模板是否已审批并审批通过。若存在疑问可联系 腾讯云短信小助手。
InternalError.JsonParseFail	解析用户参数失败，可联系 腾讯云短信小助手。
InternalError.OtherError	其他错误，请联系 腾讯云短信小助手 并提供失败手机号。
InternalError.ParseBackendResponseFail	解析运营商包体失败，可联系 sms helper 。
InternalError.RequestTimeException	请求发起时间不正常，通常是由于您的服务器时间与腾讯云服务器时间差异超过10分钟导致的，请核对服务器时间及 API 接口中的时间字段是否正常。
InternalError.RestApiInterfaceNotExist	不存在该 RESTAPI 接口，请核查 REST API 接口说明。
InternalError.SendAndRecvFail	接口超时或短信收发包超时，请检查您的网络是否有波动，或联系 腾讯云短信小助手 解决。
InternalError.SigFieldMissing	后端包体中请求包体没有 Sig 字段或 Sig 为空。
InternalError.SigVerificationFail	后端校验 Sig 失败。
InternalError.Timeout	请求下发短信超时，请参考 60008错误详解。
InternalError.UnknownError	未知错误类型。
InvalidParameter.AppidAndBizId	账号与应用id不匹配。
InvalidParameter.DirtyWordFound	存在敏感词。
InvalidParameter.InvalidParameters	参数有误，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.BeginTimeVerifyFail	参数 BeginTime 校验失败。
InvalidParameterValue.ContentLengthLimit	请求的短信内容太长，短信长度规则请参考 国内短信内容长度计算规则。
InvalidParameterValue.EndTimeVerifyFail	参数 EndTime 校验失败。
InvalidParameterValue.ImageInvalid	上传的转码图片格式错误，请参照 API 接口说明中对该字段的说明，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.IncorrectPhoneNumber	手机号格式错误。
InvalidParameterValue.InvalidDocumentType	DocumentType 字段校验错误，请参照 API 接口说明中对该字段的说明，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.InvalidInternational	International 字段校验错误，请参照 API 接口说明中对该字段的说明，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.InvalidSignPurpose	SignPurpose 字段校验错误，请参照 API 接口说明中对该字段的说明，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.InvalidStartTime	无效的拉取起始/截止时间，具体原因可能是请求的 SendDateTime 大于 EndDateTime。
InvalidParameterValue.InvalidTemplateFormat	模板格式错误，请参考正文模板审核标准。
InvalidParameterValue.InvalidUsedMethod	UsedMethod 字段校验错误，请参照 API 接口说明中对该字段的说明，如有需要请联系 腾讯云短信小助手。
InvalidParameterValue.LimitVerifyFail	参数 Limit 校验失败。
InvalidParameterValue.MarketingTemplateWithoutUnsubscribe	营销短信必须包含退订方式，请在短信模板尾部添加“拒收请回复R”后提交。可参考 关于营销短信退订标识修改的公告。
InvalidParameterValue.OffsetVerifyFail	参数 Offset 校验失败。
InvalidParameterValue.ProhibitedUseUrlInTemplateParameter	禁止在模板变量中使用 URL。
InvalidParameterValue.SdkAppIdNotExist	SdkAppId 不存在。
InvalidParameterValue.SignAlreadyPassedCheck	此签名已经通过审核，无法再次进行修改。
InvalidParameterValue.SignExistAndUnapproved	已存在相同的待审核签名。
InvalidParameterValue.SignNameLengthTooLong	签名内容长度过长。
InvalidParameterValue.TemplateParameterFormatError	验证码模板参数格式错误，验证码类模板，模板变量只能传入0 - 6位（包括6位）纯数字。
InvalidParameterValue.TemplateParameterLengthLimit	单个模板变量字符数超过12个，企业认证用户不限制单个变量值字数，您可以 变更实名认证模式，变更为企业认证用户后，该限制变更约1小时左右生效。
InvalidParameterValue.TemplateWithDirtyWords	模板内容存在敏感词，请参考正文模板审核标准。
LimitExceeded.AppCountryOrRegionDailyLimit	业务短信国家/地区日下发条数超过设定的上限，可自行到控制台应用管理>基础配置下调整国际港澳台短信发送限制。
LimitExceeded.AppCountryOrRegionInBlacklist	业务短信国家/地区不在国际港澳台短信发送限制设置的列表中而禁发，可自行到控制台应用管理>基础配置下调整国际港澳台短信发送限制。
LimitExceeded.AppDailyLimit	业务短信日下发条数超过设定的上限 ，可自行到控制台调整短信频率限制策略。
LimitExceeded.AppGlobalDailyLimit	业务短信国际/港澳台日下发条数超过设定的上限，可自行到控制台应用管理>基础配置下调整发送总量阈值。
LimitExceeded.AppMainlandChinaDailyLimit	业务短信中国大陆日下发条数超过设定的上限，可自行到控制台应用管理>基础配置下调整发送总量阈值。
LimitExceeded.DailyLimit	短信日下发条数超过设定的上限 (国际/港澳台)，如需调整限制，可联系 腾讯云短信小助手。
LimitExceeded.DeliveryFrequencyLimit	下发短信命中了频率限制策略，可自行到控制台调整短信频率限制策略，如有其他需求请联系 腾讯云短信小助手。
LimitExceeded.PhoneNumberCountLimit	调用接口单次提交的手机号个数超过200个，请遵守 API 接口输入参数 PhoneNumberSet 描述。
LimitExceeded.PhoneNumberDailyLimit	单个手机号日下发短信条数超过设定的上限，可自行到控制台调整短信频率限制策略。
LimitExceeded.PhoneNumberOneHourLimit	单个手机号1小时内下发短信条数超过设定的上限，可自行到控制台调整短信频率限制策略。
LimitExceeded.PhoneNumberSameContentDailyLimit	单个手机号下发相同内容超过设定的上限，可自行到控制台调整短信频率限制策略。
LimitExceeded.PhoneNumberThirtySecondLimit	单个手机号30秒内下发短信条数超过设定的上限，可自行到控制台调整短信频率限制策略。
MissingParameter.EmptyPhoneNumberSet	传入的号码列表为空，请确认您的参数中是否传入号码。
UnauthorizedOperation.IndividualUserMarketingSmsPermissionDeny	个人用户没有发营销短信的权限，请参考 权益区别。
UnauthorizedOperation.RequestIpNotInWhitelist	请求 IP 不在白名单中，您配置了校验请求来源 IP，但是检测到当前请求 IP 不在配置列表中，如有需要请联系 腾讯云短信小助手。
UnauthorizedOperation.RequestPermissionDeny	请求没有权限，请联系 腾讯云短信小助手。
UnauthorizedOperation.SdkAppIdIsDisabled	此 SdkAppId 禁止提供服务，如有需要请联系 腾讯云短信小助手。
UnauthorizedOperation.ServiceSuspendDueToArrears	欠费被停止服务，可自行登录腾讯云充值来缴清欠款。
UnauthorizedOperation.SmsSdkAppIdVerifyFail	SmsSdkAppId 校验失败，请检查 SmsSdkAppId 是否属于 云API密钥 的关联账户。
UnsupportedOperation.	不支持该请求。
UnsupportedOperation.ChineseMainlandTemplateToGlobalPhone	国内短信模板不支持发送国际/港澳台手机号。发送国际/港澳台手机号请使用国际/港澳台短信正文模板。
UnsupportedOperation.ContainDomesticAndInternationalPhoneNumber	群发请求里既有国内手机号也有国际手机号。请排查是否存在（1）使用国内签名或模板却发送短信到国际手机号；（2）使用国际签名或模板却发送短信到国内手机号。
UnsupportedOperation.GlobalTemplateToChineseMainlandPhone	国际/港澳台短信模板不支持发送国内手机号。发送国内手机号请使用国内短信正文模板。
UnsupportedOperation.UnsupportedRegion	不支持该地区短信下发。
`

	局_中文翻译 := utils.W文本_取出中间文本(全部错误码中文, *response.Response.SendStatusSet[0].Code, "\n")
	if 局_中文翻译 != "" {
		return 系统错误.New(局_中文翻译)
	}
	return 系统错误.New(*response.Response.SendStatusSet[0].Code)

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
	局_短信宝 := setting.Q短信平台配置().Sms短信宝

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
	Http请求.SetPathParam(`c`, 局_短信宝.F发送内容)

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
func Q七牛云_sms发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	局_Sms七牛云 := setting.Q短信平台配置().Sms七牛云

	if 局_Sms七牛云.AccessKey == "" {
		return 系统错误.New("Sms七牛云AccessKey配置无效")
	}
	if 局_Sms七牛云.SecretKey == "" {
		return 系统错误.New("Sms七牛云SecretKey配置无效")
	}
	if 局_Sms七牛云.SignatureID == "" {
		return 系统错误.New("Sms七牛云SignatureID配置无效")
	}
	if 局_Sms七牛云.TemplateID == "" {
		return 系统错误.New("Sms七牛云TemplateID配置无效")
	}
	if len(模板变量) == 0 {
		return 系统错误.New("Sms七牛云模板变量无效")
	}
	var manager *sms_七牛云.Manager
	mac := auth.New(局_Sms七牛云.AccessKey, 局_Sms七牛云.SecretKey)
	manager = sms_七牛云.NewManager(mac)
	message, err := manager.SendMessage(sms_七牛云.MessagesRequest{
		局_Sms七牛云.SignatureID,
		局_Sms七牛云.TemplateID,
		[]string{接收短信手机号},
		map[string]interface{}{
			"code": 模板变量[0],
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("七牛云发送短信回调" + message.JobID)
	return err
}
func K快验_sms发送短信验证码(模板变量 []string, 接收短信手机号 string) error {
	if global.Q快验.K快验Api_发送验证码短信(模板变量, 接收短信手机号) {
		return nil
	}
	return 系统错误.New(global.Q快验.Q取错误信息(nil))
}
