package conf

import (
	"github.com/BurntSushi/toml"
	"github.com/TensShinet/watch_gpu/client/logging"
	"reflect"
)

type configType struct {
	addr     string
	hostname string
	interval int
	low      int
	times    int
}

// 默认配置
var sysConfig = configType{
	addr:     "127.0.0.1:8080",
	hostname: "",
	interval: 3,
	times:    60, // 失败次数
}

var logger = logging.GetLogger("conf")

func init() {
	configFilePath := "client.conf"
	if _, err := toml.DecodeFile(configFilePath, &sysConfig); err != nil {
		logger.WithError(err).WithField("path", configFilePath).Fatal("failed to load configurations")
	} else {
		level := logging.GetLevel(GetString("LogLevel"))
		logger.SetLevel(level)
		logger.WithField("path", configFilePath).Debug("Configuration file successfully loaded")
	}
}

// 根据item获取获取配置值
//
// 当获取的项目不存在时返回nil
func Get(item string) interface{} {
	return nil
}

// 根据item获取获取配置值
//
// 同Get()，但返回string类型的值
func GetString(item string) string {
	r := reflect.ValueOf(sysConfig)
	return r.FieldByName(item).String()
}

// 根据item获取获取配置值
//
// 同Get()，但返回int类型的值
func GetInt(item string) int {
	r := reflect.ValueOf(sysConfig)
	return int(r.FieldByName(item).Int())
}
