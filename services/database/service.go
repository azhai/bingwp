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
	dbServ     *DBServ
	NotPtrList = errors.New("dest must be a pointer to a slice")
)

type NullString = sql.NullString
type NullInt64 = sql.NullInt64
type NullFloat64 = sql.NullFloat64
type NullBool = sql.NullBool
type NullTime = sql.NullTime

type DBServ struct {
	*sql.DB
}

// New 获取数据库服务
func New() *DBServ {
	if dbServ == nil {
		db, err := OpenService()
		CheckErr(err)
		dbServ = &DBServ{db}
	}
	return dbServ
}

// OpenService 初始化服务
func OpenService() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dba:pass@127.0.0.1/db_bingwp?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err == nil {
		ctx := context.Background()
		err = db.PingContext(ctx)
	}
	return db, err
}

// CloseService 关闭服务
func CloseService() {
	if dbServ != nil {
		_ = dbServ.Close()
		dbServ = nil
	}
}

// CheckErr 检查错误
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

type Model interface {
	// TableName 返回表名
	TableName() string
}

type ModelComment interface {
	Model
	// TableComment 返回表备注
	TableComment() string
}

type ModelInsert interface {
	Model
	// InsertSQL 插入一行的SQL语句
	InsertSQL() string
	// RowValues 插入一行所需数据
	RowValues() []any
	// SetId 设置主键值
	SetId(id int64, err error) error
}

type ModelUpdate interface {
	Model
	// UpdateSQL 更新一行的SQL语句
	UpdateSQL() string
}

func Insert(row ModelInsert) (bool, error) {
	res, err := New().Exec(row.InsertSQL(), row.RowValues()...)
	var ok bool
	if err == nil {
		ok, err = true, row.SetId(res.LastInsertId())
	}
	return ok, err
}

func InsertBatch[T ModelInsert](rows []T) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	stmt, err := New().Prepare(rows[0].InsertSQL())
	if err != nil {
		return 0, err
	}
	var (
		num int
		res sql.Result
	)
	for _, row := range rows {
		if row == nil {
			continue
		}
		if res, err = stmt.Exec(row.RowValues()...); err != nil {
			break
		}
		if err = row.SetId(res.LastInsertId()); err == nil {
			num++
		}
	}
	err = stmt.Close()
	return num, err
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
	ForeignValue() any
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
func ScanToIndex[K comparable, T ForeignScanModel](dest map[K][]T, rs *sql.Rows) error {
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
		idx := elem.ForeignValue().(K)
		dest[idx] = append(dest[idx], elem)
	}
	return rs.Err()
}

// ScanToUnique  扫描结果集到一对一外键Map
func ScanToUnique[K comparable, T ForeignScanModel](dest map[K]T, rs *sql.Rows) error {
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
		idx := elem.ForeignValue().(K)
		dest[idx] = elem
	}
	return rs.Err()
}
