package utils

import "github.com/gogf/gf/v2/util/gconv"

// 字节数组
func D到字节集(value interface{}) []byte {
	return gconv.Bytes(value)
}
func D到字节(value interface{}) byte {
	return gconv.Byte(value)
}
func D到整数(value interface{}) int {
	return gconv.Int(value)
}

func D到数值(value interface{}) float64 {
	return gconv.Float64(value)
}
func D到文本(value interface{}) string {
	return gconv.String(value)
}
func D到结构体(待转换的参数 interface{}, 结构体指针 interface{}) error {
	return gconv.Struct(待转换的参数, 结构体指针)
}
