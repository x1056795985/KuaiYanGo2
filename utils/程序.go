package utils

import (
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// 取运行目录
func C程序_取运行目录() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res

}

//调用格式： 〈逻辑型〉 运行 （文本型 欲运行的命令行，逻辑型 是否等待程序运行完毕，［整数型 被运行程序窗口显示方式］） - 系统核心支持库->系统处理
//英文名称：run
//本命令运行指定的可执行文件或者外部命令。如果成功，返回真，否则返回假。本命令为初级命令。
//参数<1>的名称为“欲运行的命令行”，类型为“文本型（text）”。
//参数<2>的名称为“是否等待程序运行完毕”，类型为“逻辑型（bool）”，初始值为“假”。
//参数<3>的名称为“被运行程序窗口显示方式”，类型为“整数型（int）”，可以被省略。参数值可以为以下常量之一：1、#隐藏窗口； 2、#普通激活； 3、#最小化激活； 4、#最大化激活； 5、#普通不激活； 6、#最小化不激活。如果省略本参数，默认为“普通激活”方式。
//
//操作系统需求： Windows、Linux

func C程序_运行Win(欲运行的命令行 string) string {
	var err error

	//cmd := exec.Command("cmd")
	cmd := exec.Command("powershell")
	in := bytes.NewBuffer(nil)
	cmd.Stdin = in //绑定输入
	var out bytes.Buffer
	cmd.Stdout = &out //绑定输出
	go func(欲运行的命令行 string) {
		// start stop restart
		in.WriteString(欲运行的命令行) //写入你的命令，可以有多行，"\n"表示回车
	}(欲运行的命令行)
	err = cmd.Start()

	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("Command finished with error: %v", err)
	}

	rt := W文本_gbk到utf8(out.String())
	//fmt.Println(rt)

	return rt
}

func C程序_延时(毫秒数 int64) bool {
	time.Sleep(time.Duration(毫秒数) * time.Millisecond)
	return true
}
