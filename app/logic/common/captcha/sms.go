package captcha

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/qiniu/go-sdk/v7/auth"
	qiniusms "github.com/qiniu/go-sdk/v7/sms"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"

	"server/app/global"
	"server/app/logic/common/setting"
)

var smsHTTPClient = &http.Client{Timeout: 8 * time.Second}

// SendSMS dispatches a verification message to the configured provider.
func SendSMS(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	switch setting.Q短信平台配置().D当前选择 {
	case 0, 1:
		return SendTencentSMS(templateVariables, phone)
	case 2:
		return SendSMSBao(templateVariables, phone)
	case 3:
		return SendQiniuSMS(templateVariables, phone)
	case 4:
		return SendKuaiYanSMS(templateVariables, phone)
	case 5:
		return SendAliyunSMS(templateVariables, phone)
	default:
		return errors.New("短信平台配置.当前选择配置无效")
	}
}

func SendTencentSMS(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	config := setting.Q短信平台配置().TX云短信Sms
	switch {
	case config.SECRET_ID == "":
		return errors.New("TX短信配置无效SECRET_ID")
	case config.SECRET_KEY == "":
		return errors.New("TX短信配置无效SECRET_KEY")
	case config.D短信应用ID == "":
		return errors.New("TX短信配置无效短信应用ID")
	case config.D短信签名 == "":
		return errors.New("TX短信配置无效短信签名")
	case config.Z正文模板ID == "":
		return errors.New("TX短信配置无效正文模板ID")
	}

	clientProfile := profile.NewClientProfile()
	clientProfile.HttpProfile.ReqMethod = http.MethodPost
	clientProfile.HttpProfile.ReqTimeout = 8
	clientProfile.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, err := tencentsms.NewClient(common.NewCredential(config.SECRET_ID, config.SECRET_KEY), "ap-guangzhou", clientProfile)
	if err != nil {
		return fmt.Errorf("创建腾讯云短信客户端失败: %w", err)
	}

	request := tencentsms.NewSendSmsRequest()
	request.SmsSdkAppId = common.StringPtr(config.D短信应用ID)
	request.SignName = common.StringPtr(config.D短信签名)
	request.TemplateId = common.StringPtr(config.Z正文模板ID)
	request.TemplateParamSet = common.StringPtrs(templateVariables)
	request.PhoneNumberSet = common.StringPtrs([]string{"+86" + phone})

	response, err := client.SendSms(request)
	if err != nil {
		return fmt.Errorf("腾讯云短信请求失败: %w", err)
	}
	if response == nil || response.Response == nil || len(response.Response.SendStatusSet) == 0 || response.Response.SendStatusSet[0] == nil {
		return errors.New("腾讯云短信返回为空")
	}
	status := response.Response.SendStatusSet[0]
	if stringValue(status.Code) == "Ok" {
		return nil
	}
	return fmt.Errorf("腾讯云短信发送失败: %s - %s", stringValue(status.Code), stringValue(status.Message))
}

