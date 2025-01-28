package db

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math"
)

// deviceStorage is a struct for the "devices" table.
type deviceStorage struct {
	db           *DB                     // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

// DeviceCRUDOperations is an interface for managing the devices table.
type DeviceCRUDOperations interface {
	Create(ctx context.Context, model *Device, opts ...Option) error

	BatchCreate(ctx context.Context, models []*Device, opts ...Option) error
	Update(ctx context.Context, id int64, updateData *DeviceUpdate) error
}

// DeviceSearchOperations is an interface for searching the devices table.
type DeviceSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Device, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Device, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Device, error)
}

// DevicePaginationOperations is an interface for pagination operations.
type DevicePaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Device, *Paginator, error)
}

// DeviceRelationLoading is an interface for loading relations.
type DeviceRelationLoading interface {
}

// DeviceAdvancedDeletion is an interface for advanced deletion operations.
type DeviceAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// DeviceRawQueryOperations is an interface for executing raw queries.
type DeviceRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// DeviceStorage is a struct for the "devices" table.
type DeviceStorage interface {
	DeviceCRUDOperations
	DeviceSearchOperations
	DevicePaginationOperations
	DeviceRelationLoading
	DeviceAdvancedDeletion
	DeviceRawQueryOperations
}

// NewDeviceStorage returns a new deviceStorage.
func NewDeviceStorage(db *DB) DeviceStorage {
	return &deviceStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *deviceStorage) TableName() string {
	return "devices"
}

// Columns returns the columns for the table.
func (t *deviceStorage) Columns() []string {
	return []string{
		"name", "value", "user_id",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *deviceStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	var db QueryExecer

	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		db = t.db.DBWrite
	} else {
		db = t.db.DBRead
	}

	return db
}

// Device is a struct for the "devices" table.
type Device struct {
	Name   string `db:"name"`
	Value  string `db:"value"`
	UserId string `db:"user_id"`
}

// TableName returns the table name.
func (t *Device) TableName() string {
	return "devices"
}

// ScanRow scans a row into a Device.
func (t *Device) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Name, &t.Value, &t.UserId)
}

// ScanRows scans a single row into the Device.
func (t *Device) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Name,
		&t.Value,
		&t.UserId,
	)
}

// DeviceFilters is a struct that holds filters for Device.
type DeviceFilters struct {
	UserId *string
}

// DeviceUserIdEq returns a condition that checks if the field equals the value.
func DeviceUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// DeviceUserIdNotEq returns a condition that checks if the field equals the value.
func DeviceUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// DeviceUserIdGT greaterThanCondition than condition.
func DeviceUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// DeviceUserIdLT less than condition.
func DeviceUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// DeviceUserIdGTE greater than or equal condition.
func DeviceUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// DeviceUserIdLTE less than or equal condition.
func DeviceUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// DeviceUserIdBetween between condition.
func DeviceUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// DeviceUserIdILike iLike condition %
func DeviceUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// DeviceUserIdLike like condition %
func DeviceUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// DeviceUserIdNotLike not like condition
func DeviceUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// DeviceUserIdIn condition
func DeviceUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// DeviceUserIdNotIn not in condition
func DeviceUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// DeviceUserIdOrderBy sorts the result in ascending order.
func DeviceUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// Create creates a new Device.
func (t *deviceStorage) Create(ctx context.Context, model *Device, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("devices").
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
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return errors.Wrap(err, "failed to create Device")
	}

	return nil
}

// BatchCreate creates multiple Device records in a single batch.
func (t *deviceStorage) BatchCreate(ctx context.Context, models []*Device, opts ...Option) error {
	if len(models) == 0 {
		return errors.New("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return errors.New("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"name",
			"value",
			"user_id",
		)

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}
		query = query.Values(
			model.Name,
			model.Value,
			model.UserId,
		)
	}

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return errors.Wrap(err, "failed to execute bulk insert")
	}
	defer rows.Close()

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "rows iteration error")
	}

	return nil
}

// DeviceUpdate is used to update an existing Device.
type DeviceUpdate struct {
	// Use regular pointer types for non-optional fields
	Name *string
	// Use regular pointer types for non-optional fields
	Value *string
	// Use regular pointer types for non-optional fields
	UserId *string
}

// Update updates an existing Device based on non-nil fields.
func (t *deviceStorage) Update(ctx context.Context, id int64, updateData *DeviceUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("devices")
	// Handle fields that are not optional using a nil check
	if updateData.Name != nil {
		query = query.Set("name", *updateData.Name) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Value != nil {
		query = query.Set("value", *updateData.Value) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.UserId != nil {
		query = query.Set("user_id", *updateData.UserId) // Dereference pointer value
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update Device")
	}

	return nil
}

// DeleteMany removes entries from the devices table using the provided filters
func (t *deviceStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("devices")

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
		return errors.New("filters are required for delete operation")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete devices")
	}

	return nil
}

// FindMany finds multiple Device based on the provided options.
func (t *deviceStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Device, error) {
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
		return nil, errors.Wrap(err, "failed to build query")
	}

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var results []*Device
	for rows.Next() {
		model := &Device{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan Device")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single Device based on the provided options.
func (t *deviceStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Device, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne Device")
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Count counts Device based on the provided options.
func (t *deviceStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, errors.Wrap(err, "failed to build query")
	}

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to scan count")
	}

	return count, nil
}

// FindManyWithPagination finds multiple Device with pagination support.
func (t *deviceStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Device, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count Device")
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
		return nil, nil, errors.Wrap(err, "failed to find Device")
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the Device for the given ID.
func (t *deviceStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Device, error) {
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
		return nil, errors.Wrap(err, "failed to build query")
	}

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model Device
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, errors.Wrap(err, "failed to scan Device")
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *deviceStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *deviceStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *deviceStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
