package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"time"
)

// botViewStorage is a struct for the "bots_view" table.
type botViewStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// BotViewCRUDOperations is an interface for managing the bots_view table.
type BotViewCRUDOperations interface {
	Create(ctx context.Context, model *BotView, opts ...Option) error
	AsyncCreate(ctx context.Context, model *BotView, opts ...Option) error
	BatchCreate(ctx context.Context, models []*BotView, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*BotView, opts ...Option) error
}

// BotViewSearchOperations is an interface for searching the bots_view table.
type BotViewSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*BotView, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*BotView, error)
}

type BotViewSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) BotViewStorage
	SetQueryBuilder(builder sq.StatementBuilderType) BotViewStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// BotViewRelationLoading is an interface for loading relations.
type BotViewRelationLoading interface {
	LoadUser(ctx context.Context, model *BotView, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*BotView, builders ...*QueryBuilder) error
}

// BotViewRawQueryOperations is an interface for executing raw queries.
type BotViewRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// BotViewStorage is a struct for the "bots_view" table.
type BotViewStorage interface {
	BotViewCRUDOperations
	BotViewSearchOperations
	BotViewRelationLoading
	BotViewRawQueryOperations
	BotViewSettings
}

// NewBotViewStorage returns a new botViewStorage.
func NewBotViewStorage(config *Config) (BotViewStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &botViewStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *botViewStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *botViewStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// TableName returns the table name.
func (t *botViewStorage) TableName() string {
	return "bots_view"
}

// GetQueryBuilder returns the query builder.
func (t *botViewStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *botViewStorage) Columns() []string {
	return []string{
		"id", "user_id", "name", "token", "is_publish", "created_at", "updated_at", "deleted_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *botViewStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *botViewStorage) SetConfig(config *Config) BotViewStorage {
	t.config = config
	return t
}

func (t *botViewStorage) SetQueryBuilder(builder sq.StatementBuilderType) BotViewStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *botViewStorage) LoadUser(ctx context.Context, model *BotView, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.UserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one UserStorage: %w", err)
	}

	model.User = relationModel
	return nil
}

// LoadBatchUser loads the User relation.
func (t *botViewStorage) LoadBatchUser(ctx context.Context, items []*BotView, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.UserId)
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(UserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many UserStorage: %w", err)
	}
	resultMap := make(map[interface{}]*User)
	for _, result := range results {
		resultMap[result.Id] = result
	}

	// Assign User to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.UserId]; ok {
			item.User = v
		}
	}

	return nil
}

// BotView is a struct for the "bots_view" table.
type BotView struct {
	Id        string
	UserId    string
	Name      string
	Token     string
	IsPublish bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	User      *User
}

// TableName returns the table name.
func (t *BotView) TableName() string {
	return "bots_view"
}

// ScanRow scans a row into a BotView.
func (t *BotView) ScanRow(row driver.Row) error {
	return row.Scan(
		&t.Id,
		&t.UserId,
		&t.Name,
		&t.Token,
		&t.IsPublish,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.DeletedAt,
	)
}

// BotViewFilters is a struct that holds filters for BotView.
type BotViewFilters struct {
	Id        *string
	UserId    *string
	CreatedAt *time.Time
}

// BotViewIdEq returns a condition that checks if the field equals the value.
func BotViewIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// BotViewUserIdEq returns a condition that checks if the field equals the value.
func BotViewUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtEq returns a condition that checks if the field equals the value.
func BotViewCreatedAtEq(value time.Time) FilterApplier {
	return EqualsCondition{Field: "created_at", Value: value}
}

// BotViewIdNotEq returns a condition that checks if the field equals the value.
func BotViewIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// BotViewUserIdNotEq returns a condition that checks if the field equals the value.
func BotViewUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtNotEq returns a condition that checks if the field equals the value.
func BotViewCreatedAtNotEq(value time.Time) FilterApplier {
	return NotEqualsCondition{Field: "created_at", Value: value}
}

// BotViewIdGT greaterThanCondition than condition.
func BotViewIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// BotViewUserIdGT greaterThanCondition than condition.
func BotViewUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtGT greaterThanCondition than condition.
func BotViewCreatedAtGT(value time.Time) FilterApplier {
	return GreaterThanCondition{Field: "created_at", Value: value}
}

// BotViewIdLT less than condition.
func BotViewIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// BotViewUserIdLT less than condition.
func BotViewUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtLT less than condition.
func BotViewCreatedAtLT(value time.Time) FilterApplier {
	return LessThanCondition{Field: "created_at", Value: value}
}

