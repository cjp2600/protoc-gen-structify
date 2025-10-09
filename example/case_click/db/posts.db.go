package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
)

// postStorage is a struct for the "posts" table.
type postStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// PostCRUDOperations is an interface for managing the posts table.
type PostCRUDOperations interface {
	Create(ctx context.Context, model *Post, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Post, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Post, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Post, opts ...Option) error
}

// PostSearchOperations is an interface for searching the posts table.
type PostSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Post, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error)
}

type PostSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) PostStorage
	SetQueryBuilder(builder sq.StatementBuilderType) PostStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// PostRelationLoading is an interface for loading relations.
type PostRelationLoading interface {
	LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error
	LoadBatchAuthor(ctx context.Context, items []*Post, builders ...*QueryBuilder) error
}

// PostRawQueryOperations is an interface for executing raw queries.
type PostRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// PostStorage is a struct for the "posts" table.
type PostStorage interface {
	PostCRUDOperations
	PostSearchOperations
	PostRelationLoading
	PostRawQueryOperations
	PostSettings
}

// NewPostStorage returns a new postStorage.
func NewPostStorage(config *Config) (PostStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &postStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
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

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *postStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
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
func (t *postStorage) applySettings(query string, settings map[string]interface{}) string {
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
func (t *postStorage) TableName() string {
	return "posts"
}

// GetQueryBuilder returns the query builder.
func (t *postStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *postStorage) Columns() []string {
	return []string{
		"id", "title", "body", "author_id",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *postStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *postStorage) SetConfig(config *Config) PostStorage {
	t.config = config
	return t
}

func (t *postStorage) SetQueryBuilder(builder sq.StatementBuilderType) PostStorage {
	t.queryBuilder = builder
	return t
}

// LoadAuthor loads the Author relation.
func (t *postStorage) LoadAuthor(ctx context.Context, model *Post, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.AuthorId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one UserStorage: %w", err)
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
		if v, ok := resultMap[item.AuthorId]; ok {
			item.Author = v
		}
	}

	return nil
}

// Post is a struct for the "posts" table.
type Post struct {
	Id       int32
	Title    string
	Body     string
	Author   *User
	AuthorId string
}

// TableName returns the table name.
func (t *Post) TableName() string {
	return "posts"
}

// ScanRow scans a row into a Post.
func (t *Post) ScanRow(row driver.Row) error {
	return row.Scan(
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

// AsyncCreate asynchronously inserts a new Post.
func (t *postStorage) AsyncCreate(ctx context.Context, model *Post, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create Post: %w", err)
	}

	return nil
}

// Create creates a new Post.
func (t *postStorage) Create(ctx context.Context, model *Post, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create Post: %w", err)
	}

	return nil
}

// BatchCreate creates multiple Post records in a single batch.
func (t *postStorage) BatchCreate(ctx context.Context, models []*Post, opts ...Option) error {
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
			model.Title,
			model.Body,
			model.AuthorId,
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

// OriginalBatchCreate creates multiple Post records in a single batch.
func (t *postStorage) OriginalBatchCreate(ctx context.Context, models []*Post, opts ...Option) error {
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
			"title",
			"body",
			"author_id",
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}

		query = query.Values(
			model.Title,
			model.Body,
			model.AuthorId,
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

// FindMany finds multiple Post based on the provided options.
func (t *postStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Post, error) {
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

	var results []*Post
	for rows.Next() {
		model := &Post{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan Post: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Post based on the provided options.
func (t *postStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Post, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Post: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *postStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *postStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *postStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *postStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *postStorage) Conn() driver.Conn {
	return t.DB()
}
