package db

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math"
)

// postStorage is a struct for the "posts" table.
type postStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// PostCRUDOperations is an interface for managing the posts table.
type PostCRUDOperations interface {
	Create(ctx context.Context, model *Post, opts ...Option) (*int32, error)
	BatchCreate(ctx context.Context, models []*Post, opts ...Option) ([]string, error)
	Update(ctx context.Context, id int32, updateData *PostUpdate) error
	DeleteById(ctx context.Context, id int32, opts ...Option) error
	FindById(ctx context.Context, id int32, opts ...Option) (*Post, error)
}

// PostSearchOperations is an interface for searching the posts table.
type PostSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Post, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Post, error)
}

// PostPaginationOperations is an interface for pagination operations.
type PostPaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Post, *Paginator, error)
}

// PostRelationLoading is an interface for loading relations.
type PostRelationLoading interface {
	LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error
	LoadBatchAuthor(ctx context.Context, items []*Post, builders ...*QueryBuilder) error
}

// PostAdvancedDeletion is an interface for advanced deletion operations.
type PostAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// PostRawQueryOperations is an interface for executing raw queries.
type PostRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// PostStorage is a struct for the "posts" table.
type PostStorage interface {
	PostCRUDOperations
	PostSearchOperations
	PostPaginationOperations
	PostRelationLoading
	PostAdvancedDeletion
	PostRawQueryOperations
}

// NewPostStorage returns a new postStorage.
func NewPostStorage(config *Config) (PostStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB is nil")
	}
	if config.DB.DBRead == nil {
		return nil, errors.New("config.DB.DBRead is nil")
	}
	if config.DB.DBWrite == nil {
		config.DB.DBWrite = config.DB.DBRead
	}

	return &postStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *postStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *postStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
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

// DB returns the underlying DB. This is useful for doing transactions.
func (t *postStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	var db QueryExecer

	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		if tx == nil {
			t.logError(ctx, errors.New("transaction is nil"), "failed to get transaction from context")
			// set default connection
			return t.config.DB.DBWrite
		}

		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		db = t.config.DB.DBWrite
	} else {
		db = t.config.DB.DBRead
	}

	return db
}

// LoadAuthor loads the Author relation.
func (t *postStorage) LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Post is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create UserStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.AuthorId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find one UserStorage")
	}

	model.Author = relationModel
	return nil
}

// LoadBatchAuthor loads the Author relation.
func (t *postStorage) LoadBatchAuthor(ctx context.Context, items []*Post, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.AuthorId)
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create UserStorage")
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(UserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many UserStorage")
	}
	resultMap := make(map[interface{}]*User)
	for _, result := range results {
		resultMap[result.Id] = result
	}

	// Assign User to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
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

// PostAuthorIdILike iLike condition %
func PostAuthorIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "author_id", Value: value}
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
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id int32
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, errors.Wrap(err, "failed to create Post")
	}

	return &id, nil
}

// BatchCreate creates multiple Post records in a single batch.
func (t *postStorage) BatchCreate(ctx context.Context, models []*Post, opts ...Option) ([]string, error) {
	if len(models) == 0 {
		return nil, errors.New("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return nil, errors.New("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"title",
			"body",
			"author_id",
		)

	for _, model := range models {
		if model == nil {
			return nil, errors.New("one of the models is nil")
		}
		query = query.Values(
			model.Title,
			model.Body,
			model.AuthorId,
		)
	}

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING RETURNING \"id\"")
	} else {
		query = query.Suffix("RETURNING \"id\"")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return nil, errors.Wrap(err, "failed to execute bulk insert")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var returnIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, errors.Wrap(err, "failed to scan id")
		}
		returnIDs = append(returnIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows iteration error")
	}

	return returnIDs, nil
}

// PostUpdate is used to update an existing Post.
type PostUpdate struct {
	// Use regular pointer types for non-optional fields
	Title *string
	// Use regular pointer types for non-optional fields
	Body *string
	// Use regular pointer types for non-optional fields
	AuthorId *string
}

// Update updates an existing Post based on non-nil fields.
func (t *postStorage) Update(ctx context.Context, id int32, updateData *PostUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("posts")
	// Handle fields that are not optional using a nil check
	if updateData.Title != nil {
		query = query.Set("title", *updateData.Title) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Body != nil {
		query = query.Set("body", *updateData.Body) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.AuthorId != nil {
		query = query.Set("author_id", *updateData.AuthorId) // Dereference pointer value
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update Post")
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
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete Post")
	}

	return nil
}

// DeleteMany removes entries from the posts table using the provided filters
func (t *postStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("posts")

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
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete posts")
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

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*Post
	for rows.Next() {
		model := &Post{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan Post")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single Post based on the provided options.
func (t *postStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne Post")
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

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to scan count")
	}

	return count, nil
}

// FindManyWithPagination finds multiple Post with pagination support.
func (t *postStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Post, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count Post")
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
		return nil, nil, errors.Wrap(err, "failed to find Post")
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

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model Post
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, errors.Wrap(err, "failed to scan Post")
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *postStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *postStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *postStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