func SendSMSBao(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	config := setting.Q短信平台配置().Sms短信宝
	switch {
	case config.User == "":
		return errors.New("Sms短信宝用户名配置无效")
	case config.ApiKey == "":
		return errors.New("Sms短信宝ApiKey配置无效")
	case !strings.Contains(config.F发送内容, "{Code}"):
		return errors.New("Sms短信宝发送内容必须包含验证码占位符 {Code}")
	}

	content := config.F发送内容
	for _, value := range templateVariables {
		content = strings.Replace(content, "{Code}", value, 1)
	}
	query := url.Values{
		"u": {config.User},
		"p": {config.ApiKey},
		"g": {config.C产品Id},
		"m": {phone},
		"c": {content},
	}
	response, err := smsHTTPClient.Get("https://api.smsbao.com/sms?" + query.Encode())
	if err != nil {
		return fmt.Errorf("短信宝请求失败: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, 1024))
	if err != nil {
		return fmt.Errorf("读取短信宝响应失败: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("短信宝HTTP状态异常: %s", response.Status)
	}
	code := strings.TrimSpace(string(body))
	if code == "0" {
		return nil
	}
	messages := map[string]string{
		"30": "短信宝Api错误",
		"40": "短信宝账号不存在",
		"41": "短信宝余额不足",
		"43": "短信宝IP地址限制",
		"50": "短信宝内容含有敏感词",
		"51": "短信宝手机号码不正确",
	}
	if message, ok := messages[code]; ok {
		return errors.New(message)
	}
	return fmt.Errorf("短信宝未知错误: %s", code)
}

func SendQiniuSMS(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	config := setting.Q短信平台配置().Sms七牛云
	switch {
	case config.AccessKey == "":
		return errors.New("Sms七牛云AccessKey配置无效")
	case config.SecretKey == "":
		return errors.New("Sms七牛云SecretKey配置无效")
	case config.SignatureID == "":
		return errors.New("Sms七牛云SignatureID配置无效")
	case config.TemplateID == "":
		return errors.New("Sms七牛云TemplateID配置无效")
	}

	manager := qiniusms.NewManager(auth.New(config.AccessKey, config.SecretKey))
	_, err := manager.SendMessage(qiniusms.MessagesRequest{
		SignatureID: config.SignatureID,
		TemplateID:  config.TemplateID,
		Mobiles:     []string{phone},
		Parameters:  map[string]any{"code": templateVariables[0]},
	})
	if err != nil {
		return fmt.Errorf("七牛云短信发送失败: %w", err)
	}
	return nil
}

func SendKuaiYanSMS(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	if global.Q快验.K快验Api_发送验证码短信(templateVariables, phone) {
		return nil
	}
	return errors.New(global.Q快验.Q取错误信息(nil))
}

func SendAliyunSMS(templateVariables []string, phone string) error {
	if err := validateSMSInput(templateVariables, phone); err != nil {
		return err
	}
	config := setting.Q短信平台配置().Sms阿里云
	switch {
	case config.AccessKeyId == "":
		return errors.New("Sms阿里云AccessKeyId配置无效")
	case config.AccessKeySecret == "":
		return errors.New("Sms阿里云AccessKeySecret配置无效")
	case config.D短信签名 == "":
		return errors.New("Sms阿里云短信签名配置无效")
	case config.Z正文模板Code == "":
		return errors.New("Sms阿里云正文模板Code配置无效")
	}

	nonce, err := randomToken(16, "0123456789")
	if err != nil {
		return fmt.Errorf("生成阿里云短信请求标识失败: %w", err)
	}
	templateParameter, err := json.Marshal(map[string]string{"code": templateVariables[0]})
	if err != nil {
		return fmt.Errorf("编码阿里云短信模板参数失败: %w", err)
	}
	parameters := map[string]string{
		"AccessKeyId":      config.AccessKeyId,
		"Action":           "SendSms",
		"Format":           "JSON",
		"PhoneNumbers":     phone,
		"RegionId":         "cn-hangzhou",
		"SignName":         config.D短信签名,
		"SignatureMethod":  "HMAC-SHA1",
		"SignatureNonce":   nonce,
		"SignatureVersion": "1.0",
		"TemplateCode":     config.Z正文模板Code,
		"TemplateParam":    string(templateParameter),
		"Timestamp":        time.Now().UTC().Format("2006-01-02T15:04:05Z"),
		"Version":          "2017-05-25",
	}

	keys := make([]string, 0, len(parameters))
	for key := range parameters {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, aliyunURLEncode(key)+"="+aliyunURLEncode(parameters[key]))
	}
	normalizedQuery := strings.Join(parts, "&")
	stringToSign := "GET&" + aliyunURLEncode("/") + "&" + aliyunURLEncode(normalizedQuery)
	mac := hmac.New(sha1.New, []byte(config.AccessKeySecret+"&"))
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	endpoint := "https://dysmsapi.aliyuncs.com/?" + normalizedQuery + "&Signature=" + aliyunURLEncode(signature)

	response, err := smsHTTPClient.Get(endpoint)
	if err != nil {
		return fmt.Errorf("阿里云短信请求失败: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("阿里云短信HTTP状态异常: %s", response.Status)
	}
	var result struct {
		Code    string `json:"Code"`
		Message string `json:"Message"`
	}
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return fmt.Errorf("解析阿里云短信响应失败: %w", err)
	}
	if result.Code == "OK" {
		return nil
	}
	return fmt.Errorf("阿里云短信发送失败: %s - %s", result.Code, result.Message)
}

func validateSMSInput(templateVariables []string, phone string) error {
	if strings.TrimSpace(phone) == "" {
		return errors.New("短信手机号不能为空")
	}
	if len(templateVariables) == 0 {
		return errors.New("短信模板变量不能为空")
	}
	return nil
}

func aliyunURLEncode(value string) string {
	encoded := url.QueryEscape(value)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	return strings.ReplaceAll(encoded, "%7E", "~")
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
