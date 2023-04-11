package logger

/*文件处理模块
针对不同的文件异常处理，最终功能为递归打开文件
文件所有者可读可写，组内其他用户可读，其他用户可读
*/

import (
	"fmt"
	"os"
)

func checkNotExist(src string) bool {
	_, err := os.Stat(src)
	return os.IsNotExist(err)
}

func checkPermission(src string) bool {
	_, err := os.Stat(src)
	return os.IsPermission(err)
}

func mkDir(src string) error {
	return os.MkdirAll(src, os.ModePerm)
}

func isNotExistMkDir(src string) error {
	if checkNotExist(src) {
		return mkDir(src)
	}
	return nil
}

func mustOpen(fileName, dir string) (*os.File, error) {
	if checkPermission(dir) {
		return nil, fmt.Errorf("permission denied dir: %s", dir)
	}

	if err := isNotExistMkDir(dir); err != nil {
		return nil, fmt.Errorf("error during make dir %s, err: %s", dir, err)
	}

	f, err := os.OpenFile(dir+string(os.PathSeparator)+fileName, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("fail to open file, err : %s", err)
	}

	return f, nil
}
