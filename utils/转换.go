package utils

import (
	"github.com/shopspring/decimal"
	"strconv"
)

// 防止精度丢失
func Float64取绝对值(值 float64) float64 {
	var 最终 float64
	if 值 < 0 {
		局_精确 := decimal.NewFromFloat(值)
		局_精确乘数 := decimal.NewFromInt(-1)
		最终, _ = 局_精确.Mul(局_精确乘数).Float64()
	} else {
		最终 = 值
	}
	return 最终
}

// 防止精度丢失
func Float64乘int64(值1 float64, 值2 int64) float64 {
	var 最终 float64
	局_精确 := decimal.NewFromFloat(值1)
	局_精确乘数 := decimal.NewFromInt(值2)
	最终, _ = 局_精确.Mul(局_精确乘数).Float64()

	return 最终
}

// 防止精度丢失
func Float64乘Float64(值1 float64, 值2 float64) float64 {
	var 最终 float64
	局_精确 := decimal.NewFromFloat(值1)
	局_精确乘数 := decimal.NewFromFloat(值2)
	最终, _ = 局_精确.Mul(局_精确乘数).Float64()

	return 最终
}

// 防止精度丢失
func Float64除int64(值1 int64, 值2 int64, 保留长度 int32) float64 {
	var 最终 float64
	局_精确 := decimal.NewFromInt(值1)
	局_精确除数 := decimal.NewFromInt(值2)
	最终, _ = 局_精确.Div(局_精确除数).Round(保留长度).Float64()

	return 最终
}

// 防止精度丢失
func Float64取负值(值 float64) float64 {
	var 最终 float64
	if 值 > 0 {
		局_精确 := decimal.NewFromFloat(值)
		局_精确乘数 := decimal.NewFromInt(-1)
		最终, _ = 局_精确.Mul(局_精确乘数).Float64()
	} else {
		最终 = 值
	}
	return 最终
}
func Float64到文本(值 float64, 保留小数点多少位 int) string {

	return strconv.FormatFloat(值, 'f', 保留小数点多少位, 64)
}
