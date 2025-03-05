package db

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
)

// messageStorage is a struct for the "messages" table.
type messageStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// MessageCRUDOperations is an interface for managing the messages table.
type MessageCRUDOperations interface {
	Create(ctx context.Context, model *Message, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Message, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Message, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*Message, opts ...Option) error
}

// MessageSearchOperations is an interface for searching the messages table.
type MessageSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Message, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Message, error)
}

type MessageSettings interface {
	Conn() driver.Conn
	SetConfig(config *Config) MessageStorage
	SetQueryBuilder(builder sq.StatementBuilderType) MessageStorage
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

// MessageRawQueryOperations is an interface for executing raw queries.
type MessageRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// MessageStorage is a struct for the "messages" table.
type MessageStorage interface {
	MessageCRUDOperations
	MessageSearchOperations
	MessageRelationLoading
	MessageRawQueryOperations
	MessageSettings
}

// NewMessageStorage returns a new messageStorage.
func NewMessageStorage(config *Config) (MessageStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB connection is nil")
	}

	return &messageStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
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
func (t *messageStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *messageStorage) SetConfig(config *Config) MessageStorage {
	t.config = config
	return t
}

func (t *messageStorage) SetQueryBuilder(builder sq.StatementBuilderType) MessageStorage {
	t.queryBuilder = builder
	return t
}

// LoadBot loads the Bot relation.
func (t *messageStorage) LoadBot(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Message is nil")
	}

	// NewBotStorage creates a new BotStorage.
	s, err := NewBotStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create BotStorage")
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
		return errors.Wrap(err, "failed to find one BotStorage")
	}

	model.Bot = relationModel
	return nil
}

// LoadFromUser loads the FromUser relation.
func (t *messageStorage) LoadFromUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Message is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create UserStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.FromUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find one UserStorage")
	}

	model.FromUser = relationModel
	return nil
}

// LoadToUser loads the ToUser relation.
func (t *messageStorage) LoadToUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Message is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create UserStorage")
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.ToUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find one UserStorage")
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
		return errors.Wrap(err, "failed to create BotStorage")
	}

	// Add the filter for the relation
	// Ensure that requestItems are not empty before adding the builder
	if len(requestItems) > 0 {
		builders = append(builders, FilterBuilder(BotIdIn(requestItems...)))
	}

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return errors.Wrap(err, "failed to find many BotStorage")
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
		if v, ok := resultMap[item.ToUserId]; ok {
			item.ToUser = v
		}
	}

	return nil
}

// Message is a struct for the "messages" table.
type Message struct {
	Id         string
	FromUserId string
	ToUserId   string
	BotId      *string
	Bot        *Bot
	FromUser   *User
	ToUser     *User
}

// TableName returns the table name.
func (t *Message) TableName() string {
	return "messages"
}

// ScanRow scans a row into a Message.
func (t *Message) ScanRow(row driver.Row) error {
	return row.Scan(
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

// AsyncCreate asynchronously inserts a new Message.
func (t *messageStorage) AsyncCreate(ctx context.Context, model *Message, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
	}

	// Set default options
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, false, args...); err != nil {
		return errors.Wrap(err, "failed to asynchronously create Message")
	}

	return nil
}

// Create creates a new Message.
func (t *messageStorage) Create(ctx context.Context, model *Message, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to create Message")
	}

	return nil
}

// BatchCreate creates multiple Message records in a single batch.
func (t *messageStorage) BatchCreate(ctx context.Context, models []*Message, opts ...Option) error {
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
			model.FromUserId,
			model.ToUserId,
			nullValue(model.BotId),
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

// OriginalBatchCreate creates multiple Message records in a single batch.
func (t *messageStorage) OriginalBatchCreate(ctx context.Context, models []*Message, opts ...Option) error {
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
			"from_user_id",
			"to_user_id",
			"bot_id",
		)

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}
		query = query.Values(
			model.FromUserId,
			model.ToUserId,
			nullValue(model.BotId),
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

	var results []*Message
	for rows.Next() {
		model := &Message{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, errors.Wrap(err, "failed to scan Message")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single Message based on the provided options.
func (t *messageStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Message, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne Message")
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *messageStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *messageStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *messageStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *messageStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *messageStorage) Conn() driver.Conn {
	return t.DB()
}
