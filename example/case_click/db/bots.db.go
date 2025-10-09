package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
	"time"
)

// botStorage is a struct for the "bots" table.
type botStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// BotCRUDOperations is an interface for managing the bots table.
type BotCRUDOperations interface {
	Create(ctx context.Context, model *Bot, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Bot, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Bot, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Bot, opts ...Option) error
}

// BotSearchOperations is an interface for searching the bots table.
type BotSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Bot, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Bot, error)
}

type BotSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) BotStorage
	SetQueryBuilder(builder sq.StatementBuilderType) BotStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// BotRelationLoading is an interface for loading relations.
type BotRelationLoading interface {
	LoadUser(ctx context.Context, model *Bot, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Bot, builders ...*QueryBuilder) error
}

// BotRawQueryOperations is an interface for executing raw queries.
type BotRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// BotStorage is a struct for the "bots" table.
type BotStorage interface {
	BotCRUDOperations
	BotSearchOperations
	BotRelationLoading
	BotRawQueryOperations
	BotSettings
}

// NewBotStorage returns a new botStorage.
func NewBotStorage(config *Config) (BotStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &botStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *botStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *botStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *botStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
	if len(conditions) == 0 {
		return query, args
	}

	// Build PREWHERE conditions using a temporary query builder
	prewhereQuery := t.queryBuilder.Select("*")
	for _, condition := range conditions {
		prewhereQuery = condition.Apply(prewhereQuery)
	}

	// Extract WHERE clause from the temporary query
	prewhereSql, prewhereArgs, err := prewhereQuery.ToSql()
	if err != nil {
		return query, args
	}

	// Extract just the WHERE part and convert it to PREWHERE
	whereIdx := strings.Index(prewhereSql, "WHERE ")
	if whereIdx == -1 {
		return query, args
	}

	prewhereClause := strings.TrimPrefix(prewhereSql[whereIdx:], "WHERE ")

	// Find the position to insert PREWHERE (after FROM and before WHERE/ORDER BY/LIMIT)
	// Split query to find WHERE position
	wherePos := strings.Index(query, " WHERE ")
	orderPos := strings.Index(query, " ORDER BY ")
	limitPos := strings.Index(query, " LIMIT ")

	insertPos := len(query)
	if wherePos != -1 {
		insertPos = wherePos
	} else if orderPos != -1 {
		insertPos = orderPos
	} else if limitPos != -1 {
		insertPos = limitPos
	}

	// Insert PREWHERE clause
	prewhereClauseFormatted := "\nPREWHERE " + prewhereClause
	newQuery := query[:insertPos] + prewhereClauseFormatted + query[insertPos:]

	// Prepend PREWHERE args to existing args
	newArgs := append(prewhereArgs, args...)

	return newQuery, newArgs
}

// applySettings applies ClickHouse SETTINGS to the query.
func (t *botStorage) applySettings(query string, settings map[string]interface{}) string {
	if len(settings) == 0 {
		return query
	}

	var settingsParts []string
	for key, value := range settings {
		switch v := value.(type) {
		case string:
			settingsParts = append(settingsParts, fmt.Sprintf("%s = '%s'", key, v))
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			settingsParts = append(settingsParts, fmt.Sprintf("%s = %v", key, v))
		case float32, float64:
			settingsParts = append(settingsParts, fmt.Sprintf("%s = %v", key, v))
		case bool:
			var boolValue int
			if v {
				boolValue = 1
			}
			settingsParts = append(settingsParts, fmt.Sprintf("%s = %d", key, boolValue))
		default:
			settingsParts = append(settingsParts, fmt.Sprintf("%s = %v", key, v))
		}
	}

	if len(settingsParts) > 0 {
		return query + "\nSETTINGS\n\t" + strings.Join(settingsParts, ",\n\t")
	}

	return query
}

// TableName returns the table name.
func (t *botStorage) TableName() string {
	return "bots"
}

