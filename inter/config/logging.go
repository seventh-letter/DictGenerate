package config

import (
	"github.com/telanflow/go-logging"
	"os"
)

func init() {
	// 日志格式配置
	var format = logging.MustStringFormatter(
		`%{color}%{message}%{color:reset}`,
	)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2.Color = true
	backendFormat := logging.NewBackendFormatter(backend2, format)
	backendStd := logging.AddModuleLevel(backendFormat)
	backendStd.SetLevel(logging.DEBUG, "")
	logging.SetBackend(backendStd)
}
