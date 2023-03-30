package utils

import (
	"os"
)

// PathExists 判断文件是否存在
func PathExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

//FileExists 判断文件夹是否存在
func FileExists(path string) (err error) {
	exists := PathExists(path)
	if exists == false {
		err = os.MkdirAll(path, os.ModePerm)
	}
	return
}
