package database

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"

	_ "github.com/lib/pq"
)

var (
	db         *sql.DB
	NotPtrList = errors.New("dest must be a pointer to a slice")
)

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

// CheckErr 检查错误
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

// ScanSource 扫描源，即sql.Rows或sql.Row
type ScanSource interface {
	Scan(dest ...any) error
	Err() error
}

// ScanModel 可分解原始数据的Model
// 用于从sql.Rows中读取数据
type ScanModel interface {
	// ScanFrom 从src中读取数据写入当前对象
	ScanFrom(src ScanSource, err error) error
}

// ForeignModel 外键Model
type ForeignModel interface {
	// ForeignValue 返回外键值
	ForeignValue() int64
}

// ForeignScanModel 外键扫描Model
type ForeignScanModel interface {
	ScanModel
	ForeignModel
}

// ScanToList 扫描结果集到列表
// dest必须是一个指向切片的指针
func ScanToList[T ScanModel](dest *[]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		*dest = append(*dest, elem)
	}
	return rs.Err()
}

// ScanToIndex 扫描结果集到一对多外键Map
func ScanToIndex[T ForeignScanModel](dest map[int64][]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem().Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		idx := elem.ForeignValue()
		dest[idx] = append(dest[idx], elem)
	}
	return rs.Err()
}

// ScanToUnique  扫描结果集到一对一外键Map
func ScanToUnique[T ForeignScanModel](dest map[int64]T, rs *sql.Rows) error {
	defer rs.Close()
	dt := reflect.TypeOf(dest).Elem()
	if dt.Kind() != reflect.Ptr {
		return NotPtrList
	}

	for rs.Next() {
		var elem = reflect.New(dt.Elem()).Interface().(T)
		if err := elem.ScanFrom(rs, nil); err != nil {
			return err
		}
		dest[elem.ForeignValue()] = elem
	}
	return rs.Err()
}
