package Ser_Ka

import (
	"fmt"
	"testing"
)

func Test_生成随机字符串(t *testing.T) {
	Name := "qq"
	Name += 生成随机字符串(10-len(Name)-1, 1)
	Name += 生成校验字符(Name)
	fmt.Println(Name)
	//fmt.Println(Ka校验卡号(Name))

}
