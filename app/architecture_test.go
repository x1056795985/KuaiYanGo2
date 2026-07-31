package app

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var 集_控制器直接数据库基线 = map[string]int{
	"admin/AgentInventory.go":    5,
	"admin/AgentUser.go":         2,
	"admin/AgentUserFull.go":     6,
	"admin/AppFull.go":           7,
	"admin/AppUserFull.go":       3,
	"admin/InitDB.go":            2,
	"admin/KaFull.go":            2,
	"admin/LinkUser.go":          3,
	"admin/Login.go":             2,
	"admin/LogRMBPayOrder.go":    4,
	"admin/PublicData.go":        2,
	"admin/PublicJs.go":          4,
	"admin/TaskPoolFull.go":      8,
	"admin/User.go":              7,
	"admin/UserConfig.go":        2,
	"agent/AgentInventoryOld.go": 3,
	"agent/AgentKa.go":           2,
	"agent/AgentUserCompat.go":   4,
	"agent/Base.go":              2,
	"agent/Menu.go":              1,
	"userSafetyApi/任务池.go":       1,
	"webApi/Ka.go":               1,
	"webUser/base.go":            2,
}

var 集_控制器Gorm导入白名单 = map[string]bool{
	"admin/AppUserFull.go":  true,
	"admin/InitDB.go":       true,
	"agent/AppUser.go":      true,
	"userSafetyApi/绑定信息.go": true,
}

func Test架构_分层依赖(t *testing.T) {
	_, 局_当前文件, _, 局_成功 := runtime.Caller(0)
	if !局_成功 {
		t.Fatal("无法确定架构测试目录")
	}
	局_app目录 := filepath.Dir(局_当前文件)

	t.Run("事务只能由Logic开启", func(t *testing.T) {
		架构_遍历Go文件(t, 局_app目录, func(相对路径 string, 文件 *ast.File) {
			if strings.HasPrefix(相对路径, "logic/") {
				return
			}
			ast.Inspect(文件, func(节点 ast.Node) bool {
				局_调用, 局_是调用 := 节点.(*ast.CallExpr)
				if !局_是调用 {
					return true
				}
				局_选择器, 局_是选择器 := 局_调用.Fun.(*ast.SelectorExpr)
				if 局_是选择器 && (局_选择器.Sel.Name == "Transaction" || 局_选择器.Sel.Name == "Begin") {
					t.Errorf("%s 不得开启事务，事务应迁入 logic", 相对路径)
				}
				return true
			})
		})
	})

	t.Run("基础包不得反向依赖业务层", func(t *testing.T) {
		架构_遍历Go文件(t, 局_app目录, func(相对路径 string, 文件 *ast.File) {
			局_检查global := strings.HasPrefix(相对路径, "global/")
			局_检查bootstrap := strings.HasPrefix(相对路径, "bootstrap/") || 相对路径 == "main.go"
			if !局_检查global && !局_检查bootstrap {
				return
			}
			for _, 局_导入 := range 文件.Imports {
				局_导入路径, 局_错误 := strconv.Unquote(局_导入.Path.Value)
				if 局_错误 != nil {
					continue
				}
				if 局_检查global && strings.HasPrefix(局_导入路径, "server/app/logic/") {
					t.Errorf("%s 不得反向导入 logic: %s", 相对路径, 局_导入路径)
				}
				if 局_检查bootstrap && strings.HasPrefix(局_导入路径, "server/app/controller/") {
					t.Errorf("%s 不得直接导入 controller: %s", 相对路径, 局_导入路径)
				}
				if 局_导入路径 == "server/app/init" || strings.HasPrefix(局_导入路径, "server/app/init/") {
					t.Errorf("%s 仍引用已迁移的 init 包", 相对路径)
				}
			}
		})
	})

	t.Run("控制器直接Gorm只减不增", func(t *testing.T) {
		局_控制器目录 := filepath.Join(局_app目录, "controller")
		架构_遍历Go文件(t, 局_控制器目录, func(相对路径 string, 文件 *ast.File) {
			局_直接调用数 := 0
			ast.Inspect(文件, func(节点 ast.Node) bool {
				局_调用, 局_是调用 := 节点.(*ast.CallExpr)
				if !局_是调用 {
					return true
				}
				局_方法, 局_是方法 := 局_调用.Fun.(*ast.SelectorExpr)
				if !局_是方法 {
					return true
				}
				局_数据库, 局_是数据库 := 局_方法.X.(*ast.SelectorExpr)
				if !局_是数据库 || 局_数据库.Sel.Name != "GVA_DB" {
					return true
				}
				局_包名, 局_是包名 := 局_数据库.X.(*ast.Ident)
				if 局_是包名 && 局_包名.Name == "global" {
					局_直接调用数++
				}
				return true
			})
			if 局_直接调用数 > 集_控制器直接数据库基线[相对路径] {
				t.Errorf("%s 直接调用 GVA_DB %d 次，超过遗留基线 %d", 相对路径, 局_直接调用数, 集_控制器直接数据库基线[相对路径])
			}
			for _, 局_导入 := range 文件.Imports {
				局_导入路径, _ := strconv.Unquote(局_导入.Path.Value)
				if 局_导入路径 == "gorm.io/gorm" && !集_控制器Gorm导入白名单[相对路径] {
					t.Errorf("%s 新增了 Controller 到 GORM 的直接依赖", 相对路径)
				}
			}
		})
	})
}

func 架构_遍历Go文件(t *testing.T, 根目录 string, 检查 func(相对路径 string, 文件 *ast.File)) {
	t.Helper()
	局_文件集 := token.NewFileSet()
	局_错误 := filepath.WalkDir(根目录, func(路径 string, 条目 fs.DirEntry, 遍历错误 error) error {
		if 遍历错误 != nil {
			return 遍历错误
		}
		if 条目.IsDir() || !strings.HasSuffix(条目.Name(), ".go") || strings.HasSuffix(条目.Name(), "_test.go") {
			return nil
		}
		局_文件, 局_解析错误 := parser.ParseFile(局_文件集, 路径, nil, 0)
		if 局_解析错误 != nil {
			return 局_解析错误
		}
		局_相对路径, 局_相对路径错误 := filepath.Rel(根目录, 路径)
		if 局_相对路径错误 != nil {
			return 局_相对路径错误
		}
		检查(filepath.ToSlash(局_相对路径), 局_文件)
		return nil
	})
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
}
