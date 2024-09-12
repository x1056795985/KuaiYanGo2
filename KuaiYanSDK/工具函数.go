package KuaiYanSDK

import (
	"crypto"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/valyala/fastjson"
	"log"
	"math/big"
	"net/url"
	"server/utils"
	"sort"
	"strings"
	"time"
)

func (k *Api快验_类) 通讯(postJson map[string]interface{}) (*fastjson.Value, bool) {
	局_文本 := k.发包并返回解密(k.加密并签名(postJson))
	//fmt.Printf("响应明文数据:%s\n", 局_文本)
	if 局_文本 == "" {
		// 直接返回即可,错误原因 在 发包并返回解密 已经有了
		return nil, false
	}

	响应, err := fastjson.Parse(局_文本)

	if err != nil || !k.X响应校验时间状态(postJson, 响应) { //如果结果解析失败,或时间,状态码错误,直接返回假
		return nil, false
	}
	return 响应, true
}

/*
 */
func (k *Api快验_类) X响应校验时间状态(请求 map[string]interface{}, 响应 *fastjson.Value) bool {

	if 请求["Status"].(int64) != 响应.GetInt64("Status") {
		k.集_错误代码 = 响应.GetInt("Status")
		k.集_错误信息 = string(响应.GetStringBytes("Msg"))
		return false
	}
	局_耗时 := 请求["Time"].(int64) - 响应.GetInt64("Time")

	if 局_耗时 > 1800 { // ' 发包和收包时间超过30分钟,就有点不对了吧,和服务器时间再差也不会差这么多
		if string(响应.GetStringBytes("Msg")) == "" {
			k.集_错误信息 = "封包时间异常"
			k.集_错误代码 = 107
		} else {
			k.集_错误代码 = 响应.GetInt("Status")
			k.集_错误信息 = string(响应.GetStringBytes("Msg"))
		}
	}
	return true
}

func (k *Api快验_类) 加密并签名(postJson map[string]interface{}) string {
	//添加公共变量
	postJson["Time"] = time.Now().Unix()
	局_随机数, _ := rand.Int(rand.Reader, big.NewInt(89999))
	postJson["Status"] = 局_随机数.Int64() + 10000
	局_Api := postJson["Api"].(string) // 如果没有Api 会报恐慌

	// 验证码id可能会缓存.所以必须是有验证码值时才携带 ,并清除缓存
	if k.集_验证码值 != "" && k.集_验证码ID != "" {
		postJson["Captcha"] = Captcha{
			Type:  k.集_验证码类型,
			Id:    k.集_验证码ID,
			Value: k.集_验证码值,
		}
		k.集_验证码类型 = 0
		k.集_验证码ID = ""
		k.集_验证码值 = ""
	}
	局_明文, _ := json.Marshal(postJson)
	//fmt.Printf("发送明文数据:" + string(局_明文))
	if k.集_CryptoType == 1 {
		// 都明文了,还加密什么,签名什么啊 直接返回就好了
		return string(局_明文)
	}

	var 局_密文, 局_签名 string
	if k.集_CryptoType == 3 && strings.Contains(强制Rsa加密接口, `"`+局_Api+`"`) {
		/*' 看看接口是否在强制RSA通讯列表里,重要包RSA通讯,保证安全,不重要封包Aes通讯,保证服务器性能,
		' 不过即使使用了RSA加密不重要的包,也是可以解开的,服务器会自动判断,但是应该RSA通讯的封包使用了AES通讯服务器会风控报错加解密方式错误
		' 简单来说就是封包安全可以向上兼容,不重要的封包可以使用更安全的RSA通讯方式,但是重要封包不能使用不安全AES的通讯方式*/
		//RSA 就是随机AES密钥  然后RSA公钥加密放到签名里
		局_随机AES密钥 := make([]byte, 24)
		rand.Read(局_随机AES密钥)
		// #对称算法_AES_192_CBC, #数据填充_PKCS7_PADDING, iv取空白字节集 (24))
		局_临时字节集 := utils.Aes加密_cbc192密匙字节数组(string(局_明文), 局_随机AES密钥)
		局_密文 = base64.StdEncoding.EncodeToString(局_临时字节集)
		//公钥加密aes密匙,
		局_签名 = k.rsa公钥加密(局_随机AES密钥)
	} else {
		//非明文 非 rsa 通用加密
		// 正常aes加密即可
		局_临时字节集 := utils.Aes加密_cbc192密匙字节数组(string(局_明文), k.J_CryptoKeyAes)
		局_密文 = base64.StdEncoding.EncodeToString(局_临时字节集)
		// 注意不是md5(明文)  是base64密文文本的md5  //base64 可以防止gbkUtf8编码不同导致的签名不正确
		局_签名 = utils.Md5String(局_密文 + string(k.J_CryptoKeyAes))
	}
	//fmt.Printf(`{"a":"%s","b":"%s"}`, 局_密文, 局_签名)
	return fmt.Sprintf(`{"a":"%s","b":"%s"}`, 局_密文, 局_签名)
}

