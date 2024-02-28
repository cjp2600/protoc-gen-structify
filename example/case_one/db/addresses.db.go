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

// addressStorage is a struct for the "addresses" table.
type addressStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type AddressStorage interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
	Create(ctx context.Context, model *Address, opts ...Option) (*string, error)
	Update(ctx context.Context, id string, updateData *AddressUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*Address, error)
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Address, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Address, *Paginator, error)
}

// NewAddressStorage returns a new addressStorage.
func NewAddressStorage(db *sql.DB) AddressStorage {
	return &addressStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *addressStorage) TableName() string {
	return "addresses"
}

// Columns returns the columns for the table.
func (t *addressStorage) Columns() []string {
	return []string{
		"id",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *addressStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *addressStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		-- Table: addresses
		CREATE TABLE IF NOT EXISTS addresses (
		id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4());
		-- Other entities
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *addressStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS addresses;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *addressStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE addresses;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *addressStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// Address is a struct for the "addresses" table.
type Address struct {
	Id string `db:"id"`
}

// TableName returns the table name.
func (t *Address) TableName() string {
	return "addresses"
}

// ScanRow scans a row into a Address.
func (t *Address) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id)
}

// ScanRows scans a single row into the Address.
func (t *Address) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
	)
}

// AddressFilters is a struct that holds filters for Address.
type AddressFilters struct {
	Id *string

	queries []FilterApplier
}

func NewFilterAddress() *AddressFilters {
	return &AddressFilters{}
}

func (f *AddressFilters) IdEq(v string) *AddressFilters {
	f.queries = append(f.queries, AddressIdEq(v))
	return f
}

func (f *AddressFilters) IdNotEq(v string) *AddressFilters {
	f.queries = append(f.queries, AddressIdNotEq(v))
	return f
}

func (f *AddressFilters) IdGT(v string) *AddressFilters {
	f.queries = append(f.queries, AddressIdGT(v))
	return f
}

func (f *AddressFilters) IdLT(v string) *AddressFilters {
	f.queries = append(f.queries, AddressIdLT(v))
	return f
}

func (f *AddressFilters) IdGTE(v string) *AddressFilters {
	f.queries = append(f.queries, AddressIdGTE(v))
	return f
}

// AddressIdEq returns a condition that checks if the field equals the value.
func AddressIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// AddressIdNotEq returns a condition that checks if the field equals the value.
func AddressIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// AddressIdGT greaterThanCondition than condition.
func AddressIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// AddressIdLT less than condition.
func AddressIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// AddressIdGTE greater than or equal condition.
func AddressIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// AddressIdLTE less than or equal condition.
func AddressIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// AddressIdLike like condition %
func AddressIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// AddressIdNotLike not like condition
func AddressIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// AddressIdIsNull is null condition
func AddressIdIsNull() FilterApplier {
	return IsNullCondition{Field: "id"}
}

// AddressIdIsNotNull is not null condition
func AddressIdIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "id"}
}

// AddressIdIn condition
func AddressIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// AddressIdNotIn not in condition
func AddressIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// AddressIdOrderBy sorts the result in ascending order.
func AddressIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// Create creates a new Address.
func (t *addressStorage) Create(ctx context.Context, model *Address, opts ...Option) (*string, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("addresses").
		Columns().
		Values()

	// add RETURNING "id" to query
	query = query.Suffix("RETURNING \"id\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	var id string
	err = t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, fmt.Errorf("failed to create Address: %w", err)
	}

	return &id, nil
}

// AddressUpdate is used to update an existing Address.
type AddressUpdate struct {
}

// Update updates an existing Address based on non-nil fields.
func (t *addressStorage) Update(ctx context.Context, id string, updateData *AddressUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("addresses")

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Address: %w", err)
	}

	return nil
}

// DeleteById - deletes a Address by its id.
func (t *addressStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("addresses").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Address: %w", err)
	}

	return nil
}

// FindById retrieves a Address by its id.
func (t *addressStorage) FindById(ctx context.Context, id string, opts ...Option) (*Address, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(AddressIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one Address: ")
	}

	return model, nil
}

// FindMany finds multiple Address based on the provided options.
func (t *addressStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Address, error) {
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
		return nil, fmt.Errorf("failed to find Address: %w", err)
	}
	defer rows.Close()

	var results []*Address
	for rows.Next() {
		model := &Address{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Address: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Address based on the provided options.
func (t *addressStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error) {
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

// Count counts Address based on the provided options.
func (t *addressStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count Address: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Address with pagination support.
func (t *addressStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Address, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Address: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Address: %w", err)
	}

	return records, paginator, nil
}

// LockForUpdate lock locks the Address for the given ID.
func (t *addressStorage) LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error {
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
	var model Address
	if err := model.ScanRow(row); err != nil {
		return fmt.Errorf("failed to scan Address: %w", err)
	}

	return nil
}
