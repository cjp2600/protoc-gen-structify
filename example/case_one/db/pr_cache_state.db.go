package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"math"
	"strings"
	"time"
)

// prCacheStateStorage is a struct for the "pr_cache_state" table.
type prCacheStateStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// PrCacheStateCRUDOperations is an interface for managing the pr_cache_state table.
type PrCacheStateCRUDOperations interface {
	Create(ctx context.Context, model *PrCacheState, opts ...Option) (*string, error)
	Upsert(ctx context.Context, model *PrCacheState, updateFields []string, opts ...Option) (*string, error)
	BatchCreate(ctx context.Context, models []*PrCacheState, opts ...Option) ([]string, error)
	Update(ctx context.Context, id string, updateData *PrCacheStateUpdate) error
	DeleteByCustomerId(ctx context.Context, customerId string, opts ...Option) error
	FindByCustomerId(ctx context.Context, id string, opts ...Option) (*PrCacheState, error)
	GetCustomerIdField(ctx context.Context, customerId string, field string) (interface{}, error)
}

// PrCacheStateSearchOperations is an interface for searching the pr_cache_state table.
type PrCacheStateSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*PrCacheState, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*PrCacheState, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*PrCacheState, error)
}

// PrCacheStatePaginationOperations is an interface for pagination operations.
type PrCacheStatePaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*PrCacheState, *Paginator, error)
}

// PrCacheStateRelationLoading is an interface for loading relations.
type PrCacheStateRelationLoading interface {
}

// PrCacheStateAdvancedDeletion is an interface for advanced deletion operations.
type PrCacheStateAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// PrCacheStateRawQueryOperations is an interface for executing raw queries.
type PrCacheStateRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// PrCacheStateStorage is a struct for the "pr_cache_state" table.
type PrCacheStateStorage interface {
	PrCacheStateCRUDOperations
	PrCacheStateSearchOperations
	PrCacheStatePaginationOperations
	PrCacheStateRelationLoading
	PrCacheStateAdvancedDeletion
	PrCacheStateRawQueryOperations
}

// NewPrCacheStateStorage returns a new prCacheStateStorage.
func NewPrCacheStateStorage(config *Config) (PrCacheStateStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB is nil")
	}
	if config.DB.DBRead == nil {
		return nil, fmt.Errorf("config.DB.DBRead is nil")
	}
	if config.DB.DBWrite == nil {
		config.DB.DBWrite = config.DB.DBRead
	}

	return &prCacheStateStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *prCacheStateStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *prCacheStateStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// TableName returns the table name.
func (t *prCacheStateStorage) TableName() string {
	return "pr_cache_state"
}

// Columns returns the columns for the table.
func (t *prCacheStateStorage) Columns() []string {
	return []string{
		"customer_id", "created_at", "last_access_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *prCacheStateStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		if tx == nil {
			t.logError(ctx, fmt.Errorf("transaction is nil"), "failed to get transaction from context")
			// set default connection
			return &dbWrapper{db: t.config.DB.DBWrite}
		}

		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		return &dbWrapper{db: t.config.DB.DBWrite}
	} else {
		return &dbWrapper{db: t.config.DB.DBRead}
	}
}

// PrCacheState is a struct for the "pr_cache_state" table.
type PrCacheState struct {
	CustomerId   string    `db:"customer_id"`
	CreatedAt    time.Time `db:"created_at"`
	LastAccessAt time.Time `db:"last_access_at"`
}

// TableName returns the table name.
func (t *PrCacheState) TableName() string {
	return "pr_cache_state"
}

// ScanRow scans a row into a PrCacheState.
func (t *PrCacheState) ScanRow(r *sql.Row) error {
	return r.Scan(&t.CustomerId, &t.CreatedAt, &t.LastAccessAt)
}

// ScanRows scans a single row into the PrCacheState.
func (t *PrCacheState) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.CustomerId,
		&t.CreatedAt,
		&t.LastAccessAt,
	)
}

// PrCacheStateFilters is a struct that holds filters for PrCacheState.
type PrCacheStateFilters struct {
	CustomerId *string
}

