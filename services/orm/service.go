package orm

import (
	"os"
	"time"

	"github.com/azhai/bingwp/utils"
	"github.com/go-goe/goe"
	"github.com/go-goe/postgres"
)

// DatabaseService 数据库服
// table tag不起作用
type DatabaseService struct {
	Daily *WallDaily `goe:"table:t_wall_daily"`
	Image *WallImage `goe:"table:t_wall_image"`
	Note  *WallNote  `goe:"table:t_wall_note"`
	*goe.DB
}

var svc *DatabaseService

// Serv 获取数据库服务
func Serv() *DatabaseService {
	return svc
}

// OpenService 初始化服务
func OpenService() (err error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=127.0.0.1 user=dba password=pass dbname=db_bingwp sslmode=disable"
	}
	cfg := postgres.Config{
		DatabaseConfig: goe.DatabaseConfig{
			Logger:           utils.NewLogger("stdout"),
			IncludeArguments: true,
			QueryThreshold:   time.Second,
		},
	}
	svc, err = goe.Open[DatabaseService](postgres.Open(dsn, cfg))
	if err == nil {
		// migrate all orm structs
		err = goe.AutoMigrate(svc)
	}
	return
}

// CloseService 关闭服务
func CloseService() {
	if svc != nil {
		_ = goe.Close(svc)
	}
}
