package utils

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// 目录不存在直接创建,目录已存在不操作返回nil  失败返回err
func M目录_创建(路径 string) error {
	return os.MkdirAll(路径, os.ModePerm)
}

//子程序名：目录_枚举子目录1
//取一个文件夹下级子目录；成功返回子目录数量，失败返回0；通过是否枚举子目录参数，可以枚举所有的子目录
//返回值类型：整数型
//参数<1>的名称为“父文件夹路径”，类型为“文本型”。注明：如：D:\Program Files；目录分割符请用\，路径不以\结尾会自动添加。
//参数<2>的名称为“子目录数组”，类型为“文本型”，接收参数数据时采用参考传递方式，允许接收空参数数据，需要接收数组数据。注明：用来装载返回的子目录路径；。
//参数<3>的名称为“是否带路径”，类型为“逻辑型”，允许接收空参数数据。注明：可为空默认为真,假=不带，真=带;。
//参数<4>的名称为“是否继续向下枚举”，类型为“逻辑型”，允许接收空参数数据。注明：为空，默认不枚举。

func M目录_枚举子目录(父文件夹路径 string, 子目录数组 *[]string, 是否带路径 bool, 是否继续向下枚举 bool) error {
	l, err := ioutil.ReadDir(父文件夹路径)
	if err != nil {
		return err
	}
	separator := "/"
	for _, f := range l {
		tmp := string(父文件夹路径 + separator + f.Name())

		if f.IsDir() {
			if 是否带路径 {
				*子目录数组 = append(*子目录数组, tmp)
			} else {
				*子目录数组 = append(*子目录数组, f.Name())
			}
			if 是否继续向下枚举 {
				err = M目录_枚举子目录(tmp, 子目录数组, 是否带路径, 是否继续向下枚举)
				if err != nil {
					return err
				}
			}
		}
	}
	return err
}

// 调用格式： 〈文本型〉 取运行目录 （） - 系统核心支持库->环境存取
// 英文名称：GetRunPath
// 取当前被执行的易程序文件所处的目录。本命令为初级命令。
//
// 操作系统需求： Windows
func M目录_取运行目录() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	return strings.Replace(dir, "\\", "/", -1)
}

func M目录_取当前目录() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	return dir
}

// 操作系统需求： Windows、Linux
func M目录_删除(欲删除的目录名称 string) error {
	return os.RemoveAll(欲删除的目录名称)
}
