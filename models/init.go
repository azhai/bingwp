package models

import (
	"net/url"

	"github.com/azhai/bingwp/cmd"
	"github.com/azhai/xgen/dialect"
)

var (
	connCfgs = make(map[string]dialect.ConnConfig)
	connKeys = url.Values{}
)

func init() {
	if cmd.IsRunTest() {
		_, _ = cmd.BackToDir(1) // 从tests退回根目录
		SetupDb()
	}
}

func SetupDb() {
	settings := cmd.GetTheSettings()
	if settings == nil {
		return
	}
	for _, c := range settings.Conns {
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
