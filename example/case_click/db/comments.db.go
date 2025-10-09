package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
	"time"
)

// commentStorage is a struct for the "comments" table.
type commentStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// CommentCRUDOperations is an interface for managing the comments table.
type CommentCRUDOperations interface {
	Create(ctx context.Context, model *Comment, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Comment, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Comment, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Comment, opts ...Option) error
}

// CommentSearchOperations is an interface for searching the comments table.
type CommentSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Comment, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Comment, error)
}

type CommentSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) CommentStorage
	SetQueryBuilder(builder sq.StatementBuilderType) CommentStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// CommentRelationLoading is an interface for loading relations.
type CommentRelationLoading interface {
	LoadUser(ctx context.Context, model *Comment, builders ...*QueryBuilder) error
	LoadPost(ctx context.Context, model *Comment, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Comment, builders ...*QueryBuilder) error
	LoadBatchPost(ctx context.Context, items []*Comment, builders ...*QueryBuilder) error
}

// CommentRawQueryOperations is an interface for executing raw queries.
type CommentRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// CommentStorage is a struct for the "comments" table.
type CommentStorage interface {
	CommentCRUDOperations
	CommentSearchOperations
	CommentRelationLoading
	CommentRawQueryOperations
	CommentSettings
}

// NewCommentStorage returns a new commentStorage.
func NewCommentStorage(config *Config) (CommentStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &commentStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *commentStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *commentStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *commentStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
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
func (t *commentStorage) applySettings(query string, settings map[string]interface{}) string {
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
func (t *commentStorage) TableName() string {
	return "comments"
}

// GetQueryBuilder returns the query builder.
func (t *commentStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *commentStorage) Columns() []string {
	return []string{
		"id", "user_id", "post_id", "text", "created_at", "updated_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *commentStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *commentStorage) SetConfig(config *Config) CommentStorage {
	t.config = config
	return t
}

func (t *commentStorage) SetQueryBuilder(builder sq.StatementBuilderType) CommentStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *commentStorage) LoadUser(ctx context.Context, model *Comment, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.UserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one UserStorage: %w", err)
	}

	model.User = relationModel
	return nil
}

// LoadPost loads the Post relation.
func (t *commentStorage) LoadPost(ctx context.Context, model *Comment, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create PostStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(PostIdEq(model.PostId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one PostStorage: %w", err)
	}

	model.Post = relationModel
	return nil
}

// LoadBatchUser loads the User relation.
func (t *commentStorage) LoadBatchUser(ctx context.Context, items []*Comment, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.UserId)
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
		if v, ok := resultMap[item.UserId]; ok {
			item.User = v
		}
	}

	return nil
}

// LoadBatchPost loads the Post relation.
func (t *commentStorage) LoadBatchPost(ctx context.Context, items []*Comment, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.PostId)
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create PostStorage: %w", err)
	}

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(PostIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many PostStorage: %w", err)
	}
	resultMap := make(map[interface{}]*Post)
	for _, result := range results {
		resultMap[result.Id] = result
	}

	// Assign Post to items
	for _, item := range items {
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.PostId]; ok {
			item.Post = v
		}
	}

	return nil
}

// Comment is a struct for the "comments" table.
type Comment struct {
	Id        string
	UserId    string
	PostId    int32
	Text      string
	CreatedAt time.Time
	UpdatedAt *time.Time
	User      *User
	Post      *Post
}

// TableName returns the table name.
func (t *Comment) TableName() string {
	return "comments"
}

// ScanRow scans a row into a Comment.
func (t *Comment) ScanRow(row driver.Row) error {
	return row.Scan(
		&t.Id,
		&t.UserId,
		&t.PostId,
		&t.Text,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

// CommentFilters is a struct that holds filters for Comment.
type CommentFilters struct {
	Id     *string
	UserId *string
	PostId *int32
}

// CommentIdEq returns a condition that checks if the field equals the value.
func CommentIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// CommentUserIdEq returns a condition that checks if the field equals the value.
func CommentUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// CommentPostIdEq returns a condition that checks if the field equals the value.
func CommentPostIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "post_id", Value: value}
}

