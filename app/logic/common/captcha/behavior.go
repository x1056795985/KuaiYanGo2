package captcha

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"server/app/global"
	"server/app/logic/common/setting"
)

const geetestValidateURL = "https://gcaptcha4.geetest.com/validate"

var captchaHTTPClient = &http.Client{Timeout: 8 * time.Second}

type geetestPayload struct {
	CaptchaID     string `json:"captcha_id"`
	LotNumber     string `json:"lot_number"`
	PassToken     string `json:"pass_token"`
	GenTime       string `json:"gen_time"`
	CaptchaOutput string `json:"captcha_output"`
}

type geetestResponse struct {
	Result string `json:"result"`
	Reason string `json:"reason"`
}

// VerifyBehavior dispatches to the behavior-captcha provider selected in settings.
func VerifyBehavior(id, payload string) error {
	switch setting.Q行为验证码平台配置().D当前选择 {
	case 0, 1:
		return VerifyGeetest(id, payload)
	case 2:
		return VerifyKuaiYanBehavior(id, payload)
	default:
		return errors.New("行为验证码平台配置.当前选择配置无效")
	}
}

// VerifyGeetest validates a GeeTest v4 payload. Provider transport failures
// remain fail-open to preserve the application's existing availability policy.
func VerifyGeetest(id, payload string) error {
	config := setting.Q行为验证码平台配置().J极验行为验证4
	if config.Y验证_ID == "" || config.Y验证_KEY == "" {
		return errors.New("极验行为验证码配置无效")
	}

	var requestData geetestPayload
	if err := json.Unmarshal([]byte(payload), &requestData); err != nil {
		return errors.New("验证码参数错误")
	}
	if requestData.CaptchaID == "" {
		requestData.CaptchaID = id
	}
	if requestData.CaptchaID == "" || requestData.LotNumber == "" || requestData.PassToken == "" || requestData.GenTime == "" || requestData.CaptchaOutput == "" {
		return errors.New("验证码内容错误")
	}

	form := url.Values{
		"lot_number":     {requestData.LotNumber},
		"captcha_output": {requestData.CaptchaOutput},
		"pass_token":     {requestData.PassToken},
		"gen_time":       {requestData.GenTime},
		"sign_token":     {hmacSHA256(config.Y验证_KEY, requestData.LotNumber)},
	}
	endpoint := geetestValidateURL + "?captcha_id=" + url.QueryEscape(config.Y验证_ID)
	httpRequest, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("创建极验请求失败: %w", err)
	}
	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := captchaHTTPClient.Do(httpRequest)
	if err != nil {
		return nil
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil
	}

	var result geetestResponse
	if err = json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return fmt.Errorf("解析极验响应失败: %w", err)
	}
	if result.Result != "success" {
		if result.Reason == "" {
			return errors.New("验证码错误")
		}
		return fmt.Errorf("验证码错误: %s", result.Reason)
	}
	return nil
}

func VerifyKuaiYanBehavior(id, payload string) error {
	if global.Q快验.K快验Api_极验验证码结果验证(id, payload) {
		return nil
	}
	return errors.New("验证码错误")
}

func hmacSHA256(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
