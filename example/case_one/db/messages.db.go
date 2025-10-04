package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"gopkg.in/guregu/null.v4"
	"math"
)

// messageStorage is a struct for the "messages" table.
type messageStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// MessageCRUDOperations is an interface for managing the messages table.
type MessageCRUDOperations interface {
	Create(ctx context.Context, model *Message, opts ...Option) (*string, error)
	BatchCreate(ctx context.Context, models []*Message, opts ...Option) ([]string, error)
	Update(ctx context.Context, id string, updateData *MessageUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*Message, error)
}

// MessageSearchOperations is an interface for searching the messages table.
type MessageSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Message, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Message, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Message, error)
}

// MessagePaginationOperations is an interface for pagination operations.
type MessagePaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Message, *Paginator, error)
}

// MessageRelationLoading is an interface for loading relations.
type MessageRelationLoading interface {
	LoadBot(ctx context.Context, model *Message, builders ...*QueryBuilder) error
	LoadFromUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error
	LoadToUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error
	LoadBatchBot(ctx context.Context, items []*Message, builders ...*QueryBuilder) error
	LoadBatchFromUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error
	LoadBatchToUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error
}

// MessageAdvancedDeletion is an interface for advanced deletion operations.
type MessageAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// MessageRawQueryOperations is an interface for executing raw queries.
type MessageRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// MessageStorage is a struct for the "messages" table.
type MessageStorage interface {
	MessageCRUDOperations
	MessageSearchOperations
	MessagePaginationOperations
	MessageRelationLoading
	MessageAdvancedDeletion
	MessageRawQueryOperations
}

// NewMessageStorage returns a new messageStorage.
func NewMessageStorage(config *Config) (MessageStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB is nil")
	}
	if config.DB.DBRead == nil {
		return nil, fmt.Errorf("config.DB.DBRead is nil")
	}
	if config.DB.DBWrite == nil {
		config.DB.DBWrite = config.DB.DBRead
	}

	return &messageStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *messageStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *messageStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// TableName returns the table name.
func (t *messageStorage) TableName() string {
	return "messages"
}

// Columns returns the columns for the table.
func (t *messageStorage) Columns() []string {
	return []string{
		"id", "from_user_id", "to_user_id", "bot_id",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *messageStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		if tx == nil {
			t.logError(ctx, fmt.Errorf("transaction is nil"), "failed to get transaction from context")
			// set default connection
			return &dbWrapper{db: t.config.DB.DBWrite}
		}

		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		return &dbWrapper{db: t.config.DB.DBWrite}
	} else {
		return &dbWrapper{db: t.config.DB.DBRead}
	}
}

// LoadBot loads the Bot relation.
func (t *messageStorage) LoadBot(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("Message is nil")
	}

	// NewBotStorage creates a new BotStorage.
	s, err := NewBotStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create BotStorage: %w", err)
	}
	// Check if the optional field is nil
	if model.BotId == nil {
		// If nil, do not attempt to load the relation
		return nil
	}
	// Add the filter for the relation with dereferenced value
	builders = append(builders, FilterBuilder(BotIdEq(*model.BotId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one BotStorage: %w", err)
	}

	model.Bot = relationModel
	return nil
}

// LoadFromUser loads the FromUser relation.
func (t *messageStorage) LoadFromUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("Message is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.FromUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one UserStorage: %w", err)
	}

	model.FromUser = relationModel
	return nil
}

// LoadToUser loads the ToUser relation.
func (t *messageStorage) LoadToUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("Message is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create UserStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.ToUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one UserStorage: %w", err)
	}

	model.ToUser = relationModel
	return nil
}

// LoadBatchBot loads the Bot relation.
func (t *messageStorage) LoadBatchBot(ctx context.Context, items []*Message, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Check if the field is nil for optional fields
		if item.BotId == nil {
			// Skip nil values for optional fields
			continue
		}
		// Append dereferenced value for optional fields
		requestItems = append(requestItems, *item.BotId)
	}

	// NewBotStorage creates a new BotStorage.
	s, err := NewBotStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create BotStorage: %w", err)
	}

	// Add the filter for the relation
	// Ensure that requestItems are not empty before adding the builder
	if len(requestItems) > 0 {
		builders = append(builders, FilterBuilder(BotIdIn(requestItems...)))
	}

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many BotStorage: %w", err)
	}
	resultMap := make(map[interface{}]*Bot)
	for _, result := range results {
		resultMap[result.Id] = result
	}

	// Assign Bot to items
	for _, item := range items {
		// Skip assignment if the field is nil
		if item.BotId == nil {
			continue
		}
		// Assign the relation if it exists in the resultMap
		if v, ok := resultMap[*item.BotId]; ok {
			item.Bot = v
		}
	}

	return nil
}

