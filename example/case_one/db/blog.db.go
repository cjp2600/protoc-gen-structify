// Code generated by protoc-gen-structify. DO NOT EDIT.
// source: example/case_one/db/blog.proto
// provider: postgres
// protoc-gen-structify: (version=v1.1.9, branch=main, revision=dfe8ab858c4222c79d71093b74b4e2b66a9f7f2b), build: (go=go1.23.5, date=2025-01-28T11:21:01+0300)
// protoc: 3.15.8
package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"strings"
)

//
// storages.
//

// blogStorages is a map of provider to init function.
type blogStorages struct {
	db *DB        // The database connection.
	tx *TxManager // The transaction manager.

	deviceStorage  DeviceStorage
	postStorage    PostStorage
	messageStorage MessageStorage
	botStorage     BotStorage
	userStorage    UserStorage
	settingStorage SettingStorage
	addressStorage AddressStorage
}

type DB struct {
	DBRead  *sql.DB
	DBWrite *sql.DB
}

// BlogStorages is the interface for the BlogStorages.
type BlogStorages interface {
	// GetDeviceStorage returns the DeviceStorage store.
	GetDeviceStorage() DeviceStorage
	// GetPostStorage returns the PostStorage store.
	GetPostStorage() PostStorage
	// GetMessageStorage returns the MessageStorage store.
	GetMessageStorage() MessageStorage
	// GetBotStorage returns the BotStorage store.
	GetBotStorage() BotStorage
	// GetUserStorage returns the UserStorage store.
	GetUserStorage() UserStorage
	// GetSettingStorage returns the SettingStorage store.
	GetSettingStorage() SettingStorage
	// GetAddressStorage returns the AddressStorage store.
	GetAddressStorage() AddressStorage
	// TxManager returns the transaction manager.
	TxManager() *TxManager
}

// NewBlogStorages returns a new BlogStorages.
func NewBlogStorages(db *DB) BlogStorages {
	if db == nil {
		panic("structify: db is required")
	}

	if db.DBRead == nil {
		panic("structify: dbRead is required")
	}

	if db.DBWrite == nil {
		db.DBWrite = db.DBRead
	}

	return &blogStorages{
		db: db,
		tx: NewTxManager(db.DBWrite),

		deviceStorage:  NewDeviceStorage(db),
		postStorage:    NewPostStorage(db),
		messageStorage: NewMessageStorage(db),
		botStorage:     NewBotStorage(db),
		userStorage:    NewUserStorage(db),
		settingStorage: NewSettingStorage(db),
		addressStorage: NewAddressStorage(db),
	}
}

// TxManager returns the transaction manager.
func (c *blogStorages) TxManager() *TxManager {
	return c.tx
}

// GetDeviceStorage returns the DeviceStorage store.
func (c *blogStorages) GetDeviceStorage() DeviceStorage {
	return c.deviceStorage
}

// GetPostStorage returns the PostStorage store.
func (c *blogStorages) GetPostStorage() PostStorage {
	return c.postStorage
}

// GetMessageStorage returns the MessageStorage store.
func (c *blogStorages) GetMessageStorage() MessageStorage {
	return c.messageStorage
}

// GetBotStorage returns the BotStorage store.
func (c *blogStorages) GetBotStorage() BotStorage {
	return c.botStorage
}

// GetUserStorage returns the UserStorage store.
func (c *blogStorages) GetUserStorage() UserStorage {
	return c.userStorage
}

// GetSettingStorage returns the SettingStorage store.
func (c *blogStorages) GetSettingStorage() SettingStorage {
	return c.settingStorage
}

// GetAddressStorage returns the AddressStorage store.
func (c *blogStorages) GetAddressStorage() AddressStorage {
	return c.addressStorage
}

//
// Json types.
//

// NullableJSON represents a JSON field that can be null.
type NullableJSON[T any] struct {
	Data  T
	Valid bool // Valid is true if the field is not NULL
}

