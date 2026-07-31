package item

import "testing"

func TestQ七牛云_Q取文件信息_未初始化(t *testing.T) {
	局_测试对象 := []*Q七牛云{nil, {}}
	for _, 局_对象 := range 局_测试对象 {
		if _, 局_错误 := 局_对象.Q取文件信息(nil, "test.txt"); 局_错误 == nil {
			t.Fatal("未初始化的七牛云对象应返回错误")
		}
	}
}
