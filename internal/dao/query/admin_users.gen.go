// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"crm_lite/internal/dao/model"
)

func newAdminUser(db *gorm.DB, opts ...gen.DOOption) adminUser {
	_adminUser := adminUser{}

	_adminUser.adminUserDo.UseDB(db, opts...)
	_adminUser.adminUserDo.UseModel(&model.AdminUser{})

	tableName := _adminUser.adminUserDo.TableName()
	_adminUser.ALL = field.NewAsterisk(tableName)
	_adminUser.ID = field.NewInt64(tableName, "id")
	_adminUser.UUID = field.NewString(tableName, "uuid")
	_adminUser.Username = field.NewString(tableName, "username")
	_adminUser.Email = field.NewString(tableName, "email")
	_adminUser.PasswordHash = field.NewString(tableName, "password_hash")
	_adminUser.RealName = field.NewString(tableName, "real_name")
	_adminUser.Phone = field.NewString(tableName, "phone")
	_adminUser.Avatar = field.NewString(tableName, "avatar")
	_adminUser.IsActive = field.NewBool(tableName, "is_active")
	_adminUser.LastLoginAt = field.NewTime(tableName, "last_login_at")
	_adminUser.CreatedAt = field.NewTime(tableName, "created_at")
	_adminUser.UpdatedAt = field.NewTime(tableName, "updated_at")
	_adminUser.DeletedAt = field.NewField(tableName, "deleted_at")

	_adminUser.fillFieldMap()

	return _adminUser
}

type adminUser struct {
	adminUserDo

	ALL          field.Asterisk
	ID           field.Int64
	UUID         field.String
	Username     field.String
	Email        field.String
	PasswordHash field.String
	RealName     field.String
	Phone        field.String
	Avatar       field.String
	IsActive     field.Bool
	LastLoginAt  field.Time
	CreatedAt    field.Time
	UpdatedAt    field.Time
	DeletedAt    field.Field

	fieldMap map[string]field.Expr
}

func (a adminUser) Table(newTableName string) *adminUser {
	a.adminUserDo.UseTable(newTableName)
	return a.updateTableName(newTableName)
}

func (a adminUser) As(alias string) *adminUser {
	a.adminUserDo.DO = *(a.adminUserDo.As(alias).(*gen.DO))
	return a.updateTableName(alias)
}

func (a *adminUser) updateTableName(table string) *adminUser {
	a.ALL = field.NewAsterisk(table)
	a.ID = field.NewInt64(table, "id")
	a.UUID = field.NewString(table, "uuid")
	a.Username = field.NewString(table, "username")
	a.Email = field.NewString(table, "email")
	a.PasswordHash = field.NewString(table, "password_hash")
	a.RealName = field.NewString(table, "real_name")
	a.Phone = field.NewString(table, "phone")
	a.Avatar = field.NewString(table, "avatar")
	a.IsActive = field.NewBool(table, "is_active")
	a.LastLoginAt = field.NewTime(table, "last_login_at")
	a.CreatedAt = field.NewTime(table, "created_at")
	a.UpdatedAt = field.NewTime(table, "updated_at")
	a.DeletedAt = field.NewField(table, "deleted_at")

	a.fillFieldMap()

	return a
}

func (a *adminUser) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := a.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (a *adminUser) fillFieldMap() {
	a.fieldMap = make(map[string]field.Expr, 13)
	a.fieldMap["id"] = a.ID
	a.fieldMap["uuid"] = a.UUID
	a.fieldMap["username"] = a.Username
	a.fieldMap["email"] = a.Email
	a.fieldMap["password_hash"] = a.PasswordHash
	a.fieldMap["real_name"] = a.RealName
	a.fieldMap["phone"] = a.Phone
	a.fieldMap["avatar"] = a.Avatar
	a.fieldMap["is_active"] = a.IsActive
	a.fieldMap["last_login_at"] = a.LastLoginAt
	a.fieldMap["created_at"] = a.CreatedAt
	a.fieldMap["updated_at"] = a.UpdatedAt
	a.fieldMap["deleted_at"] = a.DeletedAt
}

func (a adminUser) clone(db *gorm.DB) adminUser {
	a.adminUserDo.ReplaceConnPool(db.Statement.ConnPool)
	return a
}

func (a adminUser) replaceDB(db *gorm.DB) adminUser {
	a.adminUserDo.ReplaceDB(db)
	return a
}

type adminUserDo struct{ gen.DO }