// NewNullableJSON creates a new NullableJSON with a value.
func NewNullableJSON[T any](v T) NullableJSON[T] {
	return NullableJSON[T]{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface.
func (n *NullableJSON[T]) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, &n.Data); err != nil {
		n.Valid = false
		return errors.Wrap(err, "failed to unmarshal json")
	}

	n.Valid = true
	return nil
}

func (n *NullableJSON[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return json.Marshal(n.Data)
}

// ValueOrZero returns the value if valid, otherwise returns the zero value of type T.
func (n NullableJSON[T]) ValueOrZero() T {
	if !n.Valid {
		var zero T // This declares a variable of type T initialized to its zero value
		return zero
	}
	return n.Data
}

// Comment is a JSON type nested in another message.
type UserComment struct {
	Name string       `json:"name"`
	Meta *CommentMeta `json:"meta"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserComment) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m *UserComment) Value() (driver.Value, error) {
	if m == nil {
		m = &UserComment{}
	}
	return json.Marshal(m)
}

// Meta is a JSON type nested in another message.
type CommentMeta struct {
	Ip      string `json:"ip"`
	Browser string `json:"browser"`
	Os      string `json:"os"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *CommentMeta) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m *CommentMeta) Value() (driver.Value, error) {
	if m == nil {
		m = &CommentMeta{}
	}
	return json.Marshal(m)
}

// NotificationSetting is a JSON type nested in another message.
type UserNotificationSetting struct {
	RegistrationEmail bool `json:"registration_email"`
	OrderEmail        bool `json:"order_email"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserNotificationSetting) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m *UserNotificationSetting) Value() (driver.Value, error) {
	if m == nil {
		m = &UserNotificationSetting{}
	}
	return json.Marshal(m)
}

// Numr is a JSON type nested in another message.
type UserNumr struct {
	Street string `json:"street"`
	City   string `json:"city"`
	State  int32  `json:"state"`
	Zip    int64  `json:"zip"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserNumr) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m *UserNumr) Value() (driver.Value, error) {
	if m == nil {
		m = &UserNumr{}
	}
	return json.Marshal(m)
}

//
// Single repeated types.
//

// UserBallsRepeated is a JSON type nested in another message.
type UserBallsRepeated []int32

// NewBallsField returns a new UserBallsRepeated.
func NewBallsField(v []int32) UserBallsRepeated {
	return v
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserBallsRepeated) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m UserBallsRepeated) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m UserBallsRepeated) Get() []int32 {
	return m
}

func (m *UserBallsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserCommentsRepeated is a JSON type nested in another message.
type UserCommentsRepeated []UserComment

// NewCommentsField returns a new UserCommentsRepeated.
func NewCommentsField(v []UserComment) UserCommentsRepeated {
	return v
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserCommentsRepeated) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m UserCommentsRepeated) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m UserCommentsRepeated) Get() []UserComment {
	return m
}

func (m *UserCommentsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserNumrsRepeated is a JSON type nested in another message.
type UserNumrsRepeated []UserNumr

// NewNumrsField returns a new UserNumrsRepeated.
func NewNumrsField(v []UserNumr) UserNumrsRepeated {
	return v
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserNumrsRepeated) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m UserNumrsRepeated) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m UserNumrsRepeated) Get() []UserNumr {
	return m
}

func (m *UserNumrsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserPhonesRepeated is a JSON type nested in another message.
type UserPhonesRepeated []string

// NewPhonesField returns a new UserPhonesRepeated.
func NewPhonesField(v []string) UserPhonesRepeated {
	return v
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserPhonesRepeated) Scan(src interface{}) error {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m UserPhonesRepeated) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m UserPhonesRepeated) Get() []string {
	return m
}

