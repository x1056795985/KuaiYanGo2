package main

import (
	. "EFunc/utils"
	"bytes"
	"crypto/md5"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

func Test_项目编译(t *testing.T) {
	编译飞鸟快验()
}
func 编译飞鸟快验() {
	局_项目路径 := "E:\\yun\\project\\TY通用后台管理系统\\server2" //必须  \\ 路径间隔,不然打开文件夹路径错误
	局_源码 := string(W文件_读入文件(局_项目路径 + "/global/global.go"))
	局_旧版本号 := W文本_取出中间文本(局_源码, "B版本号当前: \"", "\",")
	局_临时数组 := strings.Split(局_旧版本号, ".")
	if len(局_临时数组) != 3 {
		fmt.Printf("版本号读取失败")
		return
	}
	局编译版本号, _ := strconv.Atoi(局_临时数组[2])
	局_临时数组[2] = strconv.Itoa(局编译版本号 + 1)
	局_新版本号 := strings.Join(局_临时数组, ".")
	局_源码 = strings.Replace(局_源码, "B版本号当前: \""+局_旧版本号+"\",", "B版本号当前: \""+局_新版本号+"\",", 1)
	// 保存修改后的源码文件
	err := W文件_写出文件(局_项目路径+"/global/global.go", []byte(局_源码))
	if err != nil {
		fmt.Println("保存修改后的源码文件失败:", err)
		return
	}
	局_编译名称 := "飞鸟快验" + 局_新版本号 + ".bin"
	//设置编译为linux
	cmd := exec.Command("go", "env", "-w", "GOOS=linux")
	cmd.Dir = 局_项目路径
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("命令执行失败:"+err.Error(), stderr.String())
		return
	} else {
		fmt.Println(out.String())
	}
	//设置编译使用gcc
	/*	cmd = exec.Command("go", "env", "-w", "CGO_ENABLED=0")
		cmd.Dir = 局_项目路径
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			fmt.Println("命令执行失败:"+err.Error(), stderr.String())
			return
		} else {
			fmt.Println(out.String())
		}
	*/
	// 设置环境变量
	// 设置环境变量
	cmd = exec.Command("go", "build", "-o", 局_编译名称, "main.go")
	cmd.Dir = 局_项目路径

	// 获取当前的 GOPATH 和 GOMODCACHE
	goPath := os.Getenv("GOPATH")
	goModCache := os.Getenv("GOMODCACHE")

	// 如果 GOMODCACHE 未设置，使用默认值
	if goModCache == "" && goPath != "" {
		// GOMODCACHE 默认在 GOPATH/pkg/mod
		goModCache = goPath + "/pkg/mod"
	}

	// 构建环境变量列表
	envs := []string{
		"GOOS=linux",
		"GOARCH=amd64",
		"CGO_ENABLED=0",
		"PATH=" + os.Getenv("PATH"),
		"GOCACHE=" + os.TempDir() + "/gocache", // 添加这一行
	}

	// 只有在值不为空时才添加 GOPATH 和 GOMODCACHE
	if goPath != "" {
		envs = append(envs, "GOPATH="+goPath)
	}
	if goModCache != "" {
		envs = append(envs, "GOMODCACHE="+goModCache)
	}

	cmd.Env = envs

	out.Reset()
	stderr.Reset()
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("命令执行失败:", err.Error())
		fmt.Println("stderr:", stderr.String()) // 打印详细的错误信息
		return
	}
	fmt.Println("编译命令执行成功:" + 局_编译名称)
	cmd = exec.Command("E:\\yun1\\e\\工具\\upx-4.0.2-win64\\upx.exe", 局_项目路径+"\\"+局_编译名称)
	err = cmd.Run()
	if err != nil {
		fmt.Println("压缩执行失败:", err)
		return
	}
	fmt.Println("压缩命令执行成功")
	// 执行命令打开文件夹并选中文件
	cmd = exec.Command("explorer", "/select,", 局_项目路径+"\\"+局_编译名称)
	cmd.Start()

	fmt.Println("打开文件夹并选中文件成功")

	data := W文件_读入文件(局_项目路径 + "\\" + 局_编译名称) //切片
	has := md5.Sum(data)
	局_本地文件MD5 := strings.ToUpper(fmt.Sprintf("%x", has)) //将[]byte转成16进制
	fmt.Println("MD5 :  " + 局_本地文件MD5)
}
