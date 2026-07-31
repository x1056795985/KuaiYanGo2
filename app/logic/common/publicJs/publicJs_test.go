package publicJs

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"server/app/global"
	dbm "server/app/models/db"
)

type 公共JS_测试缓存 struct {
	互斥锁 sync.RWMutex
	数据  map[string]any
}

func (j *公共JS_测试缓存) Set(key string, value any, _ time.Duration) {
	j.互斥锁.Lock()
	j.数据[key] = value
	j.互斥锁.Unlock()
}

func (j *公共JS_测试缓存) Get(key string) (any, bool) {
	j.互斥锁.RLock()
	局_值, 局_存在 := j.数据[key]
	j.互斥锁.RUnlock()
	return 局_值, 局_存在
}

func (j *公共JS_测试缓存) Delete(key string) {
	j.互斥锁.Lock()
	delete(j.数据, key)
	j.互斥锁.Unlock()
}

func (j *公共JS_测试缓存) Increment(string, int64) error        { return nil }
func (j *公共JS_测试缓存) Add(string, any, time.Duration) error { return nil }

func Test公共JS_使用稳定文件路径缓存键(t *testing.T) {
	局_数据库, 局_错误 := gorm.Open(sqlite.Open("file:public_js_cache_test?mode=memory&cache=shared"), &gorm.Config{})
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
	if 局_错误 = 局_数据库.AutoMigrate(&dbm.DB_PublicJs{}); 局_错误 != nil {
		t.Fatal(局_错误)
	}
	局_公共JS := dbm.DB_PublicJs{AppId: 2, Name: "cache_test", Value: "/云函数/cache_test.js"}
	if 局_错误 = 局_数据库.Create(&局_公共JS).Error; 局_错误 != nil {
		t.Fatal(局_错误)
	}

	局_运行目录 := t.TempDir()
	局_文件名 := filepath.Join(局_运行目录, "云函数", "cache_test.js")
	if 局_错误 = os.MkdirAll(filepath.Dir(局_文件名), 0o755); 局_错误 != nil {
		t.Fatal(局_错误)
	}
	if 局_错误 = os.WriteFile(局_文件名, []byte("function cache_test() { return 'ok'; }"), 0o600); 局_错误 != nil {
		t.Fatal(局_错误)
	}

	局_原数据库, 局_原缓存, 局_原运行目录 := global.GVA_DB, global.H缓存, global.GVA_CONFIG.Q取运行目录
	global.GVA_DB = 局_数据库
	global.H缓存 = &公共JS_测试缓存{数据: make(map[string]any)}
	global.GVA_CONFIG.Q取运行目录 = 局_运行目录
	t.Cleanup(func() {
		global.GVA_DB, global.H缓存, global.GVA_CONFIG.Q取运行目录 = 局_原数据库, 局_原缓存, 局_原运行目录
	})

	局_首次结果, 局_错误 := L_publicJs.P取值2(&gin.Context{}, 局_公共JS.AppId, 局_公共JS.Name)
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
	if 局_错误 = os.Remove(局_文件名); 局_错误 != nil {
		t.Fatal(局_错误)
	}
	局_二次结果, 局_错误 := L_publicJs.P取值2(&gin.Context{}, 局_公共JS.AppId, 局_公共JS.Name)
	if 局_错误 != nil {
		t.Fatalf("second read did not use cache: %v", 局_错误)
	}
	if 局_首次结果.Value != 局_二次结果.Value {
		t.Fatalf("cached script changed: first=%q second=%q", 局_首次结果.Value, 局_二次结果.Value)
	}
	if _, 局_存在 := global.H缓存.Get(局_文件名); !局_存在 {
		t.Fatalf("script was not cached by file path %q", 局_文件名)
	}
}