func (m *UserPhonesRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// Pagination is the pagination.
type Paginator struct {
	TotalCount int64
	Limit      int
	Page       int
	TotalPages int
}

//
// errors.
//

var (
	// ErrNotFound is returned when a record is not found.
	ErrRowNotFound = errors.New("row not found")
	// ErrNoTransaction is returned when a transaction is not provided.
	ErrNoTransaction = errors.New("no transaction provided")
	// ErrRowAlreadyExist is returned when a row already exist.
	ErrRowAlreadyExist = errors.New("row already exist")
	// ErrModelIsNil is returned when a relation model is nil.
	ErrModelIsNil = errors.New("model is nil")
)

//
// Transaction manager.
//

// txKey is the key used to store the transaction in the context.
type txKey struct{}

// TxFromContext returns the transaction from the context.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

// TxManager is a transaction manager.
type TxManager struct {
	db *sql.DB
}

// NewTxManager creates a new transaction manager.
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{
		db: db,
	}
}

// Begin begins a transaction.
func (m *TxManager) Begin(ctx context.Context) (context.Context, error) {
	if _, ok := TxFromContext(ctx); ok {
		return ctx, nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return ctx, errors.Wrap(err, "could not begin transaction")
	}

	// store the transaction in the context.
	return context.WithValue(ctx, txKey{}, tx), nil
}

// IsTxOpen returns true if a transaction is open.
func (m *TxManager) Commit(ctx context.Context) error {
	tx, ok := TxFromContext(ctx)
	if !ok {
		return errors.New("transactions wasn't opened")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "could not commit transaction")
	}

	return nil
}

// Rollback rolls back a transaction.
func (m *TxManager) Rollback(ctx context.Context) error {
	if tx, ok := TxFromContext(ctx); ok {
		err := tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return err
		}
	}

	return nil
}

// ExecFuncWithTx executes a function with a transaction.
func (m *TxManager) ExecFuncWithTx(ctx context.Context, f func(context.Context) error) error {
	// if a transaction is already open, just execute the function.
	if m.IsTxOpen(ctx) {
		return f(ctx)
	}

	ctx, err := m.Begin(ctx)
	if err != nil {
		return err
	}
	// rollback the transaction if there is an error.
	defer func() { _ = m.Rollback(ctx) }()

	if err := f(ctx); err != nil {
		return err
	}

	// commit the transaction.
	if err := m.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// IsTxOpen returns true if a transaction is open.
func (m *TxManager) IsTxOpen(ctx context.Context) bool {
	_, ok := TxFromContext(ctx)
	return ok
}

// QueryExecer is an interface that can execute queries.
type QueryExecer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// IsPgCheckViolation returns true if the error is a postgres check violation.
func IsPgUniqueViolation(err error) bool {
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return pgErr.Code == errPgUniqueViolationError
}

// IsPgCheckViolation returns true if the error is a postgres check violation.
func IsPgViolationError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return pgErr.Code == errPgCheckViolation ||
		pgErr.Code == errPgNotNullViolation ||
		pgErr.Code == errPgForeignKeyViolation ||
		pgErr.Code == errPgUniqueViolationError
}

// PgPrettyErr returns a pretty postgres error.
func PgPrettyErr(err error) error {
	if pgErr, ok := err.(*pq.Error); ok {
		return errors.New(pgErr.Detail)
	}
	return err
}

// errors for postgres.
// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
const (
	errPgCheckViolation       = "23514"
	errPgNotNullViolation     = "23502"
	errPgForeignKeyViolation  = "23503"
	errPgUniqueViolationError = "23505"
)

//
// Options.
//

// Option is a function that configures the BlogStorages.
type Option func(*Options)

// Options are the options for the BlogStorages.
type Options struct {
	// if true, then method was create/update relations
	relations bool
}

// WithRelations sets the relations flag.
// This is used to determine if the relations should be created or updated.
func WithRelations() Option {
	return func(o *Options) {
		o.relations = true
	}
}

// FilterApplier is a condition filters.
type FilterApplier interface {
	Apply(query sq.SelectBuilder) sq.SelectBuilder
	ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder
}

// CustomFilter is a custom filter.
type CustomFilter interface {
	ApplyFilter(query sq.SelectBuilder, params any) sq.SelectBuilder
}

