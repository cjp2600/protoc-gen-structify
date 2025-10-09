package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
	"time"
)

// addressStorage is a struct for the "addresses" table.
type addressStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// AddressCRUDOperations is an interface for managing the addresses table.
type AddressCRUDOperations interface {
	Create(ctx context.Context, model *Address, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Address, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Address, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Address, opts ...Option) error
}

// AddressSearchOperations is an interface for searching the addresses table.
type AddressSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Address, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error)
}

type AddressSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) AddressStorage
	SetQueryBuilder(builder sq.StatementBuilderType) AddressStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// AddressRelationLoading is an interface for loading relations.
type AddressRelationLoading interface {
	LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error
}

// AddressRawQueryOperations is an interface for executing raw queries.
type AddressRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// AddressStorage is a struct for the "addresses" table.
type AddressStorage interface {
	AddressCRUDOperations
	AddressSearchOperations
	AddressRelationLoading
	AddressRawQueryOperations
	AddressSettings
}

// NewAddressStorage returns a new addressStorage.
func NewAddressStorage(config *Config) (AddressStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &addressStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *addressStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *addressStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *addressStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
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
func (t *addressStorage) applySettings(query string, settings map[string]interface{}) string {
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
func (t *addressStorage) TableName() string {
	return "addresses"
}

// GetQueryBuilder returns the query builder.
func (t *addressStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *addressStorage) Columns() []string {
	return []string{
		"id", "street", "city", "state", "zip", "user_id", "created_at", "updated_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *addressStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *addressStorage) SetConfig(config *Config) AddressStorage {
	t.config = config
	return t
}

func (t *addressStorage) SetQueryBuilder(builder sq.StatementBuilderType) AddressStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *addressStorage) LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error {
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
func (t *addressStorage) LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error {
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

// Address is a struct for the "addresses" table.
type Address struct {
	Id        string
	Street    string
	City      string
	State     int32
	Zip       int64
	User      *User
	UserId    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// TableName returns the table name.
func (t *Address) TableName() string {
	return "addresses"
}

// ScanRow scans a row into a Address.
func (t *Address) ScanRow(row driver.Row) error {
	return row.Scan(
		&t.Id,
		&t.Street,
		&t.City,
		&t.State,
		&t.Zip,
		&t.UserId,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

// AddressFilters is a struct that holds filters for Address.
type AddressFilters struct {
	Id     *string
	UserId *string
}

// AddressIdEq returns a condition that checks if the field equals the value.
func AddressIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// AddressUserIdEq returns a condition that checks if the field equals the value.
func AddressUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// AddressIdNotEq returns a condition that checks if the field equals the value.
func AddressIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// AddressUserIdNotEq returns a condition that checks if the field equals the value.
func AddressUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// AddressIdGT greaterThanCondition than condition.
func AddressIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// AddressUserIdGT greaterThanCondition than condition.
func AddressUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// AddressIdLT less than condition.
func AddressIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// AddressUserIdLT less than condition.
func AddressUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// AddressIdGTE greater than or equal condition.
func AddressIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// AddressUserIdGTE greater than or equal condition.
func AddressUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// AddressIdLTE less than or equal condition.
func AddressIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// AddressUserIdLTE less than or equal condition.
func AddressUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// AddressIdBetween between condition.
func AddressIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// AddressUserIdBetween between condition.
func AddressUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// AddressIdILike iLike condition %
func AddressIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// AddressUserIdILike iLike condition %
func AddressUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// AddressIdLike like condition %
func AddressIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// AddressUserIdLike like condition %
func AddressUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// AddressIdNotLike not like condition
func AddressIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// AddressUserIdNotLike not like condition
func AddressUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// AddressIdIn condition
func AddressIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// AddressUserIdIn condition
func AddressUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// AddressIdNotIn not in condition
func AddressIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// AddressUserIdNotIn not in condition
func AddressUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// AddressIdOrderBy sorts the result in ascending order.
func AddressIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// AddressUserIdOrderBy sorts the result in ascending order.
func AddressUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// AsyncCreate asynchronously inserts a new Address.
func (t *addressStorage) AsyncCreate(ctx context.Context, model *Address, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("addresses").
		Columns(
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create Address: %w", err)
	}

	return nil
}

// Create creates a new Address.
func (t *addressStorage) Create(ctx context.Context, model *Address, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("addresses").
		Columns(
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create Address: %w", err)
	}

	return nil
}

// BatchCreate creates multiple Address records in a single batch.
func (t *addressStorage) BatchCreate(ctx context.Context, models []*Address, opts ...Option) error {
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
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

// OriginalBatchCreate creates multiple Address records in a single batch.
func (t *addressStorage) OriginalBatchCreate(ctx context.Context, models []*Address, opts ...Option) error {
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
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}

		query = query.Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

// FindMany finds multiple Address based on the provided options.
func (t *addressStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Address, error) {
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

	var results []*Address
	for rows.Next() {
		model := &Address{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan Address: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Address based on the provided options.
func (t *addressStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Address: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *addressStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *addressStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *addressStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *addressStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *addressStorage) Conn() driver.Conn {
	return t.DB()
}
