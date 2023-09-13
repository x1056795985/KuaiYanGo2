package utils

import (
	"bytes"
	E "github.com/duolabmeng6/goefun/eTool"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)
import . "github.com/duolabmeng6/efun/efun"

func W网页_取域名(Url string) string {
	var 域名 string
	if E取文本左边(Url, 8) == "https://" {
		域名 = E.E文本_取出中间文本(Url, "https://", "/")
	}
	if E取文本左边(Url, 7) == "http://" {
		域名 = E.E文本_取出中间文本(Url, "http://", "/")
	}
	return 域名
}

func 网页_访问_对象(网址 string, 访问方式 int, 提交信息 string, 提交Cookies string, 返回Cookies *string, 附加协议头 string, 返回协议头 *string, 返回状态代码 *int, 禁止重定向 bool, 字节集提交 []byte, 代理地址 string, 超时 int, 代理用户名 string, 代理密码 string, 代理标识 int, 对象继承 interface{}, 是否自动合并更新Cookie bool, 是否补全必要协议头 bool, 是否处理协议头大小写 bool) []byte {
	client := &http.Client{}
	if 超时 != -1 {
		if 超时 < 1 {
			超时 = 15
		}
		client.Timeout = time.Duration(超时) * time.Second
	}

	if 代理地址 != "" {
		proxyURL, _ := url.Parse(代理地址)
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	var method string
	switch 访问方式 {
	case 0:
		method = "GET"
	case 1:
		method = "POST"
	case 2:
		method = "HEAD"
	case 3:
		method = "PUT"
	case 4:
		method = "OPTIONS"
	case 5:
		method = "DELETE"
	case 6:
		method = "TRACE"
	case 7:
		method = "CONNECT"
	default:
		method = "GET"
	}

	req, err := http.NewRequest(method, 网址, bytes.NewBuffer(字节集提交))
	if err != nil {
		return []byte{}
	}

	if 禁止重定向 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	if 附加协议头 != "" {
		协议头列表 := strings.Split(附加协议头, "\n")
		for _, 协议头 := range 协议头列表 {
			if strings.TrimSpace(协议头) != "" {
				头名, 头值 := 内部_协议头取名值(协议头)
				req.Header.Set(头名, 头值)
			}
		}
	}

	if 提交Cookies != "" {
		req.Header.Set("Cookie", 提交Cookies)
	}

	resp, err := client.Do(req)
	if err != nil {
		return []byte{}
	}
	defer resp.Body.Close()

	返回状态代码 = &resp.StatusCode
	*返回协议头 = resp.Header.Get("Content-Type")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}
	}

	if 返回Cookies != nil {
		*返回Cookies = resp.Header.Get("Set-Cookie")
	}

	return body
}

func 内部_协议头取名值(协议头 string) (string, string) {
	头名值 := strings.SplitN(协议头, ":", 2)
	if len(头名值) == 2 {
		return 头名值[0], 头名值[1]
	}
	return "", ""
}