// QueryBuilder is a query builder.
type QueryBuilder struct {
	// additional options for the query.
	options []Option
	// filterOptions are the filter options.
	filterOptions []FilterApplier
	// orderOptions are the order options.
	sortOptions []FilterApplier
	// pagination is the pagination.
	pagination *Pagination
	// customFilters are the custom filters.
	customFilters []struct {
		filter CustomFilter
		params any
	}
}

// NewQueryBuilder returns a new query builder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// WithOptions sets the options for the query.
func (b *QueryBuilder) WithOptions(options ...Option) *QueryBuilder {
	b.options = options
	return b
}

// WithCustomFilter sets a custom filter for the query.
func (qb *QueryBuilder) WithCustomFilter(filter CustomFilter, params any) *QueryBuilder {
	qb.customFilters = append(qb.customFilters, struct {
		filter CustomFilter
		params any
	}{
		filter: filter,
		params: params,
	})
	return qb
}

// ApplyCustomFilters applies the custom filters to the query.
func (qb *QueryBuilder) ApplyCustomFilters(query sq.SelectBuilder) sq.SelectBuilder {
	for _, cf := range qb.customFilters {
		query = cf.filter.ApplyFilter(query, cf.params)
	}
	return query
}

// WithFilterOptions sets the filter options for the query.
func (b *QueryBuilder) WithFilter(filterOptions ...FilterApplier) *QueryBuilder {
	b.filterOptions = filterOptions
	return b
}

// WithSort sets the sort options for the query.
func (b *QueryBuilder) WithSort(sortOptions ...FilterApplier) *QueryBuilder {
	b.sortOptions = sortOptions
	return b
}

// WithPagination sets the pagination for the query.
func (b *QueryBuilder) WithPagination(pagination *Pagination) *QueryBuilder {
	b.pagination = pagination
	return b
}

// Filter is a helper function to create a new query builder with filter options.
func FilterBuilder(filterOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithFilter(filterOptions...)
}

// SortBuilder is a helper function to create a new query builder with sort options.
func SortBuilder(sortOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithSort(sortOptions...)
}

// Options is a helper function to create a new query builder with options.
func LimitBuilder(limit uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(&Pagination{
		limit: &limit,
	})
}

// Offset is a helper function to create a new query builder with options.
func OffsetBuilder(offset uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(&Pagination{
		offset: &offset,
	})
}

// Paginate is a helper function to create a new query builder with options.
func PaginateBuilder(limit, offset uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(NewPagination(limit, offset))
}

// Pagination is the pagination.
type Pagination struct {
	// limit is the limit.
	limit *uint64
	// offset is the offset.
	offset *uint64
}

// NewPagination returns a new pagination.
// If limit or offset are nil, then they will be omitted.
func NewPagination(limit, offset uint64) *Pagination {
	return &Pagination{
		limit:  &limit,
		offset: &offset,
	}
}

// Limit is a helper function to create a new pagination.
func Limit(limit uint64) *Pagination {
	return &Pagination{
		limit: &limit,
	}
}

// Offset is a helper function to create a new pagination.
func Offset(offset uint64) *Pagination {
	return &Pagination{
		offset: &offset,
	}
}

//
// Conditions for query builder.
//

type Table interface {
	TableName() string
}

type JoinType string

const (
	LeftJoin  JoinType = "LEFT"
	InnerJoin JoinType = "INNER"
	RightJoin JoinType = "RIGHT"
)

type JoinCondition struct {
	Type  JoinType
	Table Table
	On    FilterApplier
}

func Join(joinType JoinType, table Table, on FilterApplier) FilterApplier {
	return JoinCondition{Type: joinType, Table: table, On: on}
}

func toInterface[T any](s []T) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

func (c JoinCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	onQuery := c.On.Apply(sq.Select("*"))
	onClause, args, _ := onQuery.ToSql()
	onClause = strings.TrimPrefix(onClause, "SELECT * WHERE ")
	joinExpr := fmt.Sprintf("%s JOIN %s ON %s", c.Type, c.Table.TableName(), onClause)
	return query.JoinClause(sq.Expr(joinExpr, args...))
}