type Captcha struct {
	Type  int    `json:"Type"`
	Id    string `json:"Id"`
	Value string `json:"Value"`
}
type Result struct {
	Data string `json:"data"`
}

func (k *Api快验_类) 发包并返回解密(post内容 string) string {
	var 局_返回 string
	//耗时 := time.Now().UnixMilli()
	client := req.C().SetTimeout(60 * time.Second).EnableInsecureSkipVerify() //.DevMode()

	//网管 应用认证
	//client.WrapRoundTripFunc(k.腾讯云Api网关认证中间件)
	client.WrapRoundTripFunc(k.阿里云Api网关认证中间件)
	//循环3次 容错处理
	for I := 0; I < 3; I++ {
		resp, _ := client.R().
			SetHeader("Token", k.J_Token).
			SetBodyString(post内容).
			Post(k.集_AppWeb)
		局_返回 = resp.String()
		if 局_返回 != "" {
			//不为空就跳出
			break
		}
		fmt.Printf("Time:%s:通讯异常容错计次:%d\n", time.Now().String(), I)
		time.Sleep(2 * time.Second) // 延迟2秒在重试
	}
	//fmt.Printf("通讯耗时:%d\n", time.Now().UnixMilli()-耗时)
	if 局_返回 == "" {
		k.集_错误信息 = "请求失败,可能以下原因,电脑时间不准确或无网络连接，请校准时间后尝试重启软件."
		k.集_错误代码 = 0
		return ""
	}

	响应, err := fastjson.Parse(局_返回)
	if err != nil { //' 解析失败,直接按错误处理,可能是404 服务器中间件不通之类的
		k.集_错误信息 = 局_返回
		k.集_错误代码 = 0
		return ""
	}

	if 响应.GetInt("Time") > 0 { //明文直接返回
		return 局_返回
	}
	局_密文 := string(响应.GetStringBytes("a"))
	var 局_AES密匙 []byte
	if k.集_CryptoType == 3 {
		局_签名 := string(响应.GetStringBytes("b"))
		if len(局_签名) > 32 { //如果长度大于32 那就不是签名而是密文AES加密密匙,需要通过RSA公钥解密
			局_AES密匙, err = base64.StdEncoding.DecodeString(局_签名)
			if err != nil {
				k.集_错误信息 = "RSA解密失败"
				k.集_错误代码 = 0
				return ""
			}
			局_AES密匙 = k.RSA公钥解密(局_AES密匙)
		} else {
			if strings.ToUpper(utils.Md5String(局_密文+string(k.J_CryptoKeyAes))) != strings.ToUpper(局_签名) {
				k.集_错误信息 = "验签不通过"
				k.集_错误代码 = 0
				return ""
			}
			局_AES密匙 = k.J_CryptoKeyAes
		}
	} else {
		局_AES密匙 = k.J_CryptoKeyAes
	}
	return utils.Aes解密_cbc192字节集(编码_base64解码(局_密文), 局_AES密匙)
}

func (k *Api快验_类) rsa公钥加密(加密内容 []byte) (base64密文 string) {
	if k.集_公钥指针 == nil {
		return ""
	}
	//4、公钥加密
	cipherData, err := rsa.EncryptPKCS1v15(rand.Reader, k.集_公钥指针, 加密内容)
	if err != nil {
		return ""
	}
	base64密文 = base64.StdEncoding.EncodeToString(cipherData)
	return base64密文

}

