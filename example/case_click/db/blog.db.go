// Code generated by protoc-gen-structify. DO NOT EDIT.
// source: example/case_click/db/blog.proto
// provider: clickhouse
// protoc-gen-structify: (version=v1.1.26-1-ga813178, branch=main, revision=a8131782f462b816801f2e5b739bf98e6e5d4a9c), build: (go=go1.23.6, date=2025-06-14T06:38:16+0300)
// protoc: 3.16.0
package db

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
)

//
// storages.
//

// blogStorages is a map of provider to init function.
type blogStorages struct {
	config *Config // configuration for the BlogStorages.

	deviceStorage  DeviceStorage
	postStorage    PostStorage
	messageStorage MessageStorage
	botStorage     BotStorage
	botViewStorage BotViewStorage
	commentStorage CommentStorage
	userStorage    UserStorage
	settingStorage SettingStorage
	addressStorage AddressStorage
}

// configuration for the BlogStorages.
type Config struct {
	DB driver.Conn

	QueryLogMethod func(ctx context.Context, table string, query string, args ...interface{})
	ErrorLogMethod func(ctx context.Context, err error, message string)
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
	// GetBotViewStorage returns the BotViewStorage store.
	GetBotViewStorage() BotViewStorage
	// GetCommentStorage returns the CommentStorage store.
	GetCommentStorage() CommentStorage
	// GetUserStorage returns the UserStorage store.
	GetUserStorage() UserStorage
	// GetSettingStorage returns the SettingStorage store.
	GetSettingStorage() SettingStorage
	// GetAddressStorage returns the AddressStorage store.
	GetAddressStorage() AddressStorage
}

// NewBlogStorages returns a new BlogStorages.
func NewBlogStorages(config *Config) (BlogStorages, error) {
	if config == nil {
		return nil, fmt.Errorf("config is required")
	}

	if config.DB == nil {
		return nil, fmt.Errorf("db is required")
	}

	var storages = blogStorages{
		config: config,
	}

	deviceStorageImpl, err := NewDeviceStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create DeviceStorage: %w", err)
	}
	storages.deviceStorage = deviceStorageImpl

	postStorageImpl, err := NewPostStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostStorage: %w", err)
	}
	storages.postStorage = postStorageImpl

	messageStorageImpl, err := NewMessageStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create MessageStorage: %w", err)
	}
	storages.messageStorage = messageStorageImpl

	botStorageImpl, err := NewBotStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create BotStorage: %w", err)
	}
	storages.botStorage = botStorageImpl

	botViewStorageImpl, err := NewBotViewStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create BotViewStorage: %w", err)
	}
	storages.botViewStorage = botViewStorageImpl

	commentStorageImpl, err := NewCommentStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create CommentStorage: %w", err)
	}
	storages.commentStorage = commentStorageImpl

	userStorageImpl, err := NewUserStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create UserStorage: %w", err)
	}
	storages.userStorage = userStorageImpl

	settingStorageImpl, err := NewSettingStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create SettingStorage: %w", err)
	}
	storages.settingStorage = settingStorageImpl

	addressStorageImpl, err := NewAddressStorage(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create AddressStorage: %w", err)
	}
	storages.addressStorage = addressStorageImpl

	return &storages, nil
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

// GetBotViewStorage returns the BotViewStorage store.
func (c *blogStorages) GetBotViewStorage() BotViewStorage {
	return c.botViewStorage
}

// GetCommentStorage returns the CommentStorage store.
func (c *blogStorages) GetCommentStorage() CommentStorage {
	return c.commentStorage
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

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &n.Data); err != nil {
		n.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	n.Valid = true
	return nil
}

