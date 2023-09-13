package main

import (
	"context"
	"github.com/tencentyun/cos-go-sdk-v5"
	"net/http"
	"net/url"
	"server/KuaiYanSDK"
	"testing"
)

var 全_Ky KuaiYanSDK.Api快验_类

func TestName(t *testing.T) {

	/*	局_结果 := 全_Ky.C初始化配置(`{"AppWeb":"http://127.0.0.1:18888/Api?AppId=10001","CryptoKeyPublic":"-----BEGIN PUBLIC KEY-----\nMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCVp+zp4GDurEFlDGXu1gpLwgnp\nJCxQpv9WBhdlYwQcHVTOM4fHA54nFP6EWNEBPXUqYRnkAcY0GFPpVg94tCTcGYhN\nM1YE1xdt20wCnHBigp6Ftp+2LUDec73d5qqsSXYvuRj44c6sINxFlqpGsqwv/5GF\nKQA2DDRtFvgodMDSwwIDAQAB\n-----END PUBLIC KEY-----\n","CryptoType":3}`)
		全_Ky.Z置验证码信息(1, "111", "2222")
		局_结果 = 全_Ky.Q取Token()
		var 局_临时文本 string
		fmt.Printf("用户登录_通用%v,%v", 全_Ky.D登录_通用(&局_临时文本, "aaaaaa", "ssssss", "", "暂时没有动态标记", "1.0.2"), 局_临时文本)
		fmt.Printf("Q取服务器连接状态:%v", 全_Ky.Q取服务器连接状态())

		if !局_结果 {
			fmt.Printf("错误信息:%s", 全_Ky.Q取错误信息(nil))
		}*/

	/*var json = `{
	  "htmlurl": "www.baidu.com",
	  "data": [
	    {
	      "WenJianMin": "飞鸟快验",
	      "md5": "55724ed289815775ab76b5284b0df279",
	      "Lujing": "/",
	      "size": "49934239",
	      "url": "https://mangguo-updata-1251700534.cos.ap-guangzhou.myqcloud.com/%E9%A3%9E%E9%B8%9F%E5%BF%AB%E9%AA%8C/%E9%A3%9E%E9%B8%9F%E5%BF%AB%E9%AA%8C%E7%AE%A1%E7%90%86%E7%B3%BB%E7%BB%9F1.0.0",
	      "YunXing": "1"
	    },
	    {
	      "WenJianMin": "思维导图",
	      "md5": "",
	      "Lujing": "/",
	      "url": "https://mangguo-updata-1251700534.cos.ap-guangzhou.myqcloud.com/%E9%A3%9E%E9%B8%9F%E5%BF%AB%E9%AA%8C/%E5%90%8E%E5%8F%B0%E7%BB%93%E6%9E%84%E6%80%9D%E7%BB%B4%E5%AF%BC%E5%9B%BE.emmx",
	      "YunXing": "0"
	    }
	  ]
	}`*/
	//KuaiYanUpdater.K快验系统开始更新(json, a更新下载成功回调)

	腾讯对象存储下载()
}

func 腾讯对象存储下载() {
	var urla = "https://mangguo-updata-1251700534.cos.ap-guangzhou.myqcloud.com/%E9%A3%9E%E9%B8%9F%E5%BF%AB%E9%AA%8C/%E9%A3%9E%E9%B8%9F%E5%BF%AB%E9%AA%8C1.0.2"
	u, _ := url.Parse(urla)
	b := &cos.BaseURL{BucketURL: u}
	c := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  "AKIDGNGtHTR9cpnWzDCvPdcLp4artgDakeJi",
			SecretKey: "CAtNbaJn10jDSSCgvue8NM8evxjGXZLs",
		},
	})
	name := u.Path

	// 2.获取对象到本地文件
	_, err := c.Object.GetToFile(context.Background(), name, "C:\\Users\\x1056\\AppData\\Local\\Temp\\GoLand\\111.aaa", nil)
	if err != nil {
		panic(err)
	}
}
