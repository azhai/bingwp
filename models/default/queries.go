package db

import (
	xq "github.com/azhai/xgen/xquery"
	"xorm.io/xorm"
)

// Load the queries of WallDaily
func (m *WallDaily) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return Query(opts...).Get(m)
}

// Save the queries of WallDaily
func (m *WallDaily) Save(changes map[string]any) error {
	return xq.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		if len(changes) == 0 {
			return tx.Table(m).Insert(m)
		} else if m.Id == 0 {
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.Id).Update(changes)
		}
	})
}

// Load the queries of WallImage
func (m *WallImage) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return Query(opts...).Get(m)
}

// Save the queries of WallImage
func (m *WallImage) Save(changes map[string]any) error {
	return xq.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		if len(changes) == 0 {
			return tx.Table(m).Insert(m)
		} else if m.Id == 0 {
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.Id).Update(changes)
		}
	})
}

// Load the queries of WallNote
func (m *WallNote) Load(opts ...xq.QueryOption) (bool, error) {
	opts = append(opts, xq.WithTable(m))
	return Query(opts...).Get(m)
}

// Save the queries of WallNote
func (m *WallNote) Save(changes map[string]any) error {
	return xq.ExecTx(Engine(), func(tx *xorm.Session) (int64, error) {
		if len(changes) == 0 {
			return tx.Table(m).Insert(m)
		} else if m.Id == 0 {
			return tx.Table(m).Insert(changes)
		} else {
			return tx.Table(m).ID(m.Id).Update(changes)
		}
	})
}
