package jsEngine

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sync"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"

	"server/app/logic/common/publicJs"
	dbm "server/app/models/db"
	"server/app/service"
)

const 脚本引擎_最大编译程序数 = 256

var 集_Hook程序缓存 = 脚本引擎_程序缓存{程序表: make(map[[sha256.Size]byte]*goja.Program)}

type 脚本引擎_程序缓存 struct {
	互斥锁 sync.RWMutex
	程序表 map[[sha256.Size]byte]*goja.Program
}

func (j *脚本引擎_程序缓存) 编译(source string) (*goja.Program, error) {
	局_键 := sha256.Sum256([]byte(source))
	j.互斥锁.RLock()
	局_程序 := j.程序表[局_键]
	j.互斥锁.RUnlock()
	if 局_程序 != nil {
		return 局_程序, nil
	}

	局_已编译程序, 局_错误 := goja.Compile("hook.js", source, false)
	if 局_错误 != nil {
		return nil, 局_错误
	}
	j.互斥锁.Lock()
	if 局_已有程序 := j.程序表[局_键]; 局_已有程序 != nil {
		局_已编译程序 = 局_已有程序
	} else {
		if len(j.程序表) >= 脚本引擎_最大编译程序数 {
			clear(j.程序表)
		}
		j.程序表[局_键] = 局_已编译程序
	}
	j.互斥锁.Unlock()
	return 局_已编译程序, nil
}

// J脚本引擎_处理任务池Hook 执行任务池 Hook 并返回转换后的任务数据和状态。
func J脚本引擎_处理任务池Hook(c *gin.Context, appInfo *dbm.DB_AppInfo, online *dbm.DB_LinksToken, hookName, taskData string, status int) (string, int, error) {
	if c == nil {
		c = 脚本引擎_后台上下文()
	}
	局_公共JS, 局_错误 := publicJs.L_publicJs.P取值2(c, service.Js类型_任务池Hook函数, hookName)
	if 局_错误 != nil {
		return "", status, 局_错误
	}
	局_运行时 := J脚本引擎_初始化用户(c, appInfo, online, &局_公共JS)
	_ = 局_运行时.Set("$拦截原因", "")
	_ = 局_运行时.Set("$任务状态", status)
	局_输出, 局_错误 := 脚本引擎_执行字符串Hook(局_运行时, 局_公共JS.Value, 局_公共JS.Name, taskData)
	if 局_错误 != nil {
		return "", status, 局_错误
	}
	if 局_值 := 局_运行时.Get("$任务状态"); 局_值 != nil && !goja.IsUndefined(局_值) && !goja.IsNull(局_值) {
		status = int(局_值.ToInteger())
	}
	if 局_原因 := 脚本引擎_取运行时字符串(局_运行时, "$拦截原因"); 局_原因 != "" {
		return "", status, errors.New(局_原因)
	}
	return 局_输出, status, nil
}

// J脚本引擎_处理ApiHook 执行 API Hook，并安全地将返回值转换为文本。
func J脚本引擎_处理ApiHook(appInfo *dbm.DB_AppInfo, online *dbm.DB_LinksToken, hookName, plaintext string, c *gin.Context) (返回文本 string, 错误 error) {
	defer func() {
		if 局_异常 := recover(); 局_异常 != nil {
			错误 = fmt.Errorf("js函数错误: %v", 局_异常)
		}
	}()
	if c == nil {
		c = 脚本引擎_后台上下文()
	}
	局_公共JS, 局_错误 := publicJs.L_publicJs.P取值2(c, service.Js类型_ApiHook函数, hookName)
	if 局_错误 != nil {
		return plaintext, 局_错误
	}
	局_运行时 := J脚本引擎_初始化用户(c, appInfo, online, &局_公共JS)
	_ = 局_运行时.Set("$拦截原因", "")
	返回文本, 错误 = 脚本引擎_执行字符串Hook(局_运行时, 局_公共JS.Value, 局_公共JS.Name, plaintext)
	if 错误 != nil {
		return plaintext, 错误
	}
	if 局_原因 := 脚本引擎_取运行时字符串(局_运行时, "$拦截原因"); 局_原因 != "" {
		return plaintext, errors.New(局_原因)
	}
	return 返回文本, nil
}

func 脚本引擎_执行字符串Hook(vm *goja.Runtime, source, functionName, input string) (string, error) {
	局_程序, 局_错误 := 集_Hook程序缓存.编译(source)
	if 局_错误 != nil {
		return "", fmt.Errorf("JS代码编译失败: %w", 局_错误)
	}
	if _, 局_错误 = vm.RunProgram(局_程序); 局_错误 != nil {
		return "", fmt.Errorf("JS代码运行失败: %w", 局_错误)
	}
	局_函数, 局_存在 := goja.AssertFunction(vm.Get(functionName))
	if !局_存在 {
		return "", fmt.Errorf("Js中没有[%s()]函数", functionName)
	}
	局_值, 局_错误 := 局_函数(goja.Undefined(), vm.ToValue(input))
	if 局_错误 != nil {
		return "", fmt.Errorf("JS函数执行失败: %w", 局_错误)
	}
	局_结果, 局_类型正确 := 局_值.Export().(string)
	if !局_类型正确 {
		return "", fmt.Errorf("JS函数[%s]必须返回字符串", functionName)
	}
	return 局_结果, nil
}

func 脚本引擎_取运行时字符串(vm *goja.Runtime, name string) string {
	局_值 := vm.Get(name)
	if 局_值 == nil || goja.IsUndefined(局_值) || goja.IsNull(局_值) {
		return ""
	}
	return 局_值.String()
}
