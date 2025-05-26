package utils

import (
	"EFunc/utils"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

func X系统_权限检测() bool {
	//检查自身进程是否有写入数据的权限
	// 这个例子测试写权限，如果没有写权限则返回error。
	// 注意文件不存在也会返回error，需要检查error的信息来获取到底是哪个错误导致。

	utils.W文件_删除(GetCurrentAbPathByExecutable() + "/权限测试.json")

	f, err := os.Create(GetCurrentAbPathByExecutable() + "/权限测试.json")
	if err != nil {
		return false
	}
	defer f.Close()
	f.WriteString(time.Now().String())
	// 获取当前权限
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	fmt.Printf("文件权限 %v\n", fi.Mode())
	return true
}

// 获取当前执行程序所在的绝对路径
func GetCurrentAbPathByExecutable() string {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	res, _ := filepath.EvalSymlinks(filepath.Dir(exePath))
	return res
}

func X系统_取总内存G() float64 {
	if runtime.GOOS == "windows" {
		// Windows系统
		//var memStatus syscall.MemoryStatusEx
		//memStatus.Length = uint32(unsafe.Sizeof(memStatus))
		//ret, _, _ := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(
		//	uintptr(unsafe.Pointer(&memStatus)))
		//if ret == 0 {
		//	return 0
		//}
		//return memStatus.TotalPhys
		return 0
	} else {
		// Linux/Unix系统
		var info syscall.Sysinfo_t
		err := syscall.Sysinfo(&info)
		if err != nil {
			return 0
		}

		return utils.Float64除int64(float64(info.Totalram*uint64(info.Unit))/1024/1024, 1024, 2)
	}
}
