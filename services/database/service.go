package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/azhai/gozzo/logging"
	_ "github.com/lib/pq"
	"github.com/simukti/sqldb-logger"
	"github.com/simukti/sqldb-logger/logadapter/zapadapter"
)

var (
	dbServ      *DBServ
	NotPtrList  = errors.New("dest must be a pointer to a slice")
	logFile     = "rotate://./logs/sql.log?cycle=daily&comp=0"
	testLogFile = "rotate://./sql.log?cycle=daily&comp=0"
)

// IsRunTest 是否测试模式下
func IsRunTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

type NullString = sql.NullString
type NullInt64 = sql.NullInt64
type NullFloat64 = sql.NullFloat64
type NullBool = sql.NullBool
type NullTime = sql.NullTime

func NewNullString(v string) NullString {
	return sql.NullString{String: v, Valid: true}
}

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
	if err == nil && db != nil {
		if IsRunTest() {
			logFile = testLogFile
		}
		logger := logging.NewLoggerURL("info", logFile)
		loggerAdapter := zapadapter.New(logger.Desugar())
		db = sqldblogger.OpenDriver(dsn, db.Driver(), loggerAdapter)
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

type ModelChanger interface {
	Model
	// UniqFields 可作为更新条件的字段与它的值
	UniqFields() ([]string, []any)
	// SetId 设置主键值
	SetId(id int64, err error) error
	// RowValues 插入一行所需数据
	RowValues() []any
	// InsertSQL 插入一行的SQL语句
	InsertSQL() string
	// UpsertSQL 插入或更新一行的SQL语句
	UpsertSQL() string
}

func ExecUpdate(table, where string, wargs []sql.NamedArg,
	changes map[string]any) (int, error) {
	var args []sql.NamedArg
	query := "UPDATE " + table + " SET "
	for k, v := range changes {
		query += fmt.Sprintf("%s=@%s, ", k, k)
		args = append(args, sql.Named(k, v))
	}
	query = strings.TrimSuffix(query, ", ") + " WHERE " + where
	if len(wargs) > 0 {
		args = append(args, wargs...)
	}
	res, err := New().Exec(query, args)
	var num int64
	if err == nil {
		num, err = res.RowsAffected()
	}
	return int(num), err
}

func UpdateRow(row ModelChanger, changes map[string]any) (int, error) {
	table, where := row.TableName(), ""
	cols, values := row.UniqFields()
	var wargs []sql.NamedArg
	for i, k := range cols {
		where += fmt.Sprintf("%s=@_%s_ AND ", k, k)
		kk, val := fmt.Sprintf("_%s_", k), values[i]
		wargs = append(wargs, sql.Named(kk, val))
	}
	where = strings.TrimSuffix(where, " AND ")
	return ExecUpdate(table, where, wargs, changes)
}

func execInsert(row ModelChanger, query string) (bool, error) {
	res, err := New().Exec(query, row.RowValues()...)
	var ok bool
	if err == nil {
		ok, err = true, row.SetId(res.LastInsertId())
	}
	return ok, err
}

func UpsertRow(row ModelChanger) (bool, error) {
	return execInsert(row, row.UpsertSQL())
}

func InsertRow(row ModelChanger) (bool, error) {
	return execInsert(row, row.InsertSQL())
}

func InsertBatch[T ModelChanger](rows []T) (int, error) {
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

// ModelLoader 可分解原始数据的Model
// 用于从sql.Rows中读取数据
type ModelLoader interface {
	// ScanFrom 从src中读取数据写入当前对象
	ScanFrom(src ScanSource, err error) error
}

// ModelForeignLoader 外键扫描Model
type ModelForeignLoader interface {
	ModelLoader
	// ForeignIndex 返回外键的值
	ForeignIndex() any
}

// ModelSecondaryLoader 外键扫描Model
type ModelSecondaryLoader interface {
	ModelForeignLoader
	// SecondaryKey 返回次要字段的值
	SecondaryKey() string
}

// ScanToList 扫描结果集到列表
// dest必须是一个指向切片的指针
func ScanToList[T ModelLoader](dest *[]T, rs *sql.Rows) error {
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

// ScanToUnique  扫描结果集到一对一外键Map
func ScanToUnique[K comparable, T ModelForeignLoader](dest map[K]T, rs *sql.Rows) error {
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
		idx := elem.ForeignIndex().(K)
		dest[idx] = elem
	}
	return rs.Err()
}

// ScanToIndex 扫描结果集到一对多外键Map
func ScanToIndex[K comparable, T ModelForeignLoader](dest map[K][]T, rs *sql.Rows) error {
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
		idx := elem.ForeignIndex().(K)
		dest[idx] = append(dest[idx], elem)
	}
	return rs.Err()
}

// ScanToSecondary 扫描结果集到双层外键Map
func ScanToSecondary[K comparable, T ModelSecondaryLoader](dest map[K]map[string]T, rs *sql.Rows) error {
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
		idx := elem.ForeignIndex().(K)
		if _, ok := dest[idx]; !ok {
			dest[idx] = make(map[string]T)
		}
		key := elem.SecondaryKey()
		dest[idx][key] = elem
	}
	return rs.Err()
}