func (k *Api快验_类) RSA公钥解密(密文 []byte) (明文字节集 []byte) {

	明文字节集, err := publicDecrypt(k.集_公钥指针, crypto.Hash(0), nil, 密文)
	if err != nil {
		return 明文字节集
	}
	return 明文字节集
}

func 编码_base64编码(字节集 []byte) string {
	return base64.StdEncoding.EncodeToString(字节集)

}
func 编码_base64解码(字符串 string) []byte {
	局_临时字节集, err := base64.StdEncoding.DecodeString(字符串)
	if err == nil {
		return 局_临时字节集
	}
	return []byte{}
}

// copy&modified from crypt/rsa/pkcs1v5.go
func publicDecrypt(pub *rsa.PublicKey, hash crypto.Hash, hashed []byte, sig []byte) (out []byte, err error) {
	hashLen, prefix, err := pkcs1v15HashInfo(hash, len(hashed))
	if err != nil {
		return nil, err
	}

	tLen := len(prefix) + hashLen
	k := (pub.N.BitLen() + 7) / 8
	if k < tLen+11 {
		return nil, fmt.Errorf("length illegal")
	}

	c := new(big.Int).SetBytes(sig)
	m := encrypt(new(big.Int), pub, c)
	em := leftPad(m.Bytes(), k)
	out = unLeftPad(em)

	err = nil
	return
}

// copy from crypt/rsa/pkcs1v5.go
func pkcs1v15HashInfo(hash crypto.Hash, inLen int) (hashLen int, prefix []byte, err error) {
	// Special case: crypto.Hash(0) is used to indicate that the data is
	// signed directly.
	if hash == 0 {
		return inLen, nil, nil
	}

	hashLen = hash.Size()
	if inLen != hashLen {
		return 0, nil, errors.New("crypto/rsa: input must be hashed message")
	}
	prefix, ok := hashPrefixes[hash]
	if !ok {
		return 0, nil, errors.New("crypto/rsa: unsupported hash function")
	}
	return
}

// copy from crypt/rsa/pkcs1v5.go
func leftPad(input []byte, size int) (out []byte) {
	n := len(input)
	if n > size {
		n = size
	}
	out = make([]byte, size)
	copy(out[len(out)-n:], input)
	return
}
func unLeftPad(input []byte) (out []byte) {
	n := len(input)
	t := 2
	for i := 2; i < n; i++ {
		if input[i] == 0xff {
			t = t + 1
		} else {
			if input[i] == input[0] {
				t = t + int(input[1])
			}
			break
		}
	}
	out = make([]byte, n-t)
	copy(out, input[t:])
	return
}

// copy from crypt/rsa/pkcs1v5.go
var hashPrefixes = map[crypto.Hash][]byte{
	crypto.MD5:       {0x30, 0x20, 0x30, 0x0c, 0x06, 0x08, 0x2a, 0x86, 0x48, 0x86, 0xf7, 0x0d, 0x02, 0x05, 0x05, 0x00, 0x04, 0x10},
	crypto.SHA1:      {0x30, 0x21, 0x30, 0x09, 0x06, 0x05, 0x2b, 0x0e, 0x03, 0x02, 0x1a, 0x05, 0x00, 0x04, 0x14},
	crypto.SHA224:    {0x30, 0x2d, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x04, 0x05, 0x00, 0x04, 0x1c},
	crypto.SHA256:    {0x30, 0x31, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x01, 0x05, 0x00, 0x04, 0x20},
	crypto.SHA384:    {0x30, 0x41, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x02, 0x05, 0x00, 0x04, 0x30},
	crypto.SHA512:    {0x30, 0x51, 0x30, 0x0d, 0x06, 0x09, 0x60, 0x86, 0x48, 0x01, 0x65, 0x03, 0x04, 0x02, 0x03, 0x05, 0x00, 0x04, 0x40},
	crypto.MD5SHA1:   {}, // A special TLS case which doesn't use an ASN1 prefix.
	crypto.RIPEMD160: {0x30, 0x20, 0x30, 0x08, 0x06, 0x06, 0x28, 0xcf, 0x06, 0x03, 0x00, 0x31, 0x04, 0x14},
}