// PrCacheStateCustomerIdEq returns a condition that checks if the field equals the value.
func PrCacheStateCustomerIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdNotEq returns a condition that checks if the field equals the value.
func PrCacheStateCustomerIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdGT greaterThanCondition than condition.
func PrCacheStateCustomerIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdLT less than condition.
func PrCacheStateCustomerIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdGTE greater than or equal condition.
func PrCacheStateCustomerIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdLTE less than or equal condition.
func PrCacheStateCustomerIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdBetween between condition.
func PrCacheStateCustomerIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "customer_id", Min: min, Max: max}
}

// PrCacheStateCustomerIdILike iLike condition %
func PrCacheStateCustomerIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdLike like condition %
func PrCacheStateCustomerIdLike(value string) FilterApplier {
	return LikeCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdNotLike not like condition
func PrCacheStateCustomerIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "customer_id", Value: value}
}

// PrCacheStateCustomerIdIn condition
func PrCacheStateCustomerIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "customer_id", Values: values}
}

// PrCacheStateCustomerIdNotIn not in condition
func PrCacheStateCustomerIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "customer_id", Values: values}
}

// PrCacheStateCustomerIdOrderBy sorts the result in ascending order.
func PrCacheStateCustomerIdOrderBy(asc bool) FilterApplier {
	return OrderBy("customer_id", asc)
}

// Create creates a new PrCacheState.
func (t *prCacheStateStorage) Create(ctx context.Context, model *PrCacheState, opts ...Option) (*string, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("pr_cache_state").
		Columns(
			"customer_id",
			"created_at",
			"last_access_at",
		).
		Values(
			model.CustomerId,
			model.CreatedAt,
			model.LastAccessAt,
		)

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING RETURNING \"customer_id\"")
	} else {
		query = query.Suffix("RETURNING \"customer_id\"")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, fmt.Errorf("failed to create PrCacheState: %w", err)
	}

	return &id, nil
}

// Upsert creates a new PrCacheState or updates existing one on conflict.
func (t *prCacheStateStorage) Upsert(ctx context.Context, model *PrCacheState, updateFields []string, opts ...Option) (*string, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	// Build INSERT query
	query := t.queryBuilder.Insert("pr_cache_state").
		Columns(
			"customer_id",
			"created_at",
			"last_access_at",
		).
		Values(
			model.CustomerId,
			model.CreatedAt,
			model.LastAccessAt,
		)

	// Build UPDATE SET clause based on updateFields
	updateSet := make([]string, 0, len(updateFields))
	for _, field := range updateFields {
		if field == "created_at" {
			updateSet = append(updateSet, "created_at = EXCLUDED.created_at")
		}
		if field == "last_access_at" {
			updateSet = append(updateSet, "last_access_at = EXCLUDED.last_access_at")
		}
	}

	// Note: You can manually add updated_at to updateFields if needed

	// Build the complete suffix with ON CONFLICT, UPDATE SET, and RETURNING in one string
	var suffixBuilder strings.Builder

	// Add ON CONFLICT clause
	if options.ignoreConflictField != "" {
		suffixBuilder.WriteString("ON CONFLICT (")
		suffixBuilder.WriteString(options.ignoreConflictField)
		suffixBuilder.WriteString(") DO UPDATE SET ")
	} else {
		suffixBuilder.WriteString("ON CONFLICT (customer_id) DO UPDATE SET ")
	}

	// Add UPDATE SET fields
	if len(updateSet) > 0 {
		suffixBuilder.WriteString(strings.Join(updateSet, ", "))
	}

	// Add RETURNING clause
	suffixBuilder.WriteString(" RETURNING \"customer_id\"")

	// Add the complete suffix once
	query = query.Suffix(suffixBuilder.String())

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert PrCacheState: %w", err)
	}

	return &id, nil
}

