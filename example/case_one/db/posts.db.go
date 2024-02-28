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

// postStorage is a struct for the "posts" table.
type postStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type PostStorage interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
	Create(ctx context.Context, model *Post, opts ...Option) (*int32, error)
	Update(ctx context.Context, id int32, updateData *PostUpdate) error
	DeleteById(ctx context.Context, id int32, opts ...Option) error
	FindById(ctx context.Context, id int32, opts ...Option) (*Post, error)
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Post, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Post, *Paginator, error)
}

// NewPostStorage returns a new postStorage.
func NewPostStorage(db *sql.DB) PostStorage {
	return &postStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *postStorage) TableName() string {
	return "posts"
}

// Columns returns the columns for the table.
func (t *postStorage) Columns() []string {
	return []string{
		"id",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *postStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *postStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		-- Table: posts
		CREATE TABLE IF NOT EXISTS posts (
		id  SERIAL PRIMARY KEY);
		-- Other entities
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *postStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS posts;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *postStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE posts;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *postStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// Post is a struct for the "posts" table.
type Post struct {
	Id int32 `db:"id"`
}

// TableName returns the table name.
func (t *Post) TableName() string {
	return "posts"
}

// ScanRow scans a row into a Post.
func (t *Post) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id)
}

// ScanRows scans a single row into the Post.
func (t *Post) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
	)
}

// PostFilters is a struct that holds filters for Post.
type PostFilters struct {
	Id *int32
}

// PostIdEq returns a condition that checks if the field equals the value.
func PostIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// PostIdNotEq returns a condition that checks if the field equals the value.
func PostIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// PostIdGT greaterThanCondition than condition.
func PostIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// PostIdLT less than condition.
func PostIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// PostIdGTE greater than or equal condition.
func PostIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// PostIdLTE less than or equal condition.
func PostIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// PostIdLike like condition %
func PostIdLike(value int32) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// PostIdNotLike not like condition
func PostIdNotLike(value int32) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// PostIdIsNull is null condition
func PostIdIsNull() FilterApplier {
	return IsNullCondition{Field: "id"}
}

// PostIdIsNotNull is not null condition
func PostIdIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "id"}
}

// PostIdIn condition
func PostIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// PostIdNotIn not in condition
func PostIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// PostIdOrderBy sorts the result in ascending order.
func PostIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// Create creates a new Post.
func (t *postStorage) Create(ctx context.Context, model *Post, opts ...Option) (*int32, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("posts").
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

		return nil, fmt.Errorf("failed to create Post: %w", err)
	}

	return &id, nil
}

// PostUpdate is used to update an existing Post.
type PostUpdate struct {
}

// Update updates an existing Post based on non-nil fields.
func (t *postStorage) Update(ctx context.Context, id int32, updateData *PostUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("posts")

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Post: %w", err)
	}

	return nil
}

// DeleteById - deletes a Post by its id.
func (t *postStorage) DeleteById(ctx context.Context, id int32, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("posts").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Post: %w", err)
	}

	return nil
}

// FindById retrieves a Post by its id.
func (t *postStorage) FindById(ctx context.Context, id int32, opts ...Option) (*Post, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(PostIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one Post: ")
	}

	return model, nil
}

// FindMany finds multiple Post based on the provided options.
func (t *postStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Post, error) {
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
		return nil, fmt.Errorf("failed to find Post: %w", err)
	}
	defer rows.Close()

	var results []*Post
	for rows.Next() {
		model := &Post{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Post: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Post based on the provided options.
func (t *postStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error) {
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

// Count counts Post based on the provided options.
func (t *postStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count Post: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Post with pagination support.
func (t *postStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Post, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Post: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Post: %w", err)
	}

	return records, paginator, nil
}

// LockForUpdate lock locks the Post for the given ID.
func (t *postStorage) LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error {
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
	var model Post
	if err := model.ScanRow(row); err != nil {
		return fmt.Errorf("failed to scan Post: %w", err)
	}

	return nil
}
