package util

import (
	"bufio"
	"fmt"
	"github.com/kardianos/osext"
	"os"
	"path/filepath"
	"strings"
)

// Executable 获取程序所在的真实目录或真实相对路径
func Executable() string {
	executablePath, err := osext.Executable()
	if err != nil {
		executablePath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			executablePath = filepath.Dir(os.Args[0])
		}
	}

	// 读取链接
	linkedExecutablePath, err := filepath.EvalSymlinks(executablePath)
	if err != nil {
		return executablePath
	}
	return linkedExecutablePath
}

// ExecutablePath 获取程序所在目录
func ExecutablePath() string {
	return filepath.Dir(Executable())
}

// ExecutablePathJoin 返回程序所在目录的子目录
func ExecutablePathJoin(subPath string) string {
	return filepath.Join(ExecutablePath(), subPath)
}

// ContainsString 检测字符串是否在字符串数组里
func ContainsString(ss []string, s string) bool {
	for k := range ss {
		if ss[k] == s {
			return true
		}
	}
	return false
}

// ChWorkDir 切换回工作目录
func ChWorkDir() {
	dir, err := filepath.Abs("")
	if err != nil {
		return
	}

	subPath := filepath.Dir(os.Args[0])
	os.Chdir(strings.TrimSuffix(dir, subPath))
}

// 输出到文件
func OutputFile(name string, list []string) error {

	var (
		fp  *os.File
		err error
	)
	if FileExists(name) {
		fp, err = os.OpenFile(name, os.O_WRONLY|os.O_TRUNC, 0644)
	} else {
		fp, err = os.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	}

	if err != nil {
		return err
	}

	defer fp.Close()

	w := bufio.NewWriter(fp)
	for i := 0; i < len(list); i++ {
		fmt.Fprintln(w, list[i])
	}

	return w.Flush()
}

// 文件是否存在
func FileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}
