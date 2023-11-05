package Captcha

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	系统错误 "errors"
	"fmt"
	"github.com/pkg/errors"
	"github.com/valyala/fastjson"
	"io/ioutil"
	"net/http"
	"net/url"
	"server/global"
	"time"
)

func CaptCha_行为验证码验证(验证码id, 验证码内容 string) error {
	局_临时 := global.GVA_CONFIG.X行为验证码平台配置.D当前选择
	局_临时 = 1
	switch 局_临时 {
	case 0, 1:
		return J极验_滑动验证码参数验证(验证码id, 验证码内容)
	case 2:
		return K快验_极验验证码结果验证(验证码id, 验证码内容)

	default:
		return 系统错误.New("行为验证码平台配置.当前选择配置无效")
	}
}

// 1,验证通过,2 极验服务器异常,验证通过,3验证失败
func J极验_滑动验证码参数验证(验证码id, 验证码内容 string) error {
	局_极验id := global.GVA_CONFIG.X行为验证码平台配置.J极验行为验证4.Y验证_ID
	局_极验Key := global.GVA_CONFIG.X行为验证码平台配置.J极验行为验证4.Y验证_KEY

	//{"captcha_id":"ea872eea9e20dce9de4e5da4297ee704","lot_number":"cb287e80eb44498bb58e53825e003291","pass_token":"bb5de929069b7d4b271a81e41d33906a5e3aa0cd5e869bdf3fbc8ee8be787c26","gen_time":"1685944268","captcha_output":"BFZFX-sB_WLMfXEmGExfGuhvp6VqzHxjzMIoh3mpHd5VPSZVHisuUaD6HVMfjFmOCnNSYUfZGIEIexJ4CLXbcX_X4a2sJd1ooQ1V5JhElfUUFtUhAS1OdUQJZ1j_XhIg8KjXQV9BD9v3S0c0awHd334-GR86EqVXkYD9O0iDtnpfMZyzEFt7IevV-buSwexVQAR5erRhQz0imMaWic3l8yTPOLYDxm-_UDvwikMk5krDTRFqBPmWyeBAFcJ8-skL"}
	局_json, err2 := fastjson.Parse(验证码内容)
	if err2 != nil {
		return errors.New("验证码参数错误")
	}

	lot_number := string(局_json.GetStringBytes("lot_number"))
	gen_time := string(局_json.GetStringBytes("gen_time"))
	pass_token := string(局_json.GetStringBytes("pass_token"))
	captcha_output := string(局_json.GetStringBytes("captcha_output"))
	captcha_id := string(局_json.GetStringBytes("captcha_id"))
	if captcha_id == "" {
		captcha_id = 验证码id
	}

	if captcha_id == "" || captcha_output == "" || pass_token == "" || gen_time == "" {
		return errors.New("验证码内容错误")
	}

	//{"captcha_id":"ea872eea9e20dce9de4e5da4297ee704","lot_number":"cb287e80eb44498bb58e53825e003291","pass_token":"bb5de929069b7d4b271a81e41d33906a5e3aa0cd5e869bdf3fbc8ee8be787c26","gen_time":"1685944268","captcha_output":"BFZFX-sB_WLMfXEmGExfGuhvp6VqzHxjzMIoh3mpHd5VPSZVHisuUaD6HVMfjFmOCnNSYUfZGIEIexJ4CLXbcX_X4a2sJd1ooQ1V5JhElfUUFtUhAS1OdUQJZ1j_XhIg8KjXQV9BD9v3S0c0awHd334-GR86EqVXkYD9O0iDtnpfMZyzEFt7IevV-buSwexVQAR5erRhQz0imMaWic3l8yTPOLYDxm-_UDvwikMk5krDTRFqBPmWyeBAFcJ8-skL"}
	// 前端传回的数据

	// 生成签名
	// Generate signature
	// 生成签名使用标准的hmac算法，使用用户当前完成验证的流水号lot_number作为原始消息message，使用客户验证私钥作为key
	// use standard hmac algorithms to generate signatures, and take the user's current verification serial number lot_number as the original message, and the client's verification private key as the key
	// 采用sha256散列算法将message和key进行单向散列生成最终的 “sign_token” 签名
	// use sha256 hash algorithm to hash message and key in one direction to generate the final signature
	sign_token := hmac_encode(局_极验Key, lot_number)

	// 向极验转发前端数据 + “sign_token” 签名
	// send front end parameter + "sign_token" signature to geetest
	form_data := make(url.Values)
	form_data["lot_number"] = []string{lot_number}
	form_data["captcha_output"] = []string{captcha_output}
	form_data["pass_token"] = []string{pass_token}
	form_data["gen_time"] = []string{gen_time}
	form_data["sign_token"] = []string{sign_token}

	// 发起post请求
	// initialize a post request
	// 设置5s超时
	// set a 5 seconds timeout
	cli := http.Client{Timeout: time.Second * 5}
	resp, err := cli.PostForm("http://gcaptcha4.geetest.com/validate?captcha_id="+局_极验id, form_data)
	if err != nil || resp.StatusCode != 200 {
		// 当请求发生异常时，应放行通过，以免阻塞业务。
		// when geetest server interface exceptions occur, the request should pass in order not to interrupt the website's business
		return nil
	}

	res_json, _ := ioutil.ReadAll(resp.Body)
	var res_map map[string]interface{}
	// 根据极验返回的用户验证状态, 网站主进行自己的业务逻辑
	// taking the user authentication status returned from geetest into consideration, the website owner follows his own business logic
	// 响应json数据如：{"result": "success", "reason": "", "captcha_args": {}}
	// respond to json data, such as {"result": "success", "reason": "", "captcha_args": {}}

	if err = json.Unmarshal(res_json, &res_map); err != nil {
		fmt.Println("Json数据解析错误")
		return nil
	}

	result := res_map["result"]
	if result == "success" {
		//fmt.Println("极验验证通过")
		return nil
	} else {
		//fmt.Printf("极验验证失败: %v", res_map)
		return errors.New("验证码错误")
	}

}

// hmac-sha256 加密：  CAPTCHA_KEY,lot_number
// hmac-sha256 encrypt: CAPTCHA_KEY, lot_number
func hmac_encode(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func K快验_极验验证码结果验证(验证码id, 验证码内容 string) error {
	if global.Q快验.K快验Api_极验验证码结果验证(验证码id, 验证码内容) {
		return nil
	}
	return 系统错误.New("验证码错误")
}
