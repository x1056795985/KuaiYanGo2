package KuaiYanUpdater

import (
	json2 "encoding/json"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Test宝塔_更新项目数据库(t *testing.T) {
	局_数据库, 局_错误 := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
	局_建表SQL := `CREATE TABLE sites (
		id INTEGER PRIMARY KEY,
		name TEXT NOT NULL,
		path TEXT NOT NULL,
		project_config TEXT NOT NULL,
		project_type TEXT NOT NULL
	)`
	if 局_错误 = 局_数据库.Exec(局_建表SQL).Error; 局_错误 != nil {
		t.Fatal(局_错误)
	}
	局_原配置 := `{"project_exe":"old","project_cmd":"old","project_name":"test"}`
	if 局_错误 = 局_数据库.Exec(
		`INSERT INTO sites (id, name, path, project_config, project_type) VALUES (?, ?, ?, ?, ?)`,
		1, "test", "/old/server", 局_原配置, "Go",
	).Error; 局_错误 != nil {
		t.Fatal(局_错误)
	}

	局_项目信息, 局_错误 := 宝塔_读取Go项目信息(局_数据库)
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
	const 局_新路径 = "/opt/server/server"
	if 局_错误 = 宝塔_更新项目数据库(局_数据库, 局_项目信息, 局_新路径); 局_错误 != nil {
		t.Fatal(局_错误)
	}

	局_更新结果, 局_错误 := 宝塔_读取Go项目信息(局_数据库)
	if 局_错误 != nil {
		t.Fatal(局_错误)
	}
	if 局_更新结果.Path != 局_新路径 {
		t.Fatalf("path = %q, want %q", 局_更新结果.Path, 局_新路径)
	}
	var 局_项目配置 宝塔_项目配置
	if 局_错误 = json2.Unmarshal([]byte(局_更新结果.ProjectConfig), &局_项目配置); 局_错误 != nil {
		t.Fatal(局_错误)
	}
	if 局_项目配置.ProjectExe != 局_新路径 || 局_项目配置.ProjectCmd != 局_新路径 {
		t.Fatalf("project config paths = %q/%q, want %q", 局_项目配置.ProjectExe, 局_项目配置.ProjectCmd, 局_新路径)
	}
}