// LoadBatchFromUser loads the FromUser relation.
func (t *messageStorage) LoadBatchFromUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.FromUserId)
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
		if v, ok := resultMap[item.FromUserId]; ok {
			item.FromUser = v
		}
	}

	return nil
}

// LoadBatchToUser loads the ToUser relation.
func (t *messageStorage) LoadBatchToUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.ToUserId)
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
		if v, ok := resultMap[item.ToUserId]; ok {
			item.ToUser = v
		}
	}

	return nil
}

// Message is a struct for the "messages" table.
type Message struct {
	Id         string  `db:"id"`
	FromUserId string  `db:"from_user_id"`
	ToUserId   string  `db:"to_user_id"`
	BotId      *string `db:"bot_id"`
	Bot        *Bot
	FromUser   *User
	ToUser     *User
}

// TableName returns the table name.
func (t *Message) TableName() string {
	return "messages"
}

// ScanRow scans a row into a Message.
func (t *Message) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.FromUserId, &t.ToUserId, &t.BotId)
}

// ScanRows scans a single row into the Message.
func (t *Message) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.FromUserId,
		&t.ToUserId,
		&t.BotId,
	)
}

// MessageFilters is a struct that holds filters for Message.
type MessageFilters struct {
	Id       *string
	ToUserId *string
	BotId    *string
}

// MessageIdEq returns a condition that checks if the field equals the value.
func MessageIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// MessageToUserIdEq returns a condition that checks if the field equals the value.
func MessageToUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdEq returns a condition that checks if the field equals the value.
func MessageBotIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "bot_id", Value: value}
}

// MessageIdNotEq returns a condition that checks if the field equals the value.
func MessageIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// MessageToUserIdNotEq returns a condition that checks if the field equals the value.
func MessageToUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdNotEq returns a condition that checks if the field equals the value.
func MessageBotIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "bot_id", Value: value}
}

// MessageIdGT greaterThanCondition than condition.
func MessageIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// MessageToUserIdGT greaterThanCondition than condition.
func MessageToUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdGT greaterThanCondition than condition.
func MessageBotIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "bot_id", Value: value}
}

// MessageIdLT less than condition.
func MessageIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// MessageToUserIdLT less than condition.
func MessageToUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdLT less than condition.
func MessageBotIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "bot_id", Value: value}
}

// MessageIdGTE greater than or equal condition.
func MessageIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// MessageToUserIdGTE greater than or equal condition.
func MessageToUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdGTE greater than or equal condition.
func MessageBotIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "bot_id", Value: value}
}

// MessageIdLTE less than or equal condition.
func MessageIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// MessageToUserIdLTE less than or equal condition.
func MessageToUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdLTE less than or equal condition.
func MessageBotIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "bot_id", Value: value}
}

// MessageIdBetween between condition.
func MessageIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// MessageToUserIdBetween between condition.
func MessageToUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "to_user_id", Min: min, Max: max}
}

// MessageBotIdBetween between condition.
func MessageBotIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "bot_id", Min: min, Max: max}
}

// MessageBotIdIsNull checks if the bot_id is NULL.
func MessageBotIdIsNull() FilterApplier {
	return IsNullCondition{Field: "bot_id"}
}

// MessageBotIdIsNotNull checks if the bot_id is NOT NULL.
func MessageBotIdIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "bot_id"}
}

// MessageIdILike iLike condition %
func MessageIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// MessageToUserIdILike iLike condition %
func MessageToUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdILike iLike condition %
func MessageBotIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "bot_id", Value: value}
}

// MessageIdLike like condition %
func MessageIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// MessageToUserIdLike like condition %
func MessageToUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdLike like condition %
func MessageBotIdLike(value string) FilterApplier {
	return LikeCondition{Field: "bot_id", Value: value}
}

// MessageIdNotLike not like condition
func MessageIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// MessageToUserIdNotLike not like condition
func MessageToUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "to_user_id", Value: value}
}

// MessageBotIdNotLike not like condition
func MessageBotIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "bot_id", Value: value}
}

// MessageIdIn condition
func MessageIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// MessageToUserIdIn condition
func MessageToUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "to_user_id", Values: values}
}

// MessageBotIdIn condition
func MessageBotIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "bot_id", Values: values}
}

// MessageIdNotIn not in condition
func MessageIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// MessageToUserIdNotIn not in condition
func MessageToUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "to_user_id", Values: values}
}

