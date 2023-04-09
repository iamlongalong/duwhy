package sqlprovider

import (
	"context"
	"duwhy/core"

	"gorm.io/gorm"
)

// Not Implement
type DuSqlProvider struct {
}

func (sp *DuSqlProvider) GetInfoByPath(pathname string, opts *core.InfoOption) (*core.InfoItem, error) {
	return nil, nil
}

type ISqlGetter interface {
	GetInfoItemByPath(ctx context.Context, paths []string, opts *core.InfoOption) ([]*core.InfoItem, error)
}

func NewSqlGetter(db *gorm.DB, tableName string) (ISqlGetter, error) {
	return &SqlGetter{
		db:        db,
		tableName: tableName,
	}, nil
}

type SqlGetter struct {
	db *gorm.DB

	tableName string
}

// Not Implement
func (sg *SqlGetter) GetInfoItemByPath(ctx context.Context, paths []string, opts *core.InfoOption) ([]*core.InfoItem, error) {
	items := []*core.InfoItem{}
	err := sg.db.Table(sg.tableName).WithContext(ctx).Where("").Find(&items).Error
	if err != nil {
		return nil, err
	}

	return items, nil
}
