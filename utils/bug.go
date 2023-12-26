package utils

import (
	"reflect"
	"runtime"
)

// 获取当前包名 结构体名称,方法名称
func Q取包名结构体方法(结构体 interface{}) string {
	// 使用反射获取当前包名
	pkgPath := reflect.TypeOf(结构体).PkgPath()
	//fmt.Println("当前包名:", pkgPath)

	// 使用反射获取结构体名称
	structName := reflect.TypeOf(结构体).Name()
	//fmt.Println("结构体名称:", structName)

	// 使用runtime包获取当前方法的名称
	pc, _, _, _ := runtime.Caller(0)
	funcName := runtime.FuncForPC(pc).Name()
	//fmt.Println("方法名:", funcName)
	return pkgPath + "->" + structName + "->" + funcName
}