// MessageBotIdNotIn not in condition
func MessageBotIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "bot_id", Values: values}
}

// MessageIdOrderBy sorts the result in ascending order.
func MessageIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// MessageToUserIdOrderBy sorts the result in ascending order.
func MessageToUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("to_user_id", asc)
}

// MessageBotIdOrderBy sorts the result in ascending order.
func MessageBotIdOrderBy(asc bool) FilterApplier {
	return OrderBy("bot_id", asc)
}

// Create creates a new Message.
func (t *messageStorage) Create(ctx context.Context, model *Message, opts ...Option) (*string, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("messages").
		Columns(
			"from_user_id",
			"to_user_id",
			"bot_id",
		).
		Values(
			model.FromUserId,
			model.ToUserId,
			nullValue(model.BotId),
		)

	// add RETURNING "id" to query
	query = query.Suffix("RETURNING \"id\"")

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, fmt.Errorf("failed to create Message: %w", err)
	}

	return &id, nil
}

// BatchCreate creates multiple Message records in a single batch.
func (t *messageStorage) BatchCreate(ctx context.Context, models []*Message, opts ...Option) ([]string, error) {
	if len(models) == 0 {
		return nil, fmt.Errorf("no models to insert")
	}

	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	if options.relations {
		return nil, fmt.Errorf("relations are not supported in batch create")
	}

	query := t.queryBuilder.Insert(t.TableName()).
		Columns(
			"from_user_id",
			"to_user_id",
			"bot_id",
		)

	for _, model := range models {
		if model == nil {
			return nil, fmt.Errorf("one of the models is nil")
		}
		query = query.Values(
			model.FromUserId,
			model.ToUserId,
			nullValue(model.BotId),
		)
	}

	if options.ignoreConflictField != "" {
		query = query.Suffix("ON CONFLICT (" + options.ignoreConflictField + ") DO NOTHING RETURNING \"id\"")
	} else {
		query = query.Suffix("RETURNING \"id\"")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return nil, fmt.Errorf("failed to execute bulk insert: %w", err)
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
			return nil, fmt.Errorf("failed to scan id: %w", err)
		}
		returnIDs = append(returnIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return returnIDs, nil
}

// MessageUpdate is used to update an existing Message.
type MessageUpdate struct {
	// Use regular pointer types for non-optional fields
	FromUserId *string
	// Use regular pointer types for non-optional fields
	ToUserId *string
	// Use null types for optional fields
	BotId null.String
}

// Update updates an existing Message based on non-nil fields.
func (t *messageStorage) Update(ctx context.Context, id string, updateData *MessageUpdate) error {
	if updateData == nil {
		return fmt.Errorf("update data is nil")
	}

	query := t.queryBuilder.Update("messages")
	// Handle fields that are not optional using a nil check
	if updateData.FromUserId != nil {
		query = query.Set("from_user_id", *updateData.FromUserId) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.ToUserId != nil {
		query = query.Set("to_user_id", *updateData.ToUserId) // Dereference pointer value
	}
	// Handle fields that are optional and can be explicitly set to NULL
	if updateData.BotId.Valid {
		// Handle null.String specifically
		if updateData.BotId.String == "" {
			query = query.Set("bot_id", nil) // Explicitly set NULL for empty string
		} else {
			query = query.Set("bot_id", updateData.BotId.ValueOrZero())
		}
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Message: %w", err)
	}

	return nil
}

// DeleteById - deletes a Message by its id.
func (t *messageStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("messages").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Message: %w", err)
	}

	return nil
}

// DeleteMany removes entries from the messages table using the provided filters
func (t *messageStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("messages")

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
		return fmt.Errorf("filters are required for delete operation")
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete messages: %w", err)
	}

	return nil
}

// FindById retrieves a Message by its id.
func (t *messageStorage) FindById(ctx context.Context, id string, opts ...Option) (*Message, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(MessageIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, fmt.Errorf("find one Message: %w", err)
	}

	return model, nil
}

// FindMany finds multiple Message based on the provided options.
func (t *messageStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Message, error) {
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
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			t.logError(ctx, err, "failed to close rows")
		}
	}()

	var results []*Message
	for rows.Next() {
		model := &Message{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Message: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Message based on the provided options.
func (t *messageStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Message, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Message: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Count counts Message based on the provided options.
func (t *messageStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to scan count: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Message with pagination support.
func (t *messageStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Message, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Message: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Message: %w", err)
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the Message for the given ID.
func (t *messageStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Message, error) {
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
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model Message
	if err := model.ScanRow(row); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan Message: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *messageStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *messageStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *messageStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