// BatchCreate creates multiple PrCacheState records in a single batch.
func (t *prCacheStateStorage) BatchCreate(ctx context.Context, models []*PrCacheState, opts ...Option) ([]string, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return nil, fmt.Errorf("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"customer_id",
			"created_at",
			"last_access_at",
		)

	for _, model := range models {
		if model == nil {
			return nil, fmt.Errorf("one of the models is nil")
		}

		query = query.Values(
			model.CustomerId,
			model.CreatedAt,
			model.LastAccessAt,
		)
	}

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING RETURNING \"customer_id\"")
	} else {
		query = query.Suffix("RETURNING \"customer_id\"")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return nil, fmt.Errorf("failed to execute bulk insert: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var returnIDs []string
	for rows.Next() {
		var customer_id string
		if err := rows.Scan(&customer_id); err != nil {
			return nil, fmt.Errorf("failed to scan customer_id: %w", err)
		}
		returnIDs = append(returnIDs, customer_id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return returnIDs, nil
}

// PrCacheStateUpdate is used to update an existing PrCacheState.
type PrCacheStateUpdate struct {
	// Use regular pointer types for non-optional fields
	CreatedAt *time.Time
	// Use regular pointer types for non-optional fields
	LastAccessAt *time.Time
}

// Update updates an existing PrCacheState based on non-nil fields.
func (t *prCacheStateStorage) Update(ctx context.Context, id string, updateData *PrCacheStateUpdate) error {
	if updateData == nil {
		return fmt.Errorf("update data is nil")
	}

	query := t.queryBuilder.Update("pr_cache_state")
	// Handle fields that are not optional using a nil check
	if updateData.CreatedAt != nil {
		query = query.Set("created_at", *updateData.CreatedAt) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.LastAccessAt != nil {
		query = query.Set("last_access_at", *updateData.LastAccessAt) // Dereference pointer value
	}

	query = query.Where("customer_id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update PrCacheState: %w", err)
	}

	return nil
}

// DeleteByCustomerId - deletes a PrCacheState by its customer_id.
func (t *prCacheStateStorage) DeleteByCustomerId(ctx context.Context, customerId string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("pr_cache_state").Where("customer_id = ?", customerId)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete PrCacheState: %w", err)
	}

	return nil
}

// DeleteMany removes entries from the pr_cache_state table using the provided filters
func (t *prCacheStateStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("pr_cache_state")

	var withFilter bool
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.ApplyDelete(query)
			withFilter = true
		}
	}

	if !withFilter {
		return fmt.Errorf("filters are required for delete operation")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete pr_cache_state: %w", err)
	}

	return nil
}

// FindByCustomerId retrieves a PrCacheState by its customer_id.
func (t *prCacheStateStorage) FindByCustomerId(ctx context.Context, id string, opts ...Option) (*PrCacheState, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(PrCacheStateCustomerIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, fmt.Errorf("find one PrCacheState: %w", err)
	}

	return model, nil
}

// GetCustomerIdField retrieves a specific field value by customer_id.
func (t *prCacheStateStorage) GetCustomerIdField(ctx context.Context, customerId string, field string) (interface{}, error) {
	query := t.queryBuilder.Select(field).From(t.TableName()).Where("customer_id = ?", customerId)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var value interface{}
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan field value: %w", err)
	}

	return value, nil
}

// FindMany finds multiple PrCacheState based on the provided options.
func (t *prCacheStateStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*PrCacheState, error) {
	// build query
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName())

	// set default options
	options := &Options{}

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

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

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*PrCacheState
	for rows.Next() {
		model := &PrCacheState{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan PrCacheState: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single PrCacheState based on the provided options.
func (t *prCacheStateStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*PrCacheState, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne PrCacheState: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Count counts PrCacheState based on the provided options.
func (t *prCacheStateStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
	// build query
	query := t.queryBuilder.Select("COUNT(*)").From(t.TableName())

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple PrCacheState with pagination support.
func (t *prCacheStateStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*PrCacheState, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count PrCacheState: %w", err)
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Build the pagination object
	paginator := &Paginator{
		TotalCount: totalCount,
		Limit:      limit,
		Page:       page,
		TotalPages: int(math.Ceil(float64(totalCount) / float64(limit))),
	}

	// Add pagination to query builder
	builders = append(builders, PaginateBuilder(uint64(limit), uint64(offset)))

	// Find records using FindMany
	records, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find PrCacheState: %w", err)
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the PrCacheState for the given ID.
func (t *prCacheStateStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*PrCacheState, error) {
	query := t.queryBuilder.Select(t.Columns()...).From(t.TableName()).Suffix("FOR UPDATE")

	// apply options from builder
	for _, builder := range builders {
		if builder == nil {
			continue
		}

		// apply filter options
		for _, option := range builder.filterOptions {
			query = option.Apply(query)
		}

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model PrCacheState
	if err := model.ScanRow(row); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan PrCacheState: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *prCacheStateStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *prCacheStateStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *prCacheStateStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
