package services

import "github.com/azhai/goent/utils"

// Config holds application configuration loaded from .env
type Config struct {
	AppName  string
	Host     string
	Port     int
	CertDir  string
	ImageDir string
	ThumbDir string
	LogFile  string
	DBType   string
	DBDSN    string
}

// LoadConfig loads configuration from .env file with defaults.
// If .env file doesn't exist, falls back to system environment variables.
func LoadConfig() *Config {
	env, _ := utils.LoadEnvFile(".env")
	return &Config{
		AppName:  env.GetStr("APP_NAME", "Bing Wallpaper"),
		Host:     env.GetStr("HTTP_HOST", "127.0.0.1"),
		Port:     env.GetInt("HTTP_PORT", 8080),
		CertDir:  env.GetStr("CERT_DIR", "./certs"),
		ImageDir: env.GetStr("IMAGE_DIR", "./images"),
		ThumbDir: env.GetStr("THUMB_DIR", "./thumbs"),
		LogFile:  env.GetStr("LOG_FILE", "./logs/access.log"),
		DBType:   env.GetStr("DB_TYPE", "sqlite"),
		DBDSN:    env.GetStr("DB_DSN", "bingwp.db"),
	}
}