// copy from crypt/rsa/pkcs1v5.go
func encrypt(c *big.Int, pub *rsa.PublicKey, m *big.Int) *big.Int {
	e := big.NewInt(int64(pub.E))
	c.Exp(m, e, pub.N)
	return c
}
func (k *Api快验_类) 腾讯云Api网关认证中间件(rt req.RoundTripper) req.RoundTripFunc {
	return func(req *req.Request) (resp *req.Response, err error) {
		if len(k.集_Api网关ApiAppSecret) == 0 || k.集_Api网关ApiAppKey == "" {
			resp, err = rt.RoundTrip(req)
			return
		}
		// before request
		// 应用 ApiAppKey
		//const ApiAppKey = "APID3uzfbf4phbgshae2yMg04h1W0s7BpjXqF646"
		//应用 ApiAppSecret
		//const ApiAppSecret = "72Jg5ODOv1sRvskQZde99I3nKr4hf33y6529Qsc"

		const Accept = "application/json"
		const ContentType = "application/json"

		req.Headers.Set("Source", "RSA2")
		req.Headers.Set("Accept", Accept)
		if req.Headers.Get("Token") == "" { //防止没有这个参数
			req.Headers.Set("Token", "")
		}

		// 根据 Url 解析 Host 和 Path
		u, err := url.Parse(req.URL.String())
		if err != nil {
			log.Fatal(err)
		}
		Host := u.Hostname()
		Path := u.Path
		Query := u.RawQuery

		// 签名path不带环境信息
		if strings.HasPrefix(Path, "/release") {
			Path = strings.TrimPrefix(Path, "/release")
		} else if strings.HasPrefix(Path, "/test") {
			Path = strings.TrimPrefix(Path, "/test")
		} else if strings.HasPrefix(Path, "/prepub") {
			Path = strings.TrimPrefix(Path, "/prepub")
		}

		if Path == "" {
			Path = "/"
		}

		// 拼接query参数，query参数需要按字典序排序
		if len(Query) > 0 {
			args, _ := url.ParseQuery(Query)
			var keys []string
			for c := range args {
				keys = append(keys, c)
			}
			sort.Strings(keys)
			sortQuery := ""
			for _, c := range keys {
				if args[c][0] != "" {
					sortQuery = sortQuery + "&" + c + "=" + args[c][0]
				} else {
					sortQuery = sortQuery + "&" + c
				}
			}
			sortQuery = strings.TrimPrefix(sortQuery, "&")

			Path = Path + "?" + sortQuery
		}

		// 获取当前 UTC
		xDate := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
		ContentMD5 := ""
		if req.Method == "POST" {
			h := md5.New()
			h.Write(req.Body)
			md5Str := hex.EncodeToString(h.Sum(nil))
			ContentMD5 = base64.StdEncoding.EncodeToString([]byte(md5Str))
		}

		// 构造签名
		signingStr := fmt.Sprintf("source: %s\ntoken: %s\nx-date: %s\n%s\n%s\n%s\n%s\n%s", req.Headers.Get("Source"), req.Headers.Get("Token"), xDate, req.Method, Accept, ContentType,
			ContentMD5, Path)
		mac := hmac.New(sha1.New, k.集_Api网关ApiAppSecret)

		_, err = mac.Write([]byte(signingStr))
		if err != nil {
			log.Fatal(err)
		}
		signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))
		sign := fmt.Sprintf("hmac id=\"%s\", algorithm=\"hmac-sha1\", headers=\"source token x-date\", signature=\"%s\"",
			k.集_Api网关ApiAppKey, signature)

		req.Headers.Set("Host", Host)
		req.Headers.Set("Accept", Accept)
		req.Headers.Set("Content-Type", ContentType)
		req.Headers.Set("x-date", xDate)
		req.Headers.Set("Authorization", sign)
		// 构造请求

		resp, err = rt.RoundTrip(req)
		// after response
		// ...
		return
	}
}
func (k *Api快验_类) 阿里云Api网关认证中间件(rt req.RoundTripper) req.RoundTripFunc {
	return func(req *req.Request) (resp *req.Response, err error) {
		if len(k.集_Api网关ApiAppSecret) == 0 || k.集_Api网关ApiAppKey == "" {
			resp, err = rt.RoundTrip(req)
			return
		}

		_ = aliyunSign(req, k.集_Api网关ApiAppKey, string(k.集_Api网关ApiAppSecret))

		// 构造请求

		resp, err = rt.RoundTrip(req)
		// after response
		// ...
		return
	}
}
