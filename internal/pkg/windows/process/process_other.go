//go:build !windows

package process

import "errors"

// 本机不支持 Windows 进程发现，但仍可运行 mock 和离线查询。
func GetProcessCommand(string) (string, error) {
	return "", errors.New("本平台不支持 League 客户端进程发现")
}