func (c JoinCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}

// And returns a condition that combines the given conditions with AND.
type AndCondition struct {
	Where []FilterApplier
}

// And returns a condition that combines the given conditions with AND.
func And(conditions ...FilterApplier) FilterApplier {
	return AndCondition{Where: conditions}
}

// And returns a condition that combines the given conditions with AND.
func (c AndCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	for _, condition := range c.Where {
		query = condition.Apply(query)
	}
	return query
}

// And returns a condition that combines the given conditions with AND.
func (c AndCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	for _, condition := range c.Where {
		query = condition.ApplyDelete(query)
	}
	return query
}

//
// Or returns a condition that checks if any of the conditions are true.
//

// Or returns a condition that checks if any of the conditions are true.
type OrCondition struct {
	Conditions []FilterApplier
}

// Or returns a condition that checks if any of the conditions are true.
func Or(conditions ...FilterApplier) FilterApplier {
	return OrCondition{Conditions: conditions}
}

// Apply applies the condition to the query.
func (c OrCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	or := sq.Or{}
	for _, condition := range c.Conditions {
		subQuery := condition.Apply(sq.Select("*"))
		// Extract WHERE clause from the subquery
		whereParts, args, _ := subQuery.ToSql()
		whereParts = strings.TrimPrefix(whereParts, "SELECT * WHERE ")
		// Append the WHERE clause to the OR condition
		or = append(or, sq.Expr(whereParts, args...))
	}
	return query.Where(or)
}

// Apply applies the condition to the query.
func (c OrCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	or := sq.Or{}
	for _, condition := range c.Conditions {
		subQuery := condition.Apply(sq.Select("*"))
		// Extract WHERE clause from the subquery
		whereParts, args, _ := subQuery.ToSql()
		whereParts = strings.TrimPrefix(whereParts, "SELECT * WHERE ")
		// Append the WHERE clause to the OR condition
		or = append(or, sq.Expr(whereParts, args...))
	}
	return query.Where(or)
}

// EqualsCondition equals condition.
type EqualsCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c EqualsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Eq{c.Field: c.Value})
}

func (c EqualsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Eq{c.Field: c.Value})
}

// Eq returns a condition that checks if the field equals the value.
func Eq(field string, value interface{}) FilterApplier {
	return EqualsCondition{Field: field, Value: value}
}

// BetweenCondition
type BetweenCondition struct {
	Field string
	Min   interface{}
	Max   interface{}
}

// Apply applies the condition to the query.
func (c BetweenCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.Min, c.Max))
}

// ApplyDelete applies the condition to the query.
func (c BetweenCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.Min, c.Max))
}

// Between returns a condition that checks if the field is between the min and max values.
func Between(field string, min, max interface{}) FilterApplier {
	return BetweenCondition{Field: field, Min: min, Max: max}
}

// NotEqualsCondition not equals condition.
type NotEqualsCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c NotEqualsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c NotEqualsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotEq{c.Field: c.Value})
}

// NotEq returns a condition that checks if the field equals the value.
func NotEq(field string, value interface{}) FilterApplier {
	return NotEqualsCondition{Field: field, Value: value}
}

// GreaterThanCondition greaterThanCondition than condition.
type GreaterThanCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c GreaterThanCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Gt{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c GreaterThanCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Gt{c.Field: c.Value})
}

// GreaterThan returns a condition that checks if the field equals the value.
func GreaterThan(field string, value interface{}) FilterApplier {
	return GreaterThanCondition{Field: field, Value: value}
}

// LessThanCondition less than condition.
type LessThanCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LessThanCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Lt{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LessThanCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Lt{c.Field: c.Value})
}

// LessThan returns a condition that checks if the field equals the value.
func LessThan(field string, value interface{}) FilterApplier {
	return LessThanCondition{Field: field, Value: value}
}

// LessThanOrEqualCondition less than or equal condition.
type GreaterThanOrEqualCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c GreaterThanOrEqualCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.GtOrEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c GreaterThanOrEqualCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.GtOrEq{c.Field: c.Value})
}

