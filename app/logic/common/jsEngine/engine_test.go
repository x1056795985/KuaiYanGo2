package jsEngine

import (
	"crypto/sha256"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/dop251/goja"
	"github.com/gin-gonic/gin"

	"server/app/global"
	dbm "server/app/models/db"
)

func Test脚本引擎_运行时注册全部绑定(t *testing.T) {
	局_运行时 := J脚本引擎_初始化用户(nil, &dbm.DB_AppInfo{AppId: 10001, AppName: "test"}, &dbm.DB_LinksToken{}, &dbm.DB_PublicJs{})
	for _, 局_绑定 := range 集_静态绑定 {
		局_值 := 局_运行时.Get(局_绑定.名称)
		if 局_值 == nil || goja.IsUndefined(局_值) || goja.IsNull(局_值) {
			t.Errorf("binding %q was not registered", 局_绑定.名称)
		}
	}
	if _, 局_可调用 := goja.AssertFunction(局_运行时.Get("$api_短信发送")); !局_可调用 {
		t.Fatal("SMS API is not callable")
	}
}

func Test脚本引擎_执行字符串Hook(t *testing.T) {
	局_运行时 := goja.New()
	局_结果, 局_错误 := 脚本引擎_执行字符串Hook(局_运行时, `function transform(value) { return value + "-ok"; }`, "transform", "input")
	if 局_错误 != nil {
		t.Fatalf("脚本引擎_执行字符串Hook() error = %v", 局_错误)
	}
	if 局_结果 != "input-ok" {
		t.Fatalf("脚本引擎_执行字符串Hook() = %q, want input-ok", 局_结果)
	}
}

func Test脚本引擎_拒绝无效Hook函数(t *testing.T) {
	局_测试数组 := []struct {
		名称 string
		源码 string
		预期 string
	}{
		{名称: "missing", 源码: `var other = 1`, 预期: "没有[transform()]函数"},
		{名称: "wrong return type", 源码: `function transform() { return 123; }`, 预期: "必须返回字符串"},
		{名称: "exception", 源码: `function transform() { throw new Error("bad"); }`, 预期: "JS函数执行失败"},
	}
	for _, 局_测试 := range 局_测试数组 {
		t.Run(局_测试.名称, func(t *testing.T) {
			_, 局_错误 := 脚本引擎_执行字符串Hook(goja.New(), 局_测试.源码, "transform", "input")
			if 局_错误 == nil || !strings.Contains(局_错误.Error(), 局_测试.预期) {
				t.Fatalf("error = %v, want containing %q", 局_错误, 局_测试.预期)
			}
		})
	}
}

func Test脚本引擎_程序缓存并发编译(t *testing.T) {
	const 局_源码 = `function transform(value) { return value + "-ok"; }`
	局_缓存 := 脚本引擎_程序缓存{程序表: make(map[[sha256.Size]byte]*goja.Program)}

	const 局_并发数 = 64
	var 局_等待组 sync.WaitGroup
	局_错误通道 := make(chan error, 局_并发数)
	for range 局_并发数 {
		局_等待组.Add(1)
		go func() {
			defer 局_等待组.Done()
			_, 局_错误 := 局_缓存.编译(局_源码)
			局_错误通道 <- 局_错误
		}()
	}
	局_等待组.Wait()
	close(局_错误通道)

	for 局_错误 := range 局_错误通道 {
		if 局_错误 != nil {
			t.Fatalf("编译() error = %v", 局_错误)
		}
	}
	局_缓存.互斥锁.RLock()
	defer 局_缓存.互斥锁.RUnlock()
	if len(局_缓存.程序表) != 1 {
		t.Fatalf("compiled program count = %d, want 1", len(局_缓存.程序表))
	}
}

func Test脚本引擎_请求快照(t *testing.T) {
	局_请求 := httptest.NewRequest("POST", "http://example.test/path?a=1", strings.NewReader("b=2"))
	局_请求.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	局_请求.Header.Set("X-Test", "value")
	if 局_错误 := 局_请求.ParseForm(); 局_错误 != nil {
		t.Fatal(局_错误)
	}
	局_上下文 := &gin.Context{Request: 局_请求}
	局_快照 := 脚本引擎_请求快照(局_上下文)
	if 局_快照["Method"] != "POST" || 局_快照["Host"] != "example.test" {
		t.Fatalf("unexpected request snapshot: %#v", 局_快照)
	}
	局_表单 := 局_快照["Form"].(map[string]any)
	if 局_表单["a"] != "1" || 局_表单["b"] != "2" {
		t.Fatalf("unexpected form snapshot: %#v", 局_表单)
	}
}

func Test脚本引擎_拒绝模块目录穿越(t *testing.T) {
	局_原运行目录 := global.GVA_CONFIG.Q取运行目录
	global.GVA_CONFIG.Q取运行目录 = t.TempDir()
	t.Cleanup(func() { global.GVA_CONFIG.Q取运行目录 = 局_原运行目录 })

	if _, _, 局_有效 := 脚本引擎_解析模块路径("../../secret.js"); 局_有效 {
		t.Fatal("path traversal was accepted")
	}
	局_第一路径, 局_是否远程, 局_有效 := 脚本引擎_解析模块路径("https://example.test/module.js")
	if !局_有效 || !局_是否远程 || !strings.HasSuffix(局_第一路径, ".js") {
		t.Fatalf("unexpected remote module path: %q, remote=%v, ok=%v", 局_第一路径, 局_是否远程, 局_有效)
	}
	局_第二路径, _, _ := 脚本引擎_解析模块路径("https://example.test/module.js")
	if 局_第一路径 != 局_第二路径 {
		t.Fatal("remote module cache path is not stable")
	}
}

func Benchmark脚本引擎_运行时初始化(b *testing.B) {
	局_应用信息 := &dbm.DB_AppInfo{AppId: 10001, AppName: "benchmark"}
	局_在线信息 := &dbm.DB_LinksToken{Uid: 1, User: "benchmark"}
	局_公共JS := &dbm.DB_PublicJs{}
	b.ReportAllocs()
	for range b.N {
		_ = J脚本引擎_初始化用户(nil, 局_应用信息, 局_在线信息, 局_公共JS)
	}
}

func Benchmark脚本引擎_已编译Hook执行(b *testing.B) {
	const 局_源码 = `function transform(value) { return value + "-ok"; }`
	_, _ = 集_Hook程序缓存.编译(局_源码)
	b.ReportAllocs()
	for range b.N {
		if _, 局_错误 := 脚本引擎_执行字符串Hook(goja.New(), 局_源码, "transform", "input"); 局_错误 != nil {
			b.Fatal(局_错误)
		}
	}
}
