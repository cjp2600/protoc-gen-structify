package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math"
)

// deviceStorage is a struct for the "devices" table.
type deviceStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type DeviceStorage interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
	Create(ctx context.Context, model *Device, opts ...Option) error
	Update(ctx context.Context, id int64, updateData *DeviceUpdate) error
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Device, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Device, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Device, error)
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Device, *Paginator, error)
}

// NewDeviceStorage returns a new deviceStorage.
func NewDeviceStorage(db *sql.DB) DeviceStorage {
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

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *deviceStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *deviceStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		-- Table: devices
		CREATE TABLE IF NOT EXISTS devices (
		name TEXT,
		value TEXT,
		user_id UUID NOT NULL);
		-- Other entities
		CREATE UNIQUE INDEX IF NOT EXISTS devices_user_id_unique_idx ON devices USING btree (user_id);
		CREATE INDEX IF NOT EXISTS devices_user_id_idx ON devices USING btree (user_id);
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *deviceStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS devices;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *deviceStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE devices;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *deviceStorage) UpgradeTable(ctx context.Context) error {
	return nil
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
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return fmt.Errorf("failed to create Device: %w", err)
	}

	return nil
}

// DeviceUpdate is used to update an existing Device.
type DeviceUpdate struct {
	Name   *string
	Value  *string
	UserId *string
}

// Update updates an existing Device based on non-nil fields.
func (t *deviceStorage) Update(ctx context.Context, id int64, updateData *DeviceUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("devices")
	if updateData.Name != nil {
		query = query.Set("name", updateData.Name)
	}
	if updateData.Value != nil {
		query = query.Set("value", updateData.Value)
	}
	if updateData.UserId != nil {
		query = query.Set("user_id", updateData.UserId)
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Device: %w", err)
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

	rows, err := t.DB(ctx).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find Device: %w", err)
	}
	defer rows.Close()

	var results []*Device
	for rows.Next() {
		model := &Device{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Device: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Device based on the provided options.
func (t *deviceStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Device, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, err
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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count Device: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Device with pagination support.
func (t *deviceStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Device, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Device: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Device: %w", err)
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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var model Device
	if err := model.ScanRow(row); err != nil {
		return nil, fmt.Errorf("failed to scan Device: %w", err)
	}

	return &model, nil
}
