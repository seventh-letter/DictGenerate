package config

import (
	"github.com/ziiber/go-logging"
	"os"
)

func init() {
	// 日志格式配置
	var format = logging.MustStringFormatter(
		`%{message}`,
	)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2.Color = true
	backendFormat := logging.NewBackendFormatter(backend2, format)
	backendStd := logging.AddModuleLevel(backendFormat)
	backendStd.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backendStd)
}