type IAdminUserDo interface {
	gen.SubQuery
	Debug() IAdminUserDo
	WithContext(ctx context.Context) IAdminUserDo
	WithResult(fc func(tx gen.Dao)) gen.ResultInfo
	ReplaceDB(db *gorm.DB)
	ReadDB() IAdminUserDo
	WriteDB() IAdminUserDo
	As(alias string) gen.Dao
	Session(config *gorm.Session) IAdminUserDo
	Columns(cols ...field.Expr) gen.Columns
	Clauses(conds ...clause.Expression) IAdminUserDo
	Not(conds ...gen.Condition) IAdminUserDo
	Or(conds ...gen.Condition) IAdminUserDo
	Select(conds ...field.Expr) IAdminUserDo
	Where(conds ...gen.Condition) IAdminUserDo
	Order(conds ...field.Expr) IAdminUserDo
	Distinct(cols ...field.Expr) IAdminUserDo
	Omit(cols ...field.Expr) IAdminUserDo
	Join(table schema.Tabler, on ...field.Expr) IAdminUserDo
	LeftJoin(table schema.Tabler, on ...field.Expr) IAdminUserDo
	RightJoin(table schema.Tabler, on ...field.Expr) IAdminUserDo
	Group(cols ...field.Expr) IAdminUserDo
	Having(conds ...gen.Condition) IAdminUserDo
	Limit(limit int) IAdminUserDo
	Offset(offset int) IAdminUserDo
	Count() (count int64, err error)
	Scopes(funcs ...func(gen.Dao) gen.Dao) IAdminUserDo
	Unscoped() IAdminUserDo
	Create(values ...*model.AdminUser) error
	CreateInBatches(values []*model.AdminUser, batchSize int) error
	Save(values ...*model.AdminUser) error
	First() (*model.AdminUser, error)
	Take() (*model.AdminUser, error)
	Last() (*model.AdminUser, error)
	Find() ([]*model.AdminUser, error)
	FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.AdminUser, err error)
	FindInBatches(result *[]*model.AdminUser, batchSize int, fc func(tx gen.Dao, batch int) error) error
	Pluck(column field.Expr, dest interface{}) error
	Delete(...*model.AdminUser) (info gen.ResultInfo, err error)
	Update(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	Updates(value interface{}) (info gen.ResultInfo, err error)
	UpdateColumn(column field.Expr, value interface{}) (info gen.ResultInfo, err error)
	UpdateColumnSimple(columns ...field.AssignExpr) (info gen.ResultInfo, err error)
	UpdateColumns(value interface{}) (info gen.ResultInfo, err error)
	UpdateFrom(q gen.SubQuery) gen.Dao
	Attrs(attrs ...field.AssignExpr) IAdminUserDo
	Assign(attrs ...field.AssignExpr) IAdminUserDo
	Joins(fields ...field.RelationField) IAdminUserDo
	Preload(fields ...field.RelationField) IAdminUserDo
	FirstOrInit() (*model.AdminUser, error)
	FirstOrCreate() (*model.AdminUser, error)
	FindByPage(offset int, limit int) (result []*model.AdminUser, count int64, err error)
	ScanByPage(result interface{}, offset int, limit int) (count int64, err error)
	Scan(result interface{}) (err error)
	Returning(value interface{}, columns ...string) IAdminUserDo
	UnderlyingDB() *gorm.DB
	schema.Tabler
}

func (a adminUserDo) Debug() IAdminUserDo {
	return a.withDO(a.DO.Debug())
}

func (a adminUserDo) WithContext(ctx context.Context) IAdminUserDo {
	return a.withDO(a.DO.WithContext(ctx))
}

func (a adminUserDo) ReadDB() IAdminUserDo {
	return a.Clauses(dbresolver.Read)
}

func (a adminUserDo) WriteDB() IAdminUserDo {
	return a.Clauses(dbresolver.Write)
}

func (a adminUserDo) Session(config *gorm.Session) IAdminUserDo {
	return a.withDO(a.DO.Session(config))
}

func (a adminUserDo) Clauses(conds ...clause.Expression) IAdminUserDo {
	return a.withDO(a.DO.Clauses(conds...))
}

func (a adminUserDo) Returning(value interface{}, columns ...string) IAdminUserDo {
	return a.withDO(a.DO.Returning(value, columns...))
}

func (a adminUserDo) Not(conds ...gen.Condition) IAdminUserDo {
	return a.withDO(a.DO.Not(conds...))
}

func (a adminUserDo) Or(conds ...gen.Condition) IAdminUserDo {
	return a.withDO(a.DO.Or(conds...))
}