// CommentIdNotEq returns a condition that checks if the field equals the value.
func CommentIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// CommentUserIdNotEq returns a condition that checks if the field equals the value.
func CommentUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// CommentPostIdNotEq returns a condition that checks if the field equals the value.
func CommentPostIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "post_id", Value: value}
}

// CommentIdGT greaterThanCondition than condition.
func CommentIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// CommentUserIdGT greaterThanCondition than condition.
func CommentUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// CommentPostIdGT greaterThanCondition than condition.
func CommentPostIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "post_id", Value: value}
}

// CommentIdLT less than condition.
func CommentIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// CommentUserIdLT less than condition.
func CommentUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// CommentPostIdLT less than condition.
func CommentPostIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "post_id", Value: value}
}

// CommentIdGTE greater than or equal condition.
func CommentIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// CommentUserIdGTE greater than or equal condition.
func CommentUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// CommentPostIdGTE greater than or equal condition.
func CommentPostIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "post_id", Value: value}
}

// CommentIdLTE less than or equal condition.
func CommentIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// CommentUserIdLTE less than or equal condition.
func CommentUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// CommentPostIdLTE less than or equal condition.
func CommentPostIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "post_id", Value: value}
}

// CommentIdBetween between condition.
func CommentIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// CommentUserIdBetween between condition.
func CommentUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// CommentPostIdBetween between condition.
func CommentPostIdBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "post_id", Min: min, Max: max}
}

// CommentIdILike iLike condition %
func CommentIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// CommentUserIdILike iLike condition %
func CommentUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// CommentIdLike like condition %
func CommentIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// CommentUserIdLike like condition %
func CommentUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// CommentIdNotLike not like condition
func CommentIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// CommentUserIdNotLike not like condition
func CommentUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// CommentIdIn condition
func CommentIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// CommentUserIdIn condition
func CommentUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// CommentPostIdIn condition
func CommentPostIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "post_id", Values: values}
}

// CommentIdNotIn not in condition
func CommentIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// CommentUserIdNotIn not in condition
func CommentUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// CommentPostIdNotIn not in condition
func CommentPostIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "post_id", Values: values}
}

// CommentIdOrderBy sorts the result in ascending order.
func CommentIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// CommentUserIdOrderBy sorts the result in ascending order.
func CommentUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// CommentPostIdOrderBy sorts the result in ascending order.
func CommentPostIdOrderBy(asc bool) FilterApplier {
	return OrderBy("post_id", asc)
}

// AsyncCreate asynchronously inserts a new Comment.
func (t *commentStorage) AsyncCreate(ctx context.Context, model *Comment, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("comments").
		Columns(
			"user_id",
			"post_id",
			"text",
			"created_at",
			"updated_at",
		).
		Values(
			model.UserId,
			model.PostId,
			model.Text,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create Comment: %w", err)
	}

	return nil
}

// Create creates a new Comment.
func (t *commentStorage) Create(ctx context.Context, model *Comment, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("comments").
		Columns(
			"user_id",
			"post_id",
			"text",
			"created_at",
			"updated_at",
		).
		Values(
			model.UserId,
			model.PostId,
			model.Text,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create Comment: %w", err)
	}

	return nil
}

// BatchCreate creates multiple Comment records in a single batch.
func (t *commentStorage) BatchCreate(ctx context.Context, models []*Comment, opts ...Option) error {
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
			model.UserId,
			model.PostId,
			model.Text,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

// OriginalBatchCreate creates multiple Comment records in a single batch.
func (t *commentStorage) OriginalBatchCreate(ctx context.Context, models []*Comment, opts ...Option) error {
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
			"user_id",
			"post_id",
			"text",
			"created_at",
			"updated_at",
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}

		query = query.Values(
			model.UserId,
			model.PostId,
			model.Text,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

// FindMany finds multiple Comment based on the provided options.
func (t *commentStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Comment, error) {
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

	var results []*Comment
	for rows.Next() {
		model := &Comment{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan Comment: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Comment based on the provided options.
func (t *commentStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Comment, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Comment: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *commentStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *commentStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *commentStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *commentStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *commentStorage) Conn() driver.Conn {
	return t.DB()
}
