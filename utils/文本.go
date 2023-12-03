package utils

import (
	"fmt"
	"github.com/axgle/mahonia"
	"math/rand"
	"strings"
	"time"
	"unicode/utf8"
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

// 获取关键字左边文本
func W文本_取文本左边(内容 string, 关键字 string) string {
	位置 := strings.Index(内容, 关键字)
	if 位置 == -1 {
		return ""
	}

	位置 = 位置 + len(关键字)
	内容 = string([]byte(内容)[位置:])
	return 内容
}

// 获取关键字右边文本
func W文本_取文本右边(内容 string, 关键字 string) string {
	位置 := strings.Index(内容, 关键字)
	if 位置 == -1 {
		return ""
	}
	内容 = string([]byte(内容)[位置+len(关键字):])
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
func W文本_gbk到utf8(src string) string {
	srcDecoder := mahonia.NewDecoder("gbk")
	desDecoder := mahonia.NewDecoder("utf-8")
	resStr := srcDecoder.ConvertString(src)
	_, resBytes, _ := desDecoder.Translate([]byte(resStr), true)
	return string(resBytes)
}

func W文本_utf8到gbk(src string) string {
	srcDecoder := mahonia.NewDecoder("utf-8")
	desDecoder := mahonia.NewDecoder("gbk")
	resStr := srcDecoder.ConvertString(src)
	_, resBytes, _ := desDecoder.Translate([]byte(resStr), true)
	return string(resBytes)
}

// 调用格式： 〈文本型〉 取文本左边 （文本型 欲取其部分的文本，整数型 欲取出字符的数目） - 系统核心支持库->文本操作
// 英文名称：left
// 返回一个文本，其中包含指定文本中从左边算起指定数量的字符。本命令为初级命令。
// 参数<1>的名称为“欲取其部分的文本”，类型为“文本型（text）”。
// 参数<2>的名称为“欲取出字符的数目”，类型为“整数型（int）”。
//
// 操作系统需求： Windows、Linux
func W文本_取左边(欲取其部分的文本 string, 欲取出字符的数目 int) string {
	if len(欲取其部分的文本) < 欲取出字符的数目 {
		欲取出字符的数目 = len(欲取其部分的文本)
	}
	return string([]rune(欲取其部分的文本)[:欲取出字符的数目])
}

//调用格式： 〈文本型〉 取文本右边 （文本型 欲取其部分的文本，整数型 欲取出字符的数目） - 系统核心支持库->文本操作
//英文名称：right
//返回一个文本，其中包含指定文本中从右边算起指定数量的字符。本命令为初级命令。
//参数<1>的名称为“欲取其部分的文本”，类型为“文本型（text）”。
//参数<2>的名称为“欲取出字符的数目”，类型为“整数型（int）”。
//
//操作系统需求： Windows、Linux

func W文本_取右边(欲取其部分的文本 string, 欲取出字符的数目 int) string {
	l := len(欲取其部分的文本)
	lpos := l - 欲取出字符的数目
	if lpos < 0 {
		lpos = 0
	}
	return string([]rune(欲取其部分的文本)[lpos:l])
}

func W文本_删首尾空(内容 string) string {
	return strings.TrimSpace(内容)
}
func W文本_删首空(欲删除空格的文本 string) string {
	return strings.TrimLeft(欲删除空格的文本, " ")
}

//
//调用格式： 〈文本型〉 删尾空 （文本型 欲删除空格的文本） - 系统核心支持库->文本操作
//英文名称：RTrim
//返回一个文本，其中包含被删除了尾部全角或半角空格的指定文本。本命令为初级命令。
//参数<1>的名称为“欲删除空格的文本”，类型为“文本型（text）”。
//
//操作系统需求： Windows、Linux

func W文本_删尾空(欲删除空格的文本 string) string {
	return strings.TrimRight(欲删除空格的文本, " ")
}

func W文本_子文本替换(欲被替换的文本 string, 欲被替换的子文本 string, 用作替换的子文本 string) string {

	return strings.Replace(欲被替换的文本, 欲被替换的子文本, 用作替换的子文本, -1)
}

func W文本_取随机ip() string {
	rand.Seed(time.Now().Unix())
	ip := fmt.Sprintf("%d.%d.%d.%d", rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255))
	return ip
}
func W文本_到大写(value string) string {
	return strings.ToUpper(value)
}

func W文本_到小写(value string) string {
	return strings.ToLower(value)
}

// 中文占多个字节但是这里算一个长度
func W文本_取长度(value string) int {
	return utf8.RuneCountInString(value)
}

// 调用格式： 〈文本型〉 字符 （字节型 欲取其字符的字符代码） - 系统核心支持库->文本操作
// 英文名称：chr
// 返回一个文本，其中包含有与指定字符代码相关的字符。本命令为初级命令。
// 参数<1>的名称为“欲取其字符的字符代码”，类型为“字节型（byte）”。
//
// 操作系统需求： Windows、Linux
func W文本_字符(字节型 int8) string {
	return string(byte(字节型))
}

// 查找关键字位置,失败返回-1
func W文本_寻找文本(被搜寻的文本 string, 欲寻找的文本 string) int {
	return strings.Index(被搜寻的文本, 欲寻找的文本)
}

// 从后往前查找关键字位置,失败返回-1
func W文本_倒找文本(被搜寻的文本 string, 欲寻找的文本 string) int {
	return strings.LastIndex(被搜寻的文本, 欲寻找的文本)
}

// 调用格式： 〈文本型〉 取空白文本 （整数型 重复次数） - 系统核心支持库->文本操作
// 英文名称：space
// 返回具有指定数目半角空格的文本。本命令为初级命令。
// 参数<1>的名称为“重复次数”，类型为“整数型（int）”。
//
// 操作系统需求： Windows、Linux
func W文本_取空白(重复次数 int) string {
	var str string
	for i := 0; i < 重复次数; i++ {
		str = str + " "
	}
	return str
}

//调用格式： 〈文本型〉 取重复文本 （整数型 重复次数，文本型 待重复文本） - 系统核心支持库->文本操作
//英文名称：string
//返回一个文本，其中包含指定次数的文本重复结果。本命令为初级命令。
//参数<1>的名称为“重复次数”，类型为“整数型（int）”。
//参数<2>的名称为“待重复文本”，类型为“文本型（text）”。该文本将用于建立返回的文本。如果为空，将返回一个空文本。
//
//操作系统需求： Windows、Linux

func W文本_取重复(重复次数 int, 待重复文本 string) string {
	var str string
	for i := 0; i < 重复次数; i++ {
		str = str + 待重复文本
	}
	return str
}

func W文本_取行数(文本 string) int {
	lineCount := strings.Count(文本, "\n") + 1
	return lineCount
}