// BotViewIdGTE greater than or equal condition.
func BotViewIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// BotViewUserIdGTE greater than or equal condition.
func BotViewUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtGTE greater than or equal condition.
func BotViewCreatedAtGTE(value time.Time) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotViewIdLTE less than or equal condition.
func BotViewIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// BotViewUserIdLTE less than or equal condition.
func BotViewUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotViewCreatedAtLTE less than or equal condition.
func BotViewCreatedAtLTE(value time.Time) FilterApplier {
	return LessThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotViewIdBetween between condition.
func BotViewIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// BotViewUserIdBetween between condition.
func BotViewUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// BotViewCreatedAtBetween between condition.
func BotViewCreatedAtBetween(min, max time.Time) FilterApplier {
	return BetweenCondition{Field: "created_at", Min: min, Max: max}
}

// BotViewIdILike iLike condition %
func BotViewIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// BotViewUserIdILike iLike condition %
func BotViewUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// BotViewIdLike like condition %
func BotViewIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// BotViewUserIdLike like condition %
func BotViewUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// BotViewIdNotLike not like condition
func BotViewIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// BotViewUserIdNotLike not like condition
func BotViewUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// BotViewIdIn condition
func BotViewIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// BotViewUserIdIn condition
func BotViewUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// BotViewCreatedAtIn condition
func BotViewCreatedAtIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "created_at", Values: values}
}

// BotViewIdNotIn not in condition
func BotViewIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// BotViewUserIdNotIn not in condition
func BotViewUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// BotViewCreatedAtNotIn not in condition
func BotViewCreatedAtNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "created_at", Values: values}
}

// BotViewIdOrderBy sorts the result in ascending order.
func BotViewIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// BotViewUserIdOrderBy sorts the result in ascending order.
func BotViewUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// BotViewCreatedAtOrderBy sorts the result in ascending order.
func BotViewCreatedAtOrderBy(asc bool) FilterApplier {
	return OrderBy("created_at", asc)
}

// AsyncCreate asynchronously inserts a new BotView.
func (t *botViewStorage) AsyncCreate(ctx context.Context, model *BotView, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("bots_view").
		Columns(
			"user_id",
			"name",
			"token",
			"is_publish",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		Values(
			model.UserId,
			model.Name,
			model.Token,
			model.IsPublish,
			model.CreatedAt,
			model.UpdatedAt,
			nullValue(model.DeletedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create BotView: %w", err)
	}

	return nil
}

// Create creates a new BotView.
func (t *botViewStorage) Create(ctx context.Context, model *BotView, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("bots_view").
		Columns(
			"user_id",
			"name",
			"token",
			"is_publish",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		Values(
			model.UserId,
			model.Name,
			model.Token,
			model.IsPublish,
			model.CreatedAt,
			model.UpdatedAt,
			nullValue(model.DeletedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create BotView: %w", err)
	}

	return nil
}

// BatchCreate creates multiple BotView records in a single batch.
func (t *botViewStorage) BatchCreate(ctx context.Context, models []*BotView, opts ...Option) error {
	if len(models) == 0 {
		return fmt.Errorf("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return fmt.Errorf("relations are not supported in batch create")
	}

	batch, err := t.DB().PrepareBatch(ctx, "INSERT INTO "+t.TableName(), driver.WithReleaseConnection())
	if err != nil {
		return fmt.Errorf("failed to prepare batch: %w", err)
	}

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("one of the models is nil")
		}

		err := batch.Append(
			model.UserId,
			model.Name,
			model.Token,
			model.IsPublish,
			model.CreatedAt,
			model.UpdatedAt,
			nullValue(model.DeletedAt),
		)
		if err != nil {
			return fmt.Errorf("failed to append to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("failed to execute batch insert: %w", err)
	}

	return nil
}

// OriginalBatchCreate creates multiple BotView records in a single batch.
func (t *botViewStorage) OriginalBatchCreate(ctx context.Context, models []*BotView, opts ...Option) error {
	if len(models) == 0 {
		return fmt.Errorf("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return fmt.Errorf("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"user_id",
			"name",
			"token",
			"is_publish",
			"created_at",
			"updated_at",
			"deleted_at",
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}

		query = query.Values(
			model.UserId,
			model.Name,
			model.Token,
			model.IsPublish,
			model.CreatedAt,
			model.UpdatedAt,
			nullValue(model.DeletedAt),
		)
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().Query(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to execute bulk insert: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration error: %w", err)
	}

	return nil
}

// FindMany finds multiple BotView based on the provided options.
func (t *botViewStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*BotView, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

	// set default options
	options := &Options{}

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply custom table name
		query = builder.ApplyCustomTableName(query)

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)

		// apply pagination
		if builder.pagination != nil {
			if builder.pagination.limit != nil {
				query = query.Limit(*builder.pagination.limit)
			}
			if builder.pagination.offset != nil {
				query = query.Offset(*builder.pagination.offset)
			}
		}

		// apply sorting
		for _, option := range builder.sortOptions {
			query = option.Apply(query)
		}

		// apply options
		for _, o := range builder.options {
			o(options)
		}
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*BotView
	for rows.Next() {
		model := &BotView{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan BotView: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single BotView based on the provided options.
func (t *botViewStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*BotView, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne BotView: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *botViewStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *botViewStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *botViewStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *botViewStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *botViewStorage) Conn() driver.Conn {
	return t.DB()
}
