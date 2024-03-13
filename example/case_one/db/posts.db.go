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
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Post, error)
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Post, *Paginator, error)
	LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error
	LoadBatchAuthor(ctx context.Context, items []*Post, builders ...*QueryBuilder) error
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
		"id", "title", "body", "author_id",
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
		id  SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		body TEXT,
		author_id UUID NOT NULL);
		-- Other entities
		CREATE UNIQUE INDEX IF NOT EXISTS posts_author_id_unique_idx ON posts USING btree (author_id);
		CREATE INDEX IF NOT EXISTS posts_title_idx ON posts USING btree (title);
		CREATE INDEX IF NOT EXISTS posts_author_id_idx ON posts USING btree (author_id);
		-- Foreign keys for users
		ALTER TABLE posts
		ADD FOREIGN KEY (author_id) REFERENCES users(id);
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

// LoadAuthor loads the Author relation.
func (t *postStorage) LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Post is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s := NewUserStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(UserIdEq(model.AuthorId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find UserStorage: %w", err)
	}

	model.Author = relationModel
	return nil
}

// LoadBatchAuthor loads the Author relation.
func (t *postStorage) LoadBatchAuthor(ctx context.Context, items []*Post, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.AuthorId
	}

	// NewUserStorage creates a new UserStorage.
	s := NewUserStorage(t.db)

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
		if v, ok := resultMap[item.AuthorId]; ok {
			item.Author = v
		}
	}

	return nil
}

// Post is a struct for the "posts" table.
type Post struct {
	Id       int32  `db:"id"`
	Title    string `db:"title"`
	Body     string `db:"body"`
	Author   *User
	AuthorId string `db:"author_id"`
}

// TableName returns the table name.
func (t *Post) TableName() string {
	return "posts"
}

// ScanRow scans a row into a Post.
func (t *Post) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.Title, &t.Body, &t.AuthorId)
}

// ScanRows scans a single row into the Post.
func (t *Post) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.Title,
		&t.Body,
		&t.AuthorId,
	)
}

// PostFilters is a struct that holds filters for Post.
type PostFilters struct {
	Id       *int32
	AuthorId *string
}

// PostIdEq returns a condition that checks if the field equals the value.
func PostIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// PostAuthorIdEq returns a condition that checks if the field equals the value.
func PostAuthorIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "author_id", Value: value}
}

// PostIdNotEq returns a condition that checks if the field equals the value.
func PostIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// PostAuthorIdNotEq returns a condition that checks if the field equals the value.
func PostAuthorIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "author_id", Value: value}
}

// PostIdGT greaterThanCondition than condition.
func PostIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// PostAuthorIdGT greaterThanCondition than condition.
func PostAuthorIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "author_id", Value: value}
}

// PostIdLT less than condition.
func PostIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// PostAuthorIdLT less than condition.
func PostAuthorIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "author_id", Value: value}
}

// PostIdGTE greater than or equal condition.
func PostIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// PostAuthorIdGTE greater than or equal condition.
func PostAuthorIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "author_id", Value: value}
}

// PostIdLTE less than or equal condition.
func PostIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// PostAuthorIdLTE less than or equal condition.
func PostAuthorIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "author_id", Value: value}
}

// PostIdBetween between condition.
func PostIdBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// PostAuthorIdBetween between condition.
func PostAuthorIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "author_id", Min: min, Max: max}
}

// PostAuthorIdLike like condition %
func PostAuthorIdLike(value string) FilterApplier {
	return LikeCondition{Field: "author_id", Value: value}
}

// PostAuthorIdNotLike not like condition
func PostAuthorIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "author_id", Value: value}
}

// PostIdIn condition
func PostIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// PostAuthorIdIn condition
func PostAuthorIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "author_id", Values: values}
}

// PostIdNotIn not in condition
func PostIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// PostAuthorIdNotIn not in condition
func PostAuthorIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "author_id", Values: values}
}

// PostIdOrderBy sorts the result in ascending order.
func PostIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// PostAuthorIdOrderBy sorts the result in ascending order.
func PostAuthorIdOrderBy(asc bool) FilterApplier {
	return OrderBy("author_id", asc)
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
		Columns(
			"title",
			"body",
			"author_id",
		).
		Values(
			model.Title,
			model.Body,
			model.AuthorId,
		)

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
	Title    *string
	Body     *string
	AuthorId *string
}

// Update updates an existing Post based on non-nil fields.
func (t *postStorage) Update(ctx context.Context, id int32, updateData *PostUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("posts")
	if updateData.Title != nil {
		query = query.Set("title", updateData.Title)
	}
	if updateData.Body != nil {
		query = query.Set("body", updateData.Body)
	}
	if updateData.AuthorId != nil {
		query = query.Set("author_id", updateData.AuthorId)
	}

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

// SelectForUpdate lock locks the Post for the given ID.
func (t *postStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Post, error) {
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
	var model Post
	if err := model.ScanRow(row); err != nil {
		return nil, fmt.Errorf("failed to scan Post: %w", err)
	}

	return &model, nil
}
