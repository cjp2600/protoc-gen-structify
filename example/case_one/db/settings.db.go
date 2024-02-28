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

// settingStorage is a struct for the "settings" table.
type settingStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type SettingStorage interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
	Create(ctx context.Context, model *Setting, opts ...Option) (*int32, error)
	Update(ctx context.Context, id int32, updateData *SettingUpdate) error
	DeleteById(ctx context.Context, id int32, opts ...Option) error
	FindById(ctx context.Context, id int32, opts ...Option) (*Setting, error)
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Setting, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Setting, *Paginator, error)
}

// NewSettingStorage returns a new settingStorage.
func NewSettingStorage(db *sql.DB) SettingStorage {
	return &settingStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *settingStorage) TableName() string {
	return "settings"
}

// Columns returns the columns for the table.
func (t *settingStorage) Columns() []string {
	return []string{
		"id",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *settingStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *settingStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		-- Table: settings
		CREATE TABLE IF NOT EXISTS settings (
		id  SERIAL PRIMARY KEY);
		-- Other entities
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *settingStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS settings;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *settingStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE settings;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *settingStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// Setting is a struct for the "settings" table.
type Setting struct {
	Id int32 `db:"id"`
}

// TableName returns the table name.
func (t *Setting) TableName() string {
	return "settings"
}

// ScanRow scans a row into a Setting.
func (t *Setting) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id)
}

// ScanRows scans a single row into the Setting.
func (t *Setting) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
	)
}

// SettingFilters is a struct that holds filters for Setting.
type SettingFilters struct {
	Id *int32
}

// SettingIdEq returns a condition that checks if the field equals the value.
func SettingIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// SettingIdNotEq returns a condition that checks if the field equals the value.
func SettingIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// SettingIdGT greaterThanCondition than condition.
func SettingIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// SettingIdLT less than condition.
func SettingIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// SettingIdGTE greater than or equal condition.
func SettingIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// SettingIdLTE less than or equal condition.
func SettingIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// SettingIdLike like condition %
func SettingIdLike(value int32) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// SettingIdNotLike not like condition
func SettingIdNotLike(value int32) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// SettingIdIsNull is null condition
func SettingIdIsNull() FilterApplier {
	return IsNullCondition{Field: "id"}
}

// SettingIdIsNotNull is not null condition
func SettingIdIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "id"}
}

// SettingIdIn condition
func SettingIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// SettingIdNotIn not in condition
func SettingIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// SettingIdOrderBy sorts the result in ascending order.
func SettingIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// Create creates a new Setting.
func (t *settingStorage) Create(ctx context.Context, model *Setting, opts ...Option) (*int32, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("settings").
		Columns().
		Values()

	// add RETURNING "id" to query
	query = query.Suffix("RETURNING \"id\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var id int32
	err = t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, fmt.Errorf("failed to create Setting: %w", err)
	}

	return &id, nil
}

// SettingUpdate is used to update an existing Setting.
type SettingUpdate struct {
}

// Update updates an existing Setting based on non-nil fields.
func (t *settingStorage) Update(ctx context.Context, id int32, updateData *SettingUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("settings")

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Setting: %w", err)
	}

	return nil
}

// DeleteById - deletes a Setting by its id.
func (t *settingStorage) DeleteById(ctx context.Context, id int32, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("settings").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Setting: %w", err)
	}

	return nil
}

// FindById retrieves a Setting by its id.
func (t *settingStorage) FindById(ctx context.Context, id int32, opts ...Option) (*Setting, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(SettingIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one Setting: ")
	}

	return model, nil
}

// FindMany finds multiple Setting based on the provided options.
func (t *settingStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Setting, error) {
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
		return nil, fmt.Errorf("failed to find Setting: %w", err)
	}
	defer rows.Close()

	var results []*Setting
	for rows.Next() {
		model := &Setting{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Setting: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Setting based on the provided options.
func (t *settingStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error) {
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

// Count counts Setting based on the provided options.
func (t *settingStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count Setting: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Setting with pagination support.
func (t *settingStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Setting, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Setting: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Setting: %w", err)
	}

	return records, paginator, nil
}

// LockForUpdate lock locks the Setting for the given ID.
func (t *settingStorage) LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error {
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
		return fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var model Setting
	if err := model.ScanRow(row); err != nil {
		return fmt.Errorf("failed to scan Setting: %w", err)
	}

	return nil
}
