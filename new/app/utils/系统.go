package utils

import (
	"EFunc/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func X系统_权限检测() bool {
	//检查自身进程是否有写入数据的权限
	// 这个例子测试写权限，如果没有写权限则返回error。
	// 注意文件不存在也会返回error，需要检查error的信息来获取到底是哪个错误导致。

	utils.W文件_删除(GetCurrentAbPathByExecutable() + "/权限测试.json")

	f, err := os.Create(GetCurrentAbPathByExecutable() + "/权限测试.json")
	if err != nil {
		return false
	}
	defer f.Close()
	f.WriteString(time.Now().String())
	// 获取当前权限
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	fmt.Printf("文件权限 %v\n", fi.Mode())
	return true
}

// 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func Y易支付取设备类型(ua string) string {
	// pc	电脑浏览器
	// mobile	手机浏览器
	// qq	手机QQ内浏览器
	// wechat	微信内浏览器
	// alipay	支付宝客户端
	// jump	仅返回支付跳转url

	// 转换为小写以便匹配
	ua = strings.ToLower(ua)

	// 定义各类客户端的关键词
	mobileKeywords := []string{
		"mobile", "android", "iphone", "ipod", "ipad", "windows phone",
	}

	wechatKeywords := []string{
		"micromessenger",
	}

	qqKeywords := []string{
		"qq/", "qzone/",
	}

	alipayKeywords := []string{
		"alipayclient",
	}

	// 检查是否为支付宝客户端
	for _, keyword := range alipayKeywords {
		if strings.Contains(ua, keyword) {
			return "alipay" // alipay
		}
	}

	// 检查是否为微信浏览器
	for _, keyword := range wechatKeywords {
		if strings.Contains(ua, keyword) {
			return "wechat" // wechat
		}
	}

	// 检查是否为QQ浏览器
	for _, keyword := range qqKeywords {
		if strings.Contains(ua, keyword) {
			return "qq" // qq
		}
	}

	// 检查是否为移动设备
	for _, keyword := range mobileKeywords {
		if strings.Contains(ua, keyword) {
			return "mobile" // mobile
		}
	}

	// 默认返回false，表示PC或其他未识别设备
	return "pc"
}
