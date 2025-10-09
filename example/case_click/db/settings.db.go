package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
)

// settingStorage is a struct for the "settings" table.
type settingStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// SettingCRUDOperations is an interface for managing the settings table.
type SettingCRUDOperations interface {
	Create(ctx context.Context, model *Setting, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Setting, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Setting, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Setting, opts ...Option) error
}

// SettingSearchOperations is an interface for searching the settings table.
type SettingSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Setting, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error)
}

type SettingSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) SettingStorage
	SetQueryBuilder(builder sq.StatementBuilderType) SettingStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// SettingRelationLoading is an interface for loading relations.
type SettingRelationLoading interface {
	LoadUser(ctx context.Context, model *Setting, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Setting, builders ...*QueryBuilder) error
}

// SettingRawQueryOperations is an interface for executing raw queries.
type SettingRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// SettingStorage is a struct for the "settings" table.
type SettingStorage interface {
	SettingCRUDOperations
	SettingSearchOperations
	SettingRelationLoading
	SettingRawQueryOperations
	SettingSettings
}

// NewSettingStorage returns a new settingStorage.
func NewSettingStorage(config *Config) (SettingStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &settingStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *settingStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *settingStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *settingStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
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
func (t *settingStorage) applySettings(query string, settings map[string]interface{}) string {
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
func (t *settingStorage) TableName() string {
	return "settings"
}

// GetQueryBuilder returns the query builder.
func (t *settingStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *settingStorage) Columns() []string {
	return []string{
		"id", "name", "value", "user_id",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *settingStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *settingStorage) SetConfig(config *Config) SettingStorage {
	t.config = config
	return t
}

func (t *settingStorage) SetQueryBuilder(builder sq.StatementBuilderType) SettingStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *settingStorage) LoadUser(ctx context.Context, model *Setting, builders ...*QueryBuilder) error {
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
func (t *settingStorage) LoadBatchUser(ctx context.Context, items []*Setting, builders ...*QueryBuilder) error {
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

// Setting is a struct for the "settings" table.
type Setting struct {
	Id     int32
	Name   string
	Value  string
	User   *User
	UserId string
}

// TableName returns the table name.
func (t *Setting) TableName() string {
	return "settings"
}

// ScanRow scans a row into a Setting.
func (t *Setting) ScanRow(row driver.Row) error {
	return row.Scan(
		&t.Id,
		&t.Name,
		&t.Value,
		&t.UserId,
	)
}

// SettingFilters is a struct that holds filters for Setting.
type SettingFilters struct {
	Id     *int32
	UserId *string
}

// SettingIdEq returns a condition that checks if the field equals the value.
func SettingIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// SettingUserIdEq returns a condition that checks if the field equals the value.
func SettingUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// SettingIdNotEq returns a condition that checks if the field equals the value.
func SettingIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// SettingUserIdNotEq returns a condition that checks if the field equals the value.
func SettingUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// SettingIdGT greaterThanCondition than condition.
func SettingIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// SettingUserIdGT greaterThanCondition than condition.
func SettingUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// SettingIdLT less than condition.
func SettingIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// SettingUserIdLT less than condition.
func SettingUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// SettingIdGTE greater than or equal condition.
func SettingIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// SettingUserIdGTE greater than or equal condition.
func SettingUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// SettingIdLTE less than or equal condition.
func SettingIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// SettingUserIdLTE less than or equal condition.
func SettingUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// SettingIdBetween between condition.
func SettingIdBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// SettingUserIdBetween between condition.
func SettingUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// SettingUserIdILike iLike condition %
func SettingUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// SettingUserIdLike like condition %
func SettingUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// SettingUserIdNotLike not like condition
func SettingUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// SettingIdIn condition
func SettingIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// SettingUserIdIn condition
func SettingUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// SettingIdNotIn not in condition
func SettingIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// SettingUserIdNotIn not in condition
func SettingUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// SettingIdOrderBy sorts the result in ascending order.
func SettingIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// SettingUserIdOrderBy sorts the result in ascending order.
func SettingUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// AsyncCreate asynchronously inserts a new Setting.
func (t *settingStorage) AsyncCreate(ctx context.Context, model *Setting, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("settings").
		Columns(
			"name",
			"value",
			"user_id",
		).
		Values(
			model.Name,
			model.Value,
			model.UserId,
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create Setting: %w", err)
	}

	return nil
}

// Create creates a new Setting.
func (t *settingStorage) Create(ctx context.Context, model *Setting, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("settings").
		Columns(
			"name",
			"value",
			"user_id",
		).
		Values(
			model.Name,
			model.Value,
			model.UserId,
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create Setting: %w", err)
	}

	return nil
}

// BatchCreate creates multiple Setting records in a single batch.
func (t *settingStorage) BatchCreate(ctx context.Context, models []*Setting, opts ...Option) error {
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
			model.Name,
			model.Value,
			model.UserId,
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

// OriginalBatchCreate creates multiple Setting records in a single batch.
func (t *settingStorage) OriginalBatchCreate(ctx context.Context, models []*Setting, opts ...Option) error {
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
			"name",
			"value",
			"user_id",
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}

		query = query.Values(
			model.Name,
			model.Value,
			model.UserId,
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

// FindMany finds multiple Setting based on the provided options.
func (t *settingStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Setting, error) {
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

	var results []*Setting
	for rows.Next() {
		model := &Setting{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan Setting: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Setting based on the provided options.
func (t *settingStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Setting: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *settingStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *settingStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *settingStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *settingStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *settingStorage) Conn() driver.Conn {
	return t.DB()
}
