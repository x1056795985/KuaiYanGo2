package utils

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

func Reload() error {
	if runtime.GOOS == "windows" {
		return errors.New("系统不支持")
	}
	pid := os.Getpid()
	cmd := exec.Command("kill", "-1", strconv.Itoa(pid))
	return cmd.Run()
}

//func Reload热重启(继承端口 int) error {
//	if runtime.GOOS == "windows" {
//		return errors.New("不支持windows")
//	}
//
//	新服务器 := endless.NewServer(":"+strconv.Itoa(继承端口), gin.Default())
//	return 新服务器.ListenAndServe()
//
//}
