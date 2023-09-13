package utils

import (
	"regexp"
	"strconv"
)

// 校验密码   "至少八个字符，至少一个字母和一个数字："
func Z正则_校验密码(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`\S{5,18}$`, s)
	//fmt.Println("Z正则_校验密码$s", a)

	if !匹配结果 {
		*msg = "长度在5-18之间，只能包含字母、数字和下划线等非空白字符"
	}
	return 匹配结果
}

func Z正则_校验代理用户名(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`^[a-zA-Z0-9\p{Han}]+$`, s)
	if !匹配结果 {
		*msg = "只能包含英文字母、数字和ANSI编码支持的中文字符"
	}
	return 匹配结果
}

// Z正则_校验用户名  "至少6个字符,支持数字大小写字母"
func Z正则_校验用户名(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`\w{5,17}$`, s)
	if 匹配结果 {
		return 匹配结果
	}
	*msg = "长度在6-18之间，只能包含字符、数字和下划线"
	return 匹配结果
}

// Z正则_校验email "非正确email格式"
func Z正则_校验email(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`^\w+([-+.]\w+)*@\w+([-.]\w+)*\.\w+([-.]\w+)*$`, s)
	if 匹配结果 {
		return 匹配结果
	}
	*msg = "非正确email格式"
	return 匹配结果
}

// Z正则_校验纯数字 "非完全是数字"
func Z正则_校验纯数字(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`^-?\d*\.?\d+$`, s)
	if 匹配结果 {
		return 匹配结果
	}
	*msg = "非完全是数字"
	return 匹配结果
}

// Z正则_校验纯数字指定位数   "长度不为"+位数
func Z正则_校验纯数字指定位数(s string, msg *string, 位数 int) bool {
	匹配结果, _ := regexp.MatchString(`^\d{`+strconv.Itoa(位数)+`}$`, s)

	if 匹配结果 {
		return 匹配结果
	}
	*msg = "长度不为" + strconv.Itoa(位数)
	return 匹配结果
}

// Z正则_校验手机号 "非正确手机号格式"
func Z正则_校验手机号(s string, msg *string) bool {
	匹配结果, _ := regexp.MatchString(`1[3,4,5,6,7,8,9]\d[\s,-]?\d{4}[\s,-]?\d{4}`, s)
	if 匹配结果 {
		return 匹配结果
	}
	*msg = "非正确手机号格式"
	return 匹配结果
}

// Z正则_是否英数
func Z正则_是否英数(s string, msg *string) bool {
	// 定义正则表达式
	reg := regexp.MustCompile("^[A-Za-z0-9]+$")
	// 使用正则表达式进行匹配
	if reg.MatchString(s) {
		return true
	}
	*msg = "只能输入数字字母"
	return false
}
