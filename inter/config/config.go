package config

import (
	"DictGenerate/inter/logger"
	"DictGenerate/util"
	"github.com/json-iterator/go"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

var (
	ConfigFilePath  = filepath.Join(GetConfigDir(), ConfigName)
	HistoryFilePath = filepath.Join(os.TempDir(), HistoryFileName)

	// Config 配置信息, 由外部调用
	C = NewConfig(ConfigFilePath)
)

type Config struct {
	Storage *configJSONExport

	configFilePath string
	configFile     *os.File
	fileMu         sync.Mutex
}

func NewConfig(configFilePath string) *Config {
	return &Config{
		configFilePath: configFilePath,
	}
}

// 初始化配置
func (c *Config) Init() error {
	return c.init()
}

// Reload 从文件重载配置
func (c *Config) Reload() error {
	return c.init()
}

// Reset 重置默认配置
func (c *Config) Reset() error {
	c.Storage = nil
	c.initDefaultConfig()
	return c.Save()
}

func (c *Config) init() error {
	if c.configFilePath == "" {
		return ErrConfigFileNotExist
	}

	// 初始化默认配置
	c.initDefaultConfig()

	// 载入配置
	err := c.loadConfigFromFile()
	if err != nil {
		return err
	}

	return nil
}

// 载入配置
func (c *Config) loadConfigFromFile() error {

	// 打开配置文件
	if err := c.lazyOpenConfigFile(); err != nil {
		return err
	}

	// 未初始化
	info, err := c.configFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() == 0 {
		err = c.Save()
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	_, err = c.configFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	d := jsoniter.NewDecoder(c.configFile)
	err = d.Decode(c.Storage)
	if err != nil {
		return ErrConfigContentsParseError
	}

	return nil
}

// 打开配置文件
func (c *Config) lazyOpenConfigFile() (err error) {
	if c.configFile != nil {
		return nil
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(c.configFilePath), 0700); err != nil {
		return err
	}
	c.configFile, err = os.OpenFile(c.configFilePath, os.O_CREATE|os.O_RDWR, 0600)

	if err != nil {
		if os.IsPermission(err) {
			return ErrConfigFileNoPermission
		}
		if os.IsExist(err) {
			return ErrConfigFileNotExist
		}
		return err
	}
	return nil
}

// 初始化默认配置
func (c *Config) initDefaultConfig() {
	if c.Storage != nil {
		return
	}

	c.Storage = NewConfigJSONExport()
}

// Save 保存配置信息到配置文件
func (c *Config) Save() error {

	// 打开配置文件
	if err := c.lazyOpenConfigFile(); err != nil {
		return err
	}

	c.fileMu.Lock()
	defer c.fileMu.Unlock()

	data, err := jsoniter.MarshalIndent(c.Storage, "", " ")
	if err != nil {
		// json数据生成失败
		panic(err)
	}

	// 减掉多余的部分
	err = c.configFile.Truncate(int64(len(data)))
	if err != nil {
		return err
	}

	_, err = c.configFile.Seek(0, io.SeekStart)
	if err != nil {
		return err
	}

	_, err = c.configFile.Write(data)
	if err != nil {
		return err
	}

	return nil
}

// Close 关闭配置文件
func (c *Config) Close() error {
	if c.configFile != nil {
		err := c.configFile.Close()
		c.configFile = nil
		return err
	}
	return nil
}

// GetConfigDir 获取配置路径
func GetConfigDir() string {
	// 如果旧版的配置文件存在, 则使用旧版
	oldConfigDir := util.ExecutablePath()
	_, err := os.Stat(filepath.Join(oldConfigDir, ConfigName))
	if err == nil {
		return oldConfigDir
	}

	switch runtime.GOOS {
	case "windows":
		return oldConfigDir
	default:
		dataPath, ok := os.LookupEnv("HOME")
		if !ok {
			logger.Warn("Environment HOME not set")
			return oldConfigDir
		}
		configDir := filepath.Join(dataPath, ".config", AppName)

		// 检测是否可写
		err = os.MkdirAll(configDir, 0700)
		if err != nil {
			logger.Warnf("check config dir error: %s", err)
			return oldConfigDir
		}
		return configDir
	}
}
