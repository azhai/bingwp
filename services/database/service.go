package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"slices"
	"strconv"
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

// CheckErr 检查错误
func CheckErr(err error) {
	if err != nil {
		panic(err)
	}
}

// IsRunTest 是否测试模式下
func IsRunTest() bool {
	return strings.HasSuffix(os.Args[0], ".test")
}

// QuestionMarkHolders 将SQL中的顺序占位符替换为问号占位符
func QuestionMarkHolders(query string) string {
	re := regexp.MustCompile(`\$\d+`)
	return re.ReplaceAllString(query, "?")
}

// FlattenPlaceHolders 将SQL中对应slice参数的顺序占位符进行扩充
func FlattenPlaceHolders(query string, args []any) (string, []any) {
	var (
		pairs  []string
		values []any
	)
	for i, arg := range args {
		old := fmt.Sprintf("$%d", i+1)
		rv := reflect.ValueOf(arg)
		if rv.Kind() != reflect.Slice {
			values = append(values, arg)
			holder := "$" + strconv.Itoa(len(values))
			if old != holder {
				pairs = append(pairs, holder, old)
			}
			continue
		}
		var holderList []string
		for j := 0; j < rv.Len(); j++ {
			values = append(values, rv.Index(j).Interface())
			holder := "$" + strconv.Itoa(len(values))
			holderList = append(holderList, holder)
		}
		pairs = append(pairs, strings.Join(holderList, ", "), old)
	}
	slices.Reverse(pairs)
	replacer := strings.NewReplacer(pairs...)
	return replacer.Replace(query), values
}

// RegularNamedPlaceHolders 将SQL中的命名占位符替换为pq driver库的顺序占位符
func RegularNamedPlaceHolders(query string, nargs []sql.NamedArg) (string, []any) {
	var (
		pairs  []string
		values []any
	)
	for _, arg := range nargs {
		values = append(values, arg.Value)
		holder := "$" + strconv.Itoa(len(values))
		pairs = append(pairs, "@"+arg.Name, holder)
	}
	// 注意短匹配优先 overlapping matches
	replacer := strings.NewReplacer(pairs...)
	return replacer.Replace(query), values
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
	dsn string
	*sql.DB
}

// New 获取数据库服务
func New() *DBServ {
	if dbServ == nil {
		_, err := OpenService()
		CheckErr(err)
	}
	return dbServ
}

func (s *DBServ) WithLogger(filename string) {
	logger := logging.NewLoggerURL("info", filename)
	loggerAdapter := zapadapter.New(logger.Desugar())
	s.DB = sqldblogger.OpenDriver(s.dsn, s.DB.Driver(), loggerAdapter)
}

func (s *DBServ) FlattenExec(ctx context.Context,
	query string, args ...any) (sql.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	query, args = FlattenPlaceHolders(query, args)
	return s.ExecContext(ctx, query, args...)
}

func (s *DBServ) FlattenQuery(ctx context.Context,
	query string, args ...any) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	query, args = FlattenPlaceHolders(query, args)
	return s.QueryContext(ctx, query, args...)
}

func (s *DBServ) NamedExec(ctx context.Context,
	query string, nargs []sql.NamedArg) (sql.Result, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var args []any
	query, args = RegularNamedPlaceHolders(query, nargs)
	return s.ExecContext(ctx, query, args...)
}

func (s *DBServ) NamedQuery(ctx context.Context,
	query string, nargs []sql.NamedArg) (*sql.Rows, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	var args []any
	query, args = RegularNamedPlaceHolders(query, nargs)
	return s.QueryContext(ctx, query, args...)
}

// OpenService 初始化服务
func OpenService() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://dba:pass@127.0.0.1/db_bingwp?sslmode=disable"
	}
	db, err := sql.Open("postgres", dsn)
	if err == nil && db != nil {
		dbServ = &DBServ{DB: db, dsn: dsn}
		if IsRunTest() {
			logFile = testLogFile
		}
		dbServ.WithLogger(logFile)
		ctx := context.Background()
		err = dbServ.PingContext(ctx)
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
	var nargs []sql.NamedArg
	query := "UPDATE " + table + " SET "
	for k, val := range changes {
		kk := fmt.Sprintf("_S_%s_", k)
		query += fmt.Sprintf("%s = @%s, ", k, kk)
		nargs = append(nargs, sql.Named(kk, val))
	}
	query = strings.TrimSuffix(query, ", ") + " WHERE " + where
	if len(wargs) > 0 {
		nargs = append(nargs, wargs...)
	}
	res, err := New().NamedExec(nil, query, nargs)
	var num int64
	if err == nil {
		num, err = res.RowsAffected()
	}
	return int(num), err
}

func UpdateRow(row ModelChanger, changes map[string]any) (int, error) {
	table, where := row.TableName(), ""
	cols, values := row.UniqFields()
	// WHERE参数可能和SET参数有相同字段
	var wargs []sql.NamedArg
	for i, k := range cols {
		kk, val := fmt.Sprintf("_W_%s_", k), values[i]
		where += fmt.Sprintf("%s = @%s AND ", k, kk)
		wargs = append(wargs, sql.Named(kk, val))
	}
	where = strings.TrimSuffix(where, " AND ")
	return ExecUpdate(table, where, wargs, changes)
}

func execInsert(row ModelChanger, stmt *sql.Stmt, query string) (bool, error) {
	var (
		err error
		res sql.Result
	)
	args := row.RowValues()
	if stmt != nil {
		res, err = stmt.Exec(args...)
		// if res != nil {
		// 	err = row.SetId(res.LastInsertId())
		// }
	} else if len(query) > 0 {
		res, err = New().Exec(query, args...)
	}
	if err != nil || res == nil {
		return false, err
	}
	return true, err
}

func execInsertId(row ModelChanger, stmt *sql.Stmt, query string) (int64, error) {
	var (
		err    error
		lastId int64
	)
	args := row.RowValues()
	if stmt != nil {
		err = stmt.QueryRow(args...).Scan(&lastId)
	} else if len(query) > 0 {
		err = New().QueryRow(query, args...).Scan(&lastId)
	}
	err = row.SetId(lastId, err)
	return lastId, err
}

func UpsertRow(row ModelChanger) (bool, error) {
	return execInsert(row, nil, row.UpsertSQL())
}

func InsertRow(row ModelChanger) (bool, error) {
	query := row.InsertSQL()
	if !strings.Contains(query, " RETURNING ") {
		return execInsert(row, nil, query)
	}
	_, err := execInsertId(row, nil, query)
	return err == nil, err
}

func InsertBatch[T ModelChanger](rows []T) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	query := rows[0].InsertSQL()
	stmt, err := New().Prepare(query)
	if err != nil {
		return 0, err
	}
	num, withId := 0, strings.Contains(query, " RETURNING ")
	for _, row := range rows {
		if !withId {
			_, err = execInsert(row, stmt, query)
		} else {
			_, err = execInsertId(row, stmt, query)
		}
		if err != nil {
			break
		}
		num++
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
