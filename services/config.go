package services

import (
	"os"
	"path/filepath"

	"github.com/azhai/goent/drivers"
	"github.com/azhai/goent/utils"
)

// Config holds application configuration loaded from .env
type Config struct {
	AppName  string
	Host     string
	Port     int
	CertDir  string
	ImageDir string
	ThumbDir string
	LogFile  string
	Database drivers.DatabaseConfig
}

func (c *Config) GetCertFile() (string, string, bool) {
	var err error
	certPath := filepath.Join(c.CertDir, "cert.pem")
	keyPath := filepath.Join(c.CertDir, "key.pem")
	if _, err = os.Stat(certPath); err == nil {
		_, err = os.Stat(keyPath)
	}
	return certPath, keyPath, err == nil
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
		Database: drivers.LoadConfig(env, "bingwp.db"),
	}
}
