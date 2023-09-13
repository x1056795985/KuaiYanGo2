package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// W文件_是否存在 判断一个文件或文件夹是否存在
// 输入文件路径，根据返回的bool值来判断文件或文件夹是否存在
func W文件_是否存在(路径 string) bool {
	_, err := os.Stat(路径)
	if err == nil {
		return true
	}

	return false
}

// W文件_写到文件
func W_创建目录(路径 string) {
	err := os.MkdirAll(路径, os.ModePerm)
	if err != nil {
		return
	}
	file, _ := os.OpenFile(路径, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Errorf("W_写到文件文件失败: %s \n", err)
		}
	}(file)

	_, err = file.WriteString("")
	if err != nil {
		fmt.Errorf("W_写到文件文件失败2: %s \n", err)
		return
	}
}

// 取运行目录
func C程序_取运行目录() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res

}
