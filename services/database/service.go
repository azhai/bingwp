package database

import (
	"context"
	"database/sql"
	"os"

	_ "github.com/lib/pq"
)

var db *sql.DB

// GetDB 获取数据库服务
func GetDB() *sql.DB {
	if db == nil {
		err := OpenService()
		CheckErr(err)
	}
	return db
}

// OpenService 初始化服务
func OpenService() (err error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dba:pass@127.0.0.1/db_bingwp?sslmode=disable"
	}
	db, err = sql.Open("postgres", dsn)
	CheckErr(err)
	ctx := context.Background()
	err = db.PingContext(ctx)
	CheckErr(err)
	return
}

// CloseService 关闭服务
func CloseService() {
	if db != nil {
		_ = db.Close()
	}
}

func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}
