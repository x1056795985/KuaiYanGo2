package realname

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"server/app/logic/common/setting"
)

// 腾讯云实名认证 身份二要素核验(姓名+身份证号) https://cloud.tencent.com/document/api/1007/33188
// 项目未引入faceid服务包,按约定直接用http+TC3-HMAC-SHA256签名调用,签名算法 https://cloud.tencent.com/document/api/213/30654
type 腾讯云实名认证 struct{}

var 集_HTTP客户端 = &http.Client{Timeout: 10 * time.Second}

// 认证 实现I实名认证接口,核验一致(收费结果码Result=0)时返回腾讯云唯一请求标识RequestId
func (j *腾讯云实名认证) R认证(姓名 string, 身份证号 string) (唯一Id string, err error) {
	配置 := setting.Q实名认证配置().TX云实名
	switch {
	case 配置.SECRET_ID == "":
		return "", errors.New("腾讯云实名配置无效SECRET_ID")
	case 配置.SECRET_KEY == "":
		return "", errors.New("腾讯云实名配置无效SECRET_KEY")
	}

	局_请求体, err := json.Marshal(map[string]string{"Name": 姓名, "IdCard": 身份证号})
	if err != nil {
		return "", fmt.Errorf("编码腾讯云实名请求失败: %w", err)
	}

	局_时间戳 := time.Now().Unix()
	局_授权头, err := 腾讯云TC3签名(配置.SECRET_ID, 配置.SECRET_KEY, 局_时间戳, 局_请求体)
	if err != nil {
		return "", err
	}

	局_请求, err := http.NewRequest(http.MethodPost, "https://faceid.tencentcloudapi.com", bytes.NewReader(局_请求体))
	if err != nil {
		return "", fmt.Errorf("创建腾讯云实名请求失败: %w", err)
	}
	局_请求.Header.Set("Content-Type", "application/json")
	局_请求.Header.Set("X-TC-Action", "IdCardVerification")
	局_请求.Header.Set("X-TC-Version", "2018-03-01")
	局_请求.Header.Set("X-TC-Timestamp", fmt.Sprintf("%d", 局_时间戳))
	局_请求.Header.Set("Authorization", 局_授权头)

	局_响应, err := 集_HTTP客户端.Do(局_请求)
	if err != nil {
		return "", fmt.Errorf("腾讯云实名请求失败: %w", err)
	}
	defer 局_响应.Body.Close()
	if 局_响应.StatusCode != http.StatusOK {
		return "", fmt.Errorf("腾讯云实名HTTP状态异常: %s", 局_响应.Status)
	}
	局_响应体, err := io.ReadAll(io.LimitReader(局_响应.Body, 1<<20))
	if err != nil {
		return "", fmt.Errorf("读取腾讯云实名响应失败: %w", err)
	}

	var 局_结果 struct {
		Response struct {
			Result      string `json:"Result"` // 0=一致 -1=不一致 -2非法身份证号 -3非法姓名 -4证件库服务异常 -5证件库中无此记录 -6权威比对系统升级中 -7超当日认证次数限制
			Description string `json:"Description"`
			RequestId   string `json:"RequestId"`
			Error       *struct {
				Code    string `json:"Code"`
				Message string `json:"Message"`
			} `json:"Error"`
		} `json:"Response"`
	}
	if err = json.Unmarshal(局_响应体, &局_结果); err != nil {
		return "", fmt.Errorf("解析腾讯云实名响应失败: %w", err)
	}
	if 局_结果.Response.Error != nil {
		return "", fmt.Errorf("腾讯云实名请求失败: %s - %s", 局_结果.Response.Error.Code, 局_结果.Response.Error.Message)
	}
	if 局_结果.Response.Result != "0" {
		return "", fmt.Errorf("实名认证失败:%s", strings.TrimSpace(局_结果.Response.Description))
	}
	if 局_结果.Response.RequestId == "" {
		return "", errors.New("实名认证成功但平台未返回唯一标识")
	}
	return 局_结果.Response.RequestId, nil
}

// 腾讯云TC3签名 生成实名认证的TC3-HMAC-SHA256授权头,POST JSON固定签名 content-type;host;x-tc-action 三个头
func 腾讯云TC3签名(SECRET_ID string, SECRET_KEY string, 时间戳 int64, 请求体 []byte) (授权头 string, err error) {
	return TC3签名(TC3签名参数{
		SecretId:    SECRET_ID,
		SecretKey:   SECRET_KEY,
		服务:          "faceid",
		主机:          "faceid.tencentcloudapi.com",
		Action:      "IdCardVerification",
		ContentType: "application/json",
		请求体:         请求体,
		时间戳:         时间戳,
	})
}

// TC3签名参数 TC3-HMAC-SHA256签名入参
type TC3签名参数 struct {
	SecretId    string
	SecretKey   string
	服务          string
	主机          string
	Action      string
	ContentType string
	请求体         []byte
	时间戳         int64
	签名头         string // 参与签名的请求头列表,留空默认 content-type;host;x-tc-action(当前官方文档形式)
}

// TC3签名 按腾讯云API 3.0签名规范生成授权头 https://cloud.tencent.com/document/api/213/30654
func TC3签名(参数 TC3签名参数) (授权头 string, err error) {
	const 局_算法名 = "TC3-HMAC-SHA256"
	局_签名头 := 参数.签名头
	if 局_签名头 == "" {
		局_签名头 = "content-type;host;x-tc-action"
	}
	局_日期 := time.Unix(参数.时间戳, 0).UTC().Format("2006-01-02")

	局_请求体哈希 := sha256.Sum256(参数.请求体)

	//拼接规范请求串,请求头值需转小写,头内容必须与签名头声明一致
	局_规范头 := "content-type:" + 参数.ContentType + "\nhost:" + 参数.主机 + "\n"
	if strings.Contains(局_签名头, "x-tc-action") {
		局_规范头 += "x-tc-action:" + strings.ToLower(参数.Action) + "\n"
	}
	局_规范请求串 := strings.Join([]string{
		http.MethodPost,
		"/",
		"",
		局_规范头,
		局_签名头,
		hex.EncodeToString(局_请求体哈希[:]),
	}, "\n")

	局_规范请求串哈希 := sha256.Sum256([]byte(局_规范请求串))
	//拼接签名字符串
	局_签名字符串原文 := strings.Join([]string{
		局_算法名,
		fmt.Sprintf("%d", 参数.时间戳),
		局_日期+"/"+参数.服务+"/tc3_request",
		hex.EncodeToString(局_规范请求串哈希[:]),
	}, "\n")

	//计算派生签名密钥 SecretDate -> SecretService -> SecretSigning
	局_签名密钥 := []byte("TC3" + 参数.SecretKey)
	for _, 局_数据 := range []string{局_日期, 参数.服务, "tc3_request", 局_签名字符串原文} {
		局_摘要 := hmac.New(sha256.New, 局_签名密钥)
		局_摘要.Write([]byte(局_数据))
		局_签名密钥 = 局_摘要.Sum(nil)
	}

	return 局_算法名 + " Credential=" + 参数.SecretId + "/" + 局_日期 + "/" + 参数.服务 + "/tc3_request" +
		", SignedHeaders=" + 局_签名头 + ", Signature=" + hex.EncodeToString(局_签名密钥), nil
}
