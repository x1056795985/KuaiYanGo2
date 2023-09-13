package utils

import (
	"math/rand"
	"strings"
	"time"
)

// 文本_是否存在关键字  关键字为空 直接返回 真
func W文本_是否包含关键字(内容, 关键字 string) bool {
	return strings.Contains(内容, 关键字)
}

// 文本取出中间文本
func W文本_取出中间文本(内容 string, 左边文本 string, 右边文本 string) string {
	左边位置 := strings.Index(内容, 左边文本)
	if 左边位置 == -1 {
		return ""
	}
	左边位置 = 左边位置 + len(左边文本)
	内容 = string([]byte(内容)[左边位置:])

	var 右边位置 int
	if 右边文本 == "" {
		右边位置 = len(内容)
	} else {
		右边位置 = strings.Index(内容, 右边文本)
		if 右边位置 == -1 {
			return ""
		}
	}
	内容 = string([]byte(内容)[:右边位置])
	return 内容
}

func W文本_取随机字符串(字符串长度 int) string {
	var strByte = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890")
	var strByteLen = len(strByte)
	bytes := make([]byte, 字符串长度)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 字符串长度; i++ {
		bytes[i] = strByte[r.Intn(strByteLen-1)]
	}
	if bytes[0] == strByte[strByteLen-1] { //第一位不能是0 防止意外
		bytes[0] = strByte[strByteLen-2]
	}

	return string(bytes)
}

func W文本_取随机字符串_数字(字符串长度 int) string {
	var strByte = []byte("1234567890")
	var strByteLen = len(strByte)
	bytes := make([]byte, 字符串长度)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 字符串长度; i++ {
		bytes[i] = strByte[r.Intn(strByteLen-1)]
	}
	if bytes[0] == strByte[strByteLen-1] { //第一位不能是0 防止意外
		bytes[0] = strByte[strByteLen-2]
	}

	return string(bytes)
}

// 调用格式： 〈文本型数组〉 分割文本 （文本型 待分割文本，［文本型 用作分割的文本］，［整数型 要返回的子文本数目］） - 系统核心支持库->文本操作
// 英文名称：split
// 将指定文本进行分割，返回分割后的一维文本数组。本命令为初级命令。
// 参数<1>的名称为“待分割文本”，类型为“文本型（text）”。如果参数值是一个长度为零的文本，则返回一个空数组，即没有任何成员的数组。
// 参数<2>的名称为“用作分割的文本”，类型为“文本型（text）”，可以被省略。参数值用于标识子文本边界。如果被省略，则默认使用半角逗号字符作为分隔符。如果是一个长度为零的文本，则返回的数组仅包含一个成员，即完整的“待分割文本”。
// 参数<3>的名称为“要返回的子文本数目”，类型为“整数型（int）”，可以被省略。如果被省略，则默认返回所有的子文本。
//
// 操作系统需求： Windows、Linux
func W文本_分割文本(待分割文本 string, 用作分割的文本 string) []string {
	return strings.Split(待分割文本, 用作分割的文本)
}