// Value converts NullableJSON to a string representation for ClickHouse.
func (n *NullableJSON[T]) Value() (string, error) {
	if !n.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(n.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// ValueOrZero returns the value if valid, otherwise returns the zero value of type T.
func (n NullableJSON[T]) ValueOrZero() T {
	if !n.Valid {
		var zero T
		return zero
	}
	return n.Data
}

// Comment is a JSON type nested in another message.
type UserComment struct {
	Name string           `json:"name"`
	Meta *UserCommentMeta `json:"meta"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserComment) Scan(src interface{}) error {
	if str, ok := src.(string); ok {
		return json.Unmarshal([]byte(str), m)
	}
	return fmt.Errorf("can't convert %T to string", src)
}

// Value converts the struct to a JSON string.
func (m *UserComment) Value() (string, error) {
	if m == nil {
		m = &UserComment{}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Meta is a JSON type nested in another message.
type UserCommentMeta struct {
	Ip      string `json:"ip"`
	Browser string `json:"browser"`
	Os      string `json:"os"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserCommentMeta) Scan(src interface{}) error {
	if str, ok := src.(string); ok {
		return json.Unmarshal([]byte(str), m)
	}
	return fmt.Errorf("can't convert %T to string", src)
}

// Value converts the struct to a JSON string.
func (m *UserCommentMeta) Value() (string, error) {
	if m == nil {
		m = &UserCommentMeta{}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// NotificationSetting is a JSON type nested in another message.
type UserNotificationSetting struct {
	RegistrationEmail bool `json:"registration_email"`
	OrderEmail        bool `json:"order_email"`
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserNotificationSetting) Scan(src interface{}) error {
	if str, ok := src.(string); ok {
		return json.Unmarshal([]byte(str), m)
	}
	return fmt.Errorf("can't convert %T to string", src)
}

// Value converts the struct to a JSON string.
func (m *UserNotificationSetting) Value() (string, error) {
	if m == nil {
		m = &UserNotificationSetting{}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
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
	if str, ok := src.(string); ok {
		return json.Unmarshal([]byte(str), m)
	}
	return fmt.Errorf("can't convert %T to string", src)
}

// Value converts the struct to a JSON string.
func (m *UserNumr) Value() (string, error) {
	if m == nil {
		m = &UserNumr{}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

//
// Single repeated types.
//

// UserBallsRepeated is a JSON type nested in another message.
type UserBallsRepeated struct {
	Data  []int32
	Valid bool // Valid is true if the field is not NULL
}

// NewBallsField creates a new UserBallsRepeated.
func NewBallsField(v []int32) UserBallsRepeated {
	return UserBallsRepeated{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserBallsRepeated) Scan(src interface{}) error {
	if src == nil {
		m.Valid = false
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &m.Data); err != nil {
		m.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.Valid = true
	return nil
}

// Value converts the struct to a JSON string.
func (m UserBallsRepeated) Value() (string, error) {
	if !m.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Get returns the value of the field or the zero value if invalid.
func (m UserBallsRepeated) Get() []int32 {
	if !m.Valid {
		var zero []int32
		return zero
	}
	return m.Data
}

func (m UserBallsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserCommentsRepeated is a JSON type nested in another message.
type UserCommentsRepeated struct {
	Data  []UserComment
	Valid bool // Valid is true if the field is not NULL
}

// NewCommentsField creates a new UserCommentsRepeated.
func NewCommentsField(v []UserComment) UserCommentsRepeated {
	return UserCommentsRepeated{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserCommentsRepeated) Scan(src interface{}) error {
	if src == nil {
		m.Valid = false
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &m.Data); err != nil {
		m.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.Valid = true
	return nil
}

// Value converts the struct to a JSON string.
func (m UserCommentsRepeated) Value() (string, error) {
	if !m.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Get returns the value of the field or the zero value if invalid.
func (m UserCommentsRepeated) Get() []UserComment {
	if !m.Valid {
		var zero []UserComment
		return zero
	}
	return m.Data
}

func (m UserCommentsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserNumrsRepeated is a JSON type nested in another message.
type UserNumrsRepeated struct {
	Data  []UserNumr
	Valid bool // Valid is true if the field is not NULL
}

// NewNumrsField creates a new UserNumrsRepeated.
func NewNumrsField(v []UserNumr) UserNumrsRepeated {
	return UserNumrsRepeated{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserNumrsRepeated) Scan(src interface{}) error {
	if src == nil {
		m.Valid = false
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &m.Data); err != nil {
		m.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.Valid = true
	return nil
}

// Value converts the struct to a JSON string.
func (m UserNumrsRepeated) Value() (string, error) {
	if !m.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Get returns the value of the field or the zero value if invalid.
func (m UserNumrsRepeated) Get() []UserNumr {
	if !m.Valid {
		var zero []UserNumr
		return zero
	}
	return m.Data
}

func (m UserNumrsRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

// UserPhonesRepeated is a JSON type nested in another message.
type UserPhonesRepeated struct {
	Data  []string
	Valid bool // Valid is true if the field is not NULL
}

// NewPhonesField creates a new UserPhonesRepeated.
func NewPhonesField(v []string) UserPhonesRepeated {
	return UserPhonesRepeated{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *UserPhonesRepeated) Scan(src interface{}) error {
	if src == nil {
		m.Valid = false
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &m.Data); err != nil {
		m.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.Valid = true
	return nil
}

// Value converts the struct to a JSON string.
func (m UserPhonesRepeated) Value() (string, error) {
	if !m.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Get returns the value of the field or the zero value if invalid.
func (m UserPhonesRepeated) Get() []string {
	if !m.Valid {
		var zero []string
		return zero
	}
	return m.Data
}

func (m UserPhonesRepeated) String() string {
	return fmt.Sprintf("%v", m.Get())
}

//
// errors.
//

var (
	// ErrNotFound is returned when a record is not found.
	ErrRowNotFound = fmt.Errorf("row not found")
	// ErrNoTransaction is returned when a transaction is not provided.
	ErrNoTransaction = fmt.Errorf("no transaction provided")
	// ErrRowAlreadyExist is returned when a row already exist.
	ErrRowAlreadyExist = fmt.Errorf("row already exist")
	// ErrModelIsNil is returned when a relation model is nil.
	ErrModelIsNil = fmt.Errorf("model is nil")
)

//
// Transaction manager.
//

// QueryExecer is an interface that can execute queries.
type QueryExecer interface {
	driver.Conn
}

//
// Options.
//

// Option is a function that configures the BlogStorages.
type Option func(*Options)

// Options are the options for the BlogStorages.
type Options struct {
	// if true, then method was create/update relations
	relations bool
	// uniqField is the unique field.
	uniqField string
	// waitAsyncInsert is the wait flag. wait_for_async_insert = 1
	waitAsyncInsert bool
}

// WithWaitAsyncInsert sets the waitAsyncInsert flag.
func WithWaitAsyncInsert() Option {
	return func(o *Options) {
		o.waitAsyncInsert = true
	}
}

// WithRelations sets the relations flag.
// This is used to determine if the relations should be created or updated.
func WithRelations() Option {
	return func(o *Options) {
		o.relations = true
	}
}

// WithUniqField sets the unique field.
func WithUniqField(field string) Option {
	return func(o *Options) {
		o.uniqField = field
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
	// customTableName is the custom table name.
	customTableName string
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

// WithCustomTableName sets a custom table name for the query.
func (qb *QueryBuilder) WithCustomTableName(tableName string) *QueryBuilder {
	qb.customTableName = tableName
	return qb
}

// nullValue returns the null value.
func nullValue[T any](v *T) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

// Apply customTableName to the query.
func (qb *QueryBuilder) ApplyCustomTableName(query sq.SelectBuilder) sq.SelectBuilder {
	if qb.customTableName != "" {
		query = query.From(qb.customTableName)
	}
	return query
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
