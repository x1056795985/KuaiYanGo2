package KuaiYanUpdater

import (
	"EFunc/utils"
	"crypto/md5"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/tencentyun/cos-go-sdk-v5"
	"github.com/valyala/fastjson"
	"net/http"
	"path/filepath"

	"strings"
	"time"
)

var J_系统更新状态 = 0 //0未更新 1 更新中 2更新失败 3下载成功
var J_系统更新提示 = ""
var J_任务列表 []更新文件列表
var 集_运行目录 = ""

func K快验系统开始更新(更新文件文本 string, 更新成功后处理程序 func(执行程序本地路径 string)) {
	if J_系统更新状态 != 0 {
		fmt.Println("已经在更新程序了,不要再调用了")
		return
	}
	J_系统更新状态 = 1
	集_运行目录 = utils.C程序_取运行目录()
	/*	if runtime.GOOS == "windows" {
		集_运行目录 = "."
	}*/
	局_json, err := fastjson.Parse(更新文件文本)
	if err != nil {
		J_系统更新提示 = "更新失败,请重试"
		J_系统更新状态 = 2
		return
	}
	var 执行程序路径 = ""
	局_文件 := 局_json.GetArray("data")
	J_任务列表 = make([]更新文件列表, len(局_文件))
	for 索引 := range 局_文件 {
		/*		{
				"WenJianMin":"文件名.exe",
				"md5":"e10adc3949ba59abbe56e057f20f883e(小写文件md5可选,有就校验,空就只校验文件名)",
				"Lujing":"\/(下载本地相对路径)",
				"size":"12345",
				"url":"https:\/\/www.baidu.com\/文件名.exe(下载路径)",
				"YunXing":"1(值为更新完成后会运行这个文件,只能有一个文件值为1)"
			}*/
		局_临时文件名 := string(局_文件[索引].GetStringBytes("WenJianMin"))
		if 局_临时文件名 == "" {
			局_临时文件名 = string(局_文件[索引].GetStringBytes("url"))
			局_临时文件名 = filepath.Base(局_临时文件名) //取出路径文件名
		}

		var 局_本地路径 string
		局_本地路径 = 集_运行目录
		if string(局_文件[索引].GetStringBytes("Lujing")) == "" {
			局_本地路径 += "/"
		} else {
			局_本地路径 += string(局_文件[索引].GetStringBytes("Lujing"))
		}
		局_本地路径 += 局_临时文件名
		if utils.W文件_是否存在(集_运行目录 + string(局_文件[索引].GetStringBytes("Lujing"))) {
			_ = utils.M目录_创建(集_运行目录 + string(局_文件[索引].GetStringBytes("Lujing")))
		}
		J_任务列表[索引] = 更新文件列表{
			本地文件名:  局_本地路径,
			远程下载地址: string(局_文件[索引].GetStringBytes("url")),
			更新结束后是否需要自动执行该文件: string(局_文件[索引].GetStringBytes("YunXing")) == "1",
			是否已下载: false,
		}
		if J_任务列表[索引].更新结束后是否需要自动执行该文件 {
			执行程序路径 = J_任务列表[索引].本地文件名
		}
		J_系统更新提示 = "正在读取并校验更新文件，请稍候....."

		局_临时文件MD5 := strings.ToUpper(string(局_文件[索引].GetStringBytes("md5")))
		if 局_临时文件MD5 != "" { // 有md5 就校验,没有就文件名校验
			局_本地文件MD5 := ""
			data := utils.W文件_读入文件(局_本地路径) //切片
			if data != nil {
				has := md5.Sum(data)
				局_本地文件MD5 = strings.ToUpper(fmt.Sprintf("%x", has)) //将[]byte转成16进制

				if 局_临时文件MD5 == 局_本地文件MD5 {
					J_任务列表[索引].是否已下载 = true //文件已经存在直接跳过
					continue                //到循环尾
				}
			}
		} else if utils.W文件_是否存在(局_本地路径) { //不推荐文件名,可能会出现不准确的情况
			J_任务列表[索引].是否已下载 = true //文件已经存在直接跳过
		}

	}

	if len(J_任务列表) == 0 {
		goto 标签_更新成功

	}
	//开始下载列表
	for 索引 := range J_任务列表 {
		if J_任务列表[索引].是否已下载 { //不需要下载
			continue //到循环尾
		}

		callback := func(info req.DownloadInfo) {
			if info.Response.Response != nil {
				J_系统更新提示 = fmt.Sprintf("下载更新中:%v/%v,已下载: %.2f%%\n", 索引+1, len(J_任务列表), float64(info.DownloadedSize)/float64(info.Response.ContentLength)*100.0)
				fmt.Printf(J_系统更新提示)
			}
		}

		client := req.C() //.DevMode()

		transport := client.GetTransport()
		transport.WrapRoundTripFunc(func(rt http.RoundTripper) req.HttpRoundTripFunc {
			return func(req *http.Request) (resp *http.Response, err error) {
				// before request
				// ...
				req.Header.Add("x-cos-traffic-limit", "10485760") //限速1280kb/s

				//	权限只读id和key
				secretID := "AKIDGNGtHTR9cpnWzDCvPdcLp4artgDakeJi"
				secretKey := "CAtNbaJn10jDSSCgvue8NM8evxjGXZLs"
				startTime := time.Unix(time.Now().Unix()-3600, 0)
				endTime := time.Unix(time.Now().Unix()+36000, 0) //有效期 1小时
				authTime := &cos.AuthTime{
					SignStartTime: startTime,
					SignEndTime:   endTime,
					KeyStartTime:  startTime,
					KeyEndTime:    endTime,
				}

				cos.AddAuthorizationHeader(secretID, secretKey, "", req, authTime)
				resp, err = rt.RoundTrip(req)
				// after response
				// ...
				return
			}
		})

		_, err1 := client.R().
			SetOutputFile(J_任务列表[索引].本地文件名).
			SetDownloadCallback(callback).
			Get(J_任务列表[索引].远程下载地址)

		if err1 != nil {
			J_系统更新提示 = "文件:" + J_任务列表[索引].远程下载地址 + ",下载失败," + err1.Error()
			J_系统更新状态 = 2
			fmt.Println(J_系统更新提示)
			J_任务列表 = make([]更新文件列表, 0)
			return
		}

	}
标签_更新成功:
	J_系统更新提示 = "下载成功"

	//J_任务列表 = make([]更新文件列表, 0) // 全部先下载成功 更新成功的不要清除,不然无法判断是否更新了
	J_系统更新状态 = 3
	if 更新成功后处理程序 != nil {
		J_系统更新提示 = "等待重启"
		更新成功后处理程序(执行程序路径)
	}

}

type 更新文件列表 struct {
	本地文件名            string
	远程下载地址           string
	是否已下载            bool
	更新结束后是否需要自动执行该文件 bool //注意，该数据类型中此成员为真只允许有一个。建议设为主程序。
}
