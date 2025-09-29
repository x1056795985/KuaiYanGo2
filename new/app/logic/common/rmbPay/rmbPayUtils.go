package rmbPay

import (
	"EFunc/utils"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"server/Service/Ser_RMBPayOrder"
	"server/new/app/models/common"
	"server/new/app/models/constant"
	"strconv"
	"strings"
	"time"
)

// 生成18位订单号  线程安全
// 年前两位月日时分秒0001计数 每秒999999订单内没问题
func (j *rmbPay) 获取新订单号() string {

	j.锁.Lock() //加锁
	当前时间戳 := time.Now().Unix()
	if 当前时间戳 == j.订单号时间戳 {
		j.订单号计数++
		if j.订单号计数 > 999999 {
			j.订单号计数 = 1
		}
	} else {
		j.订单号时间戳 = 当前时间戳
		j.订单号初始值 = utils.H汇编_取随机数(100, 899999) //随机一个初始值,防止每次都从1开始
		j.订单号计数 = j.订单号初始值 + 1
	}
	局_计数 := j.订单号计数
	j.锁.Unlock() //解锁
	//获取当前年后两位月日时分秒 组成订单号前缀
	var 最终订单号 = time.Unix(当前时间戳, 0).Format("20060102150405")
	最终订单号 = 最终订单号[2:] //删除年左侧20两位

	if 局_计数 == j.订单号初始值 {
		//如果相等说明当前秒内,已经超过了99999 并重置为1 后递增到j.订单号初始值 的,所以实际已经达到999999了次调用了
		fmt.Println("恭喜生成订单号大于每秒100w建议更换算法")
		return ""
	}
	最终订单号 = 最终订单号 + strings.Repeat("0", 6-len(strconv.Itoa(局_计数))) + strconv.Itoa(局_计数)
	return 最终订单号
}

func (j *rmbPay) S生成二维码并转base64(内容 string) string {
	局_二维码base64 := ""
	png, err := qrcode.Encode(内容, qrcode.Medium, 256)
	if err == nil {
		局_二维码base64 = base64.StdEncoding.EncodeToString(png)
	}
	return 局_二维码base64
}

func (j *rmbPay) Q取提示信息(参数 *common.PayParams) string {

	if 参数.User == "" && 参数.ProcessingType == constant.D订单类型_支付购卡 {
		return "支付购卡:" + 参数.User + "_" + Ser_RMBPayOrder.C处理类型[参数.ProcessingType]
	}

	if 参数.User == "" {
		return "用户不存在"
	}

	return "用户:" + 参数.User + "_" + j.Map订单类型[参数.ProcessingType]
}
func (j *rmbPay) Z支付订单回调关键字转换(回调地址 string, 参数 *common.PayParams) string {
	ReturnURL := strings.Replace(回调地址, "{OrderId}", 参数.PayOrder, -1)
	ReturnURL = strings.Replace(ReturnURL, "{User}", 参数.User, -1)
	ReturnURL = strings.Replace(ReturnURL, "{Type}", 参数.Type, -1)
	ReturnURL = strings.Replace(ReturnURL, "{ProcessingType}", strconv.Itoa(参数.ProcessingType), -1)
	var extra = "{}"
	if 字节数组, err := json.Marshal(参数.E额外信息); err == nil {
		extra = string(字节数组)
	}
	ReturnURL = strings.Replace(ReturnURL, "{Extra}", extra, -1)
	return ReturnURL
}

func (j *rmbPay) Pay_显示名称转原名(显示名称 string) string {
	局_数组 := j.Pay_取支付通道基本信息()

	for i := range 局_数组 {
		if 局_数组[i].Alias == 显示名称 {
			return 局_数组[i].Name
		}
	}
	return 显示名称
}
