package jsEngine

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"server/app/global"
)

type 缓存_测试实现 struct {
	互斥锁 sync.RWMutex
	数据  map[string]any
}

func 测试缓存_新建() *缓存_测试实现 {
	return &缓存_测试实现{数据: make(map[string]any)}
}

func (j *缓存_测试实现) Set(key string, value any, _ time.Duration) {
	j.互斥锁.Lock()
	j.数据[key] = value
	j.互斥锁.Unlock()
}

func (j *缓存_测试实现) Get(key string) (any, bool) {
	j.互斥锁.RLock()
	局_值, 局_存在 := j.数据[key]
	j.互斥锁.RUnlock()
	return 局_值, 局_存在
}

func (j *缓存_测试实现) Delete(key string) {
	j.互斥锁.Lock()
	delete(j.数据, key)
	j.互斥锁.Unlock()
}

func (j *缓存_测试实现) Increment(string, int64) error        { return nil }
func (j *缓存_测试实现) Add(string, any, time.Duration) error { return nil }

func Test脚本引擎_合并Cookie(t *testing.T) {
	局_结果 := 脚本引擎_合并Cookie("b=old;a=1;", []*http.Cookie{{Name: "b", Value: "new"}, {Name: "c", Value: "3"}})
	if 局_预期 := "a=1;b=new;c=3;"; 局_结果 != 局_预期 {
		t.Fatalf("脚本引擎_合并Cookie() = %q, want %q", 局_结果, 局_预期)
	}
}

func Test脚本引擎_缓存拒绝异常类型(t *testing.T) {
	局_原缓存 := global.H缓存
	局_缓存 := 测试缓存_新建()
	global.H缓存 = 局_缓存
	t.Cleanup(func() { global.H缓存 = 局_原缓存 })
	局_缓存.Set(脚本引擎_缓存键前缀+"key", 123, 0)
	if 局_结果 := 脚本引擎_取缓存("key"); 局_结果 != "" {
		t.Fatalf("脚本引擎_取缓存() = %q, want empty string", 局_结果)
	}
}
