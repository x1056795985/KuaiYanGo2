// 请求socks代理服务器(已设置IP白名单)
// http和https网页均适用

package main

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"github.com/idoubi/goz"
	"golang.org/x/net/proxy"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

func testaa() {
	testCurl()
	return
	// 演示IP，生产环境中请替换为提取到的代理信息
	proxy_str := "115.223.31.58:24963"

	// 目标网页
	page_url := "https://xui.ptlogin2.qq.com/cgi-bin/xlogin?pt_enable_pwd=1&appid=716027609&pt_3rd_aid=1101817502&daid=381&pt_skey_valid=0&style=35&force_qr=1&autorefresh=1&s_url=http%3A%2F%2Fconnect.qq.com&refer_cgi=m_authorize&ucheck=1&fall_to_wv=1&status_os=11&redirect_uri=auth%3A%2F%2Ftauth.qq.com%2F&client_id=1105412664&pf=openmobile_android&response_type=token&scope=all&sdkp=a&sdkv=3.5.7.lite&sign=f1e1bf720d40eebf1e95d70310bfabf2&status_machine=Redmi+K30+5G+Speed&switch=1&time=1722679468945&show_download_ui=true&h5sig=363wXmww3ftebM39EAw9mOP44xM-8kLXcAB0aJqsnFk&loginty=6"

	// 设置代理
	dialer, err := proxy.SOCKS5("tcp", proxy_str, nil, proxy.Direct)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	// 请求目标网页
	client := &http.Client{Transport: &http.Transport{Dial: dialer.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: false,
		}}}
	req, _ := http.NewRequest("GET", page_url, nil)
	req.Header.Add("Accept-Encoding", "gzip") //使用gzip压缩传输数据让访问更快
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.9")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Cookie", "RK=xdXltRtnTQ; ptcz=7a8475df02ca9e76d967ae73c05c4f5d13534a0e8e5db42190fe867adda129a1; _qimei_uuid42=1851d17362410021666a69fc5c751fbabc95c705d2; pac_uid=0_hYTKnPytbG8sR; _qimei_q36=; _qimei_h38=19c5cc9d666a69fc5c751fba0200000531851d; pgv_pvid=1799300232; _clck=3939281914|1|fmb|0; eas_sid=f1V7v2X1v2d6V4o6T0o9n6L8Y7; wetest_lang=zh-cn; _ga=GA1.1.965724382.1721610668; _ga_0KGGHBND6H=GS1.1.1721610667.1.0.1721610677.0.0.0; pt2gguin=o1056795985; fqm_pvqid=59649e03-ff36-4c21-9daf-2045a3501a95; qq_domain_video_guid_verify=c6848a15de04dab6; _qimei_fingerprint=06214c607b5b6557a9ec59ad378d27e3; _uetvid=925d8a50681111efb092f11845ce8b2c; __aegis_uid=de1e7f0000010681-9c702758e2f44fcc-8839; _qpsvr_localtk=0.6715049313264683; clientuin=1056795985; ETK=; ptnick_1056795985=e69a97e69c88e99a90e890bd; dev_mid_sig=c1ff02f8983989e524b796d8ff23c9511ec929e62c3ed3ffdbe478ffbda2d6d6ba77b9f9f9d43e17; olu=929f1c85edec9eeee1fe0aca696800002927a6dc9cf62399; superuin=o1056795985; supertoken=2171612221; superkey=Qgg2APKEC6yEAS5Bz8ej8b9hpalz7tZH4hRUgiNz86A_; pt_recent_uins=5576d15228fa054d94f85fd82caafc24d27b90f85ac70cc7e6b7c38ed3442030db872d33b4ec5e5526b2c600153caefb1c412d896a9150ad; pt_guid_sig=cdbb3bec56d17a77ba412db359c8babe203844fec1023d19e94879e3a8014aec; pt_login_sig=qeS0egrYyJzK6yltBh5GEgpbiOkw7CxAuAZKMhQJJivIQ76vSbepMW5VHSqLwptW; pt_clientip=8cff2758e2f46389; pt_serverip=a9d37f000001bd34; pt_local_token=1073535076; uikey=463114de553f82e00ff950f3abffaf8f6d8d1c624c9f9aca56852fe9981130a9; dlock=381_1725766196_1_; qrsig=8a372d13f4922787fd7f54bbdf083eaf9addf0954ee0dec7689816a6d531894557c4a2b79c9184c61124f2779a2803210390ab700c378236635f633fc82f4dbc")
	req.Header.Add("Host", "xui.ptlogin2.qq.com")
	req.Header.Add("If-Modified-Since", "Thu, 15 Aug 2024 07:46:00 GMT")
	req.Header.Add("Upgrade-Insecure-Requests", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/122.0.6261.95 Safari/537.36")

	res, err := client.Do(req)

	if err != nil {
		// 请求发生异常
		fmt.Println(err.Error())
	} else {
		defer res.Body.Close() //保证最后关闭Body

		fmt.Println("status code:", res.StatusCode) // 获取状态码

		// 有gzip压缩时,需要解压缩读取返回内容
		if res.Header.Get("Content-Encoding") == "gzip" {
			reader, _ := gzip.NewReader(res.Body) // gzip解压缩
			defer reader.Close()
			io.Copy(os.Stdout, reader)
			os.Exit(0) // 正常退出
		}

		// 无gzip压缩, 读取返回内容
		body, _ := ioutil.ReadAll(res.Body)
		fmt.Println(string(body))
	}
}

func testCurl() {
	cli := goz.NewClient()

	resp, err := cli.Get("https://xui.ptlogin2.qq.com/cgi-bin/xlogin?pt_enable_pwd=1&appid=716027609&pt_3rd_aid=1101817502&daid=381&pt_skey_valid=0&style=35&force_qr=1&autorefresh=1&s_url=http%3A%2F%2Fconnect.qq.com&refer_cgi=m_authorize&ucheck=1&fall_to_wv=1&status_os=11&redirect_uri=auth%3A%2F%2Ftauth.qq.com%2F&client_id=1105412664&pf=openmobile_android&response_type=token&scope=all&sdkp=a&sdkv=3.5.7.lite&sign=f1e1bf720d40eebf1e95d70310bfabf2&status_machine=Redmi+K30+5G+Speed&switch=1&time=1722679468945&show_download_ui=true&h5sig=363wXmww3ftebM39EAw9mOP44xM-8kLXcAB0aJqsnFk&loginty=6", goz.Options{
		Timeout: 5.0,
		Proxy:   "socks5://114.229.203.116:29363",
		Headers: map[string]interface{}{
			"User-Agent": "testing/1.0",
			"Accept":     "application/json",
			"host":       "220.194.117.167",
			"X-Foo":      []string{"Bar", "Baz"},
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	fmt.Println(resp.GetStatusCode())
	body, err := resp.GetBody()
	if err != nil {
		return
	}

	fmt.Println(body.String())
	// Output: 200

}
