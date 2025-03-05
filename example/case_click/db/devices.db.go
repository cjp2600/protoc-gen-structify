package db

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

// deviceStorage is a struct for the "devices" table.
type deviceStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// DeviceCRUDOperations is an interface for managing the devices table.
type DeviceCRUDOperations interface {
	Create(ctx context.Context, model *Device, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Device, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Device, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Device, opts ...Option) error
}

// DeviceSearchOperations is an interface for searching the devices table.
type DeviceSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Device, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Device, error)
}

type DeviceSettings interface {
	Conn() driver.Conn
	SetConfig(config *Config) DeviceStorage
	SetQueryBuilder(builder sq.StatementBuilderType) DeviceStorage
}

// DeviceRelationLoading is an interface for loading relations.
type DeviceRelationLoading interface {
}

// DeviceRawQueryOperations is an interface for executing raw queries.
type DeviceRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// DeviceStorage is a struct for the "devices" table.
type DeviceStorage interface {
	DeviceCRUDOperations
	DeviceSearchOperations
	DeviceRelationLoading
	DeviceRawQueryOperations
	DeviceSettings
}

// NewDeviceStorage returns a new deviceStorage.
func NewDeviceStorage(config *Config) (DeviceStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB connection is nil")
	}

	return &deviceStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *deviceStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *deviceStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
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
func (t *deviceStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *deviceStorage) SetConfig(config *Config) DeviceStorage {
	t.config = config
	return t
}

func (t *deviceStorage) SetQueryBuilder(builder sq.StatementBuilderType) DeviceStorage {
	t.queryBuilder = builder
	return t
}

// Device is a struct for the "devices" table.
type Device struct {
	Name   string
	Value  string
	UserId string
}

// TableName returns the table name.
func (t *Device) TableName() string {
	return "devices"
}

// ScanRow scans a row into a Device.
func (t *Device) ScanRow(row driver.Row) error {
	return row.Scan(
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

// AsyncCreate asynchronously inserts a new Device.
func (t *deviceStorage) AsyncCreate(ctx context.Context, model *Device, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
	}

	// Set default options
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
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, false, args...); err != nil {
		return errors.Wrap(err, "failed to asynchronously create Device")
	}

	return nil
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
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
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

	batch, err := t.DB().PrepareBatch(ctx, "INSERT INTO "+t.TableName(), driver.WithReleaseConnection())
	if err != nil {
		return errors.Wrap(err, "failed to prepare batch")
	}

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}

		err := batch.Append(
			model.Name,
			model.Value,
			model.UserId,
		)
		if err != nil {
			return errors.Wrap(err, "failed to append to batch")
		}
	}

	if err := batch.Send(); err != nil {
		return errors.Wrap(err, "failed to execute batch insert")
	}

	return nil
}

// OriginalBatchCreate creates multiple Device records in a single batch.
func (t *deviceStorage) OriginalBatchCreate(ctx context.Context, models []*Device, opts ...Option) error {
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().Query(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to execute bulk insert")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	if err := rows.Err(); err != nil {
		return errors.Wrap(err, "rows iteration error")
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
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB().Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*Device
	for rows.Next() {
		model := &Device{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
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

// Select executes a raw query and returns the result.
func (t *deviceStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *deviceStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *deviceStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *deviceStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *deviceStorage) Conn() driver.Conn {
	return t.DB()
}