func (a adminUserDo) Select(conds ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Select(conds...))
}

func (a adminUserDo) Where(conds ...gen.Condition) IAdminUserDo {
	return a.withDO(a.DO.Where(conds...))
}

func (a adminUserDo) Order(conds ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Order(conds...))
}

func (a adminUserDo) Distinct(cols ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Distinct(cols...))
}

func (a adminUserDo) Omit(cols ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Omit(cols...))
}

func (a adminUserDo) Join(table schema.Tabler, on ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Join(table, on...))
}

func (a adminUserDo) LeftJoin(table schema.Tabler, on ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.LeftJoin(table, on...))
}

func (a adminUserDo) RightJoin(table schema.Tabler, on ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.RightJoin(table, on...))
}

func (a adminUserDo) Group(cols ...field.Expr) IAdminUserDo {
	return a.withDO(a.DO.Group(cols...))
}

func (a adminUserDo) Having(conds ...gen.Condition) IAdminUserDo {
	return a.withDO(a.DO.Having(conds...))
}

func (a adminUserDo) Limit(limit int) IAdminUserDo {
	return a.withDO(a.DO.Limit(limit))
}

func (a adminUserDo) Offset(offset int) IAdminUserDo {
	return a.withDO(a.DO.Offset(offset))
}

func (a adminUserDo) Scopes(funcs ...func(gen.Dao) gen.Dao) IAdminUserDo {
	return a.withDO(a.DO.Scopes(funcs...))
}

func (a adminUserDo) Unscoped() IAdminUserDo {
	return a.withDO(a.DO.Unscoped())
}

func (a adminUserDo) Create(values ...*model.AdminUser) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Create(values)
}

func (a adminUserDo) CreateInBatches(values []*model.AdminUser, batchSize int) error {
	return a.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (a adminUserDo) Save(values ...*model.AdminUser) error {
	if len(values) == 0 {
		return nil
	}
	return a.DO.Save(values)
}

func (a adminUserDo) First() (*model.AdminUser, error) {
	if result, err := a.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.AdminUser), nil
	}
}

func (a adminUserDo) Take() (*model.AdminUser, error) {
	if result, err := a.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.AdminUser), nil
	}
}

func (a adminUserDo) Last() (*model.AdminUser, error) {
	if result, err := a.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.AdminUser), nil
	}
}

func (a adminUserDo) Find() ([]*model.AdminUser, error) {
	result, err := a.DO.Find()
	return result.([]*model.AdminUser), err
}

func (a adminUserDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.AdminUser, err error) {
	buf := make([]*model.AdminUser, 0, batchSize)
	err = a.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (a adminUserDo) FindInBatches(result *[]*model.AdminUser, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return a.DO.FindInBatches(result, batchSize, fc)
}

func (a adminUserDo) Attrs(attrs ...field.AssignExpr) IAdminUserDo {
	return a.withDO(a.DO.Attrs(attrs...))
}

func (a adminUserDo) Assign(attrs ...field.AssignExpr) IAdminUserDo {
	return a.withDO(a.DO.Assign(attrs...))
}

func (a adminUserDo) Joins(fields ...field.RelationField) IAdminUserDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Joins(_f))
	}
	return &a
}

func (a adminUserDo) Preload(fields ...field.RelationField) IAdminUserDo {
	for _, _f := range fields {
		a = *a.withDO(a.DO.Preload(_f))
	}
	return &a
}

func (a adminUserDo) FirstOrInit() (*model.AdminUser, error) {
	if result, err := a.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.AdminUser), nil
	}
}

func (a adminUserDo) FirstOrCreate() (*model.AdminUser, error) {
	if result, err := a.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.AdminUser), nil
	}
}

func (a adminUserDo) FindByPage(offset int, limit int) (result []*model.AdminUser, count int64, err error) {
	result, err = a.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = a.Offset(-1).Limit(-1).Count()
	return
}

func (a adminUserDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = a.Count()
	if err != nil {
		return
	}

	err = a.Offset(offset).Limit(limit).Scan(result)
	return
}

func (a adminUserDo) Scan(result interface{}) (err error) {
	return a.DO.Scan(result)
}

func (a adminUserDo) Delete(models ...*model.AdminUser) (result gen.ResultInfo, err error) {
	return a.DO.Delete(models)
}

func (a *adminUserDo) withDO(do gen.Dao) *adminUserDo {
	a.DO = *do.(*gen.DO)
	return a
}
