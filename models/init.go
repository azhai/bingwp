package models

import (
	"net/url"

	"github.com/azhai/gozzo/config"
	"github.com/azhai/xgen/dialect"
)

var (
	connCfgs = make(map[string]dialect.ConnConfig)
	connKeys = url.Values{}
)

func init() {
	if config.IsRunTest() {
		_, _ = config.BackToDir(1) // 从tests退回根目录
		SetupDb()
	}
}

func SetupDb() {
	for _, c := range config.GetConnConfigs() {
		connCfgs[c.Key] = c
		connKeys.Add(c.Type, c.Key)
	}
}

func GetConnTypes() []string {
	var result []string
	for ct := range connKeys {
		result = append(result, ct)
	}
	return result
}

func GetConnKeys(connType string) []string {
	if keys, ok := connKeys[connType]; ok {
		return keys
	}
	return nil
}

func GetConnConfig(key string) dialect.ConnConfig {
	if cfg, ok := connCfgs[key]; ok {
		return cfg
	}
	return dialect.ConnConfig{}
}