// GetQueryBuilder returns the query builder.
func (t *botStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *botStorage) Columns() []string {
	return []string{
		"id", "user_id", "name", "token", "is_publish", "created_at", "updated_at", "deleted_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *botStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *botStorage) SetConfig(config *Config) BotStorage {
	t.config = config
	return t
}

func (t *botStorage) SetQueryBuilder(builder sq.StatementBuilderType) BotStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *botStorage) LoadUser(ctx context.Context, model *Bot, builders ...*QueryBuilder) error {
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
func (t *botStorage) LoadBatchUser(ctx context.Context, items []*Bot, builders ...*QueryBuilder) error {
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

// Bot is a struct for the "bots" table.
type Bot struct {
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
func (t *Bot) TableName() string {
	return "bots"
}

// ScanRow scans a row into a Bot.
func (t *Bot) ScanRow(row driver.Row) error {
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

// BotFilters is a struct that holds filters for Bot.
type BotFilters struct {
	Id        *string
	UserId    *string
	CreatedAt *time.Time
}

// BotIdEq returns a condition that checks if the field equals the value.
func BotIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// BotUserIdEq returns a condition that checks if the field equals the value.
func BotUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// BotCreatedAtEq returns a condition that checks if the field equals the value.
func BotCreatedAtEq(value time.Time) FilterApplier {
	return EqualsCondition{Field: "created_at", Value: value}
}

// BotIdNotEq returns a condition that checks if the field equals the value.
func BotIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// BotUserIdNotEq returns a condition that checks if the field equals the value.
func BotUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// BotCreatedAtNotEq returns a condition that checks if the field equals the value.
func BotCreatedAtNotEq(value time.Time) FilterApplier {
	return NotEqualsCondition{Field: "created_at", Value: value}
}

// BotIdGT greaterThanCondition than condition.
func BotIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// BotUserIdGT greaterThanCondition than condition.
func BotUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// BotCreatedAtGT greaterThanCondition than condition.
func BotCreatedAtGT(value time.Time) FilterApplier {
	return GreaterThanCondition{Field: "created_at", Value: value}
}

// BotIdLT less than condition.
func BotIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// BotUserIdLT less than condition.
func BotUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// BotCreatedAtLT less than condition.
func BotCreatedAtLT(value time.Time) FilterApplier {
	return LessThanCondition{Field: "created_at", Value: value}
}

// BotIdGTE greater than or equal condition.
func BotIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// BotUserIdGTE greater than or equal condition.
func BotUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotCreatedAtGTE greater than or equal condition.
func BotCreatedAtGTE(value time.Time) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotIdLTE less than or equal condition.
func BotIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// BotUserIdLTE less than or equal condition.
func BotUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotCreatedAtLTE less than or equal condition.
func BotCreatedAtLTE(value time.Time) FilterApplier {
	return LessThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotIdBetween between condition.
func BotIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// BotUserIdBetween between condition.
func BotUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// BotCreatedAtBetween between condition.
func BotCreatedAtBetween(min, max time.Time) FilterApplier {
	return BetweenCondition{Field: "created_at", Min: min, Max: max}
}

// BotIdILike iLike condition %
func BotIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// BotUserIdILike iLike condition %
func BotUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// BotIdLike like condition %
func BotIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// BotUserIdLike like condition %
func BotUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// BotIdNotLike not like condition
func BotIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// BotUserIdNotLike not like condition
func BotUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// BotIdIn condition
func BotIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// BotUserIdIn condition
func BotUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// BotCreatedAtIn condition
func BotCreatedAtIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "created_at", Values: values}
}

// BotIdNotIn not in condition
func BotIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// BotUserIdNotIn not in condition
func BotUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// BotCreatedAtNotIn not in condition
func BotCreatedAtNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "created_at", Values: values}
}

// BotIdOrderBy sorts the result in ascending order.
func BotIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// BotUserIdOrderBy sorts the result in ascending order.
func BotUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// BotCreatedAtOrderBy sorts the result in ascending order.
func BotCreatedAtOrderBy(asc bool) FilterApplier {
	return OrderBy("created_at", asc)
}

// AsyncCreate asynchronously inserts a new Bot.
func (t *botStorage) AsyncCreate(ctx context.Context, model *Bot, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("bots").
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
		return fmt.Errorf("failed to asynchronously create Bot: %w", err)
	}

	return nil
}

// Create creates a new Bot.
func (t *botStorage) Create(ctx context.Context, model *Bot, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("bots").
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
		return fmt.Errorf("failed to create Bot: %w", err)
	}

	return nil
}

// BatchCreate creates multiple Bot records in a single batch.
func (t *botStorage) BatchCreate(ctx context.Context, models []*Bot, opts ...Option) error {
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

// OriginalBatchCreate creates multiple Bot records in a single batch.
func (t *botStorage) OriginalBatchCreate(ctx context.Context, models []*Bot, opts ...Option) error {
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

// FindMany finds multiple Bot based on the provided options.
func (t *botStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Bot, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

	// set default options
	options := &Options{}

	// collect settings from all builders
	allSettings := make(map[string]interface{})

	// collect PREWHERE conditions
	var prewhereConditions []FilterApplier

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply custom table name
		query = builder.ApplyCustomTableName(query)

		// collect PREWHERE conditions (ClickHouse specific)
		prewhereConditions = append(prewhereConditions, builder.prewhereOptions...)

		// apply filter options (WHERE)
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

		// collect settings
		for k, v := range builder.settings {
			allSettings[k] = v
		}
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	// apply ClickHouse PREWHERE if present
	if len(prewhereConditions) > 0 {
		sqlQuery, args = t.applyPrewhere(sqlQuery, args, prewhereConditions)
	}

	// apply ClickHouse SETTINGS if present
	if len(allSettings) > 0 {
		sqlQuery = t.applySettings(sqlQuery, allSettings)
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

	var results []*Bot
	for rows.Next() {
		model := &Bot{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan Bot: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Bot based on the provided options.
func (t *botStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Bot, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Bot: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *botStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *botStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *botStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *botStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *botStorage) Conn() driver.Conn {
	return t.DB()
}