// GreaterThanOrEqual returns a condition that checks if the field equals the value.
func GreaterThanOrEq(field string, value interface{}) FilterApplier {
	return GreaterThanOrEqualCondition{Field: field, Value: value}
}

// LessThanOrEqualCondition less than or equal condition.
type LessThanOrEqualCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LessThanOrEqualCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.LtOrEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LessThanOrEqualCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.LtOrEq{c.Field: c.Value})
}

func LessThanOrEq(field string, value interface{}) FilterApplier {
	return LessThanOrEqualCondition{Field: field, Value: value}
}

// ILikeCondition ilike condition.
type ILikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c ILikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.ILike{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c ILikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.ILike{c.Field: c.Value})
}

// ILike returns a condition that checks if the field equals the value.
func ILike(field string, value interface{}) FilterApplier {
	return ILikeCondition{Field: field, Value: value}
}

// LikeCondition like condition.
type LikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Like{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Like{c.Field: c.Value})
}

// Like returns a condition that checks if the field equals the value.
func Like(field string, value interface{}) FilterApplier {
	return LikeCondition{Field: field, Value: value}
}

// NotLikeCondition not like condition.
type NotLikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c NotLikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotLike{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c NotLikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotLike{c.Field: c.Value})
}

// NotLike returns a condition that checks if the field equals the value.
func NotLike(field string, value interface{}) FilterApplier {
	return NotLikeCondition{Field: field, Value: value}
}

// IsNullCondition represents the IS NULL condition.
type IsNullCondition struct {
	Field string
}

// Apply applies the condition to the query.
func (c IsNullCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(c.Field + " IS NULL"))
}

// ApplyDelete applies the condition to the query.
func (c IsNullCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(c.Field + " IS NULL"))
}

// IsNull returns a condition that checks if the field is null.
func IsNull(field string) FilterApplier {
	return IsNullCondition{Field: field}
}

// IsNotNullCondition represents the IS NOT NULL condition.
type IsNotNullCondition struct {
	Field string
}

// Apply applies the condition to the query.
func (c IsNotNullCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(c.Field + " IS NOT NULL"))
}

// ApplyDelete applies the condition to the query.
func (c IsNotNullCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(c.Field + " IS NOT NULL"))
}

// IsNotNull returns a condition that checks if the field is not null.
func IsNotNull(field string) FilterApplier {
	return IsNotNullCondition{Field: field}
}

// InCondition represents the IN condition.
type InCondition struct {
	Field  string
	Values []interface{}
}

// Apply applies the condition to the query.
func (c InCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Eq{c.Field: c.Values})
}

// ApplyDelete applies the condition to the query.
func (c InCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Eq{c.Field: c.Values})
}

// In returns a condition that checks if the field is in the given values.
func In(field string, values ...interface{}) FilterApplier {
	return InCondition{Field: field, Values: values}
}

// NotInCondition represents the NOT IN condition.
type NotInCondition struct {
	Field  string
	Values []interface{}
}

// Apply applies the condition to the query.
func (c NotInCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotEq{c.Field: c.Values})
}

// ApplyDelete applies the condition to the query.
func (c NotInCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotEq{c.Field: c.Values})
}

// NotIn returns a condition that checks if the field is not in the given values.
func NotIn(field string, values ...interface{}) FilterApplier {
	return NotInCondition{Field: field, Values: values}
}

// OrderCondition represents the ORDER BY condition.
type OrderCondition struct {
	Column string
	Asc    bool
}

// Apply applies the condition to the query.
func OrderBy(column string, asc bool) FilterApplier {
	return OrderCondition{Column: column, Asc: asc}
}

// Apply applies the condition to the query.
func (c OrderCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if c.Asc {
		return query.OrderBy(c.Column + " ASC")
	}

	// default to descending.
	return query.OrderBy(c.Column + " DESC")
}

// ApplyDelete applies the condition to the query.
func (c OrderCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}
