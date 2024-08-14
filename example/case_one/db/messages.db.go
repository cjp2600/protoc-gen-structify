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

// messageStorage is a struct for the "messages" table.
type messageStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

// MessageTableManager is an interface for managing the messages table.
type MessageTableManager interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
}

// MessageCRUDOperations is an interface for managing the messages table.
type MessageCRUDOperations interface {
	Create(ctx context.Context, model *Message, opts ...Option) (*string, error)
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
	LoadFromUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error
	LoadToUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error
	LoadBatchFromUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error
	LoadBatchToUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error
}

// MessageAdvancedDeletion is an interface for advanced deletion operations.
type MessageAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// MessageRawQueryOperations is an interface for executing raw queries.
type MessageRawQueryOperations interface {
	Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// MessageStorage is a struct for the "messages" table.
type MessageStorage interface {
	MessageTableManager
	MessageCRUDOperations
	MessageSearchOperations
	MessagePaginationOperations
	MessageRelationLoading
	MessageAdvancedDeletion
	MessageRawQueryOperations
}

// NewMessageStorage returns a new messageStorage.
func NewMessageStorage(db *sql.DB) MessageStorage {
	return &messageStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *messageStorage) TableName() string {
	return "messages"
}

// Columns returns the columns for the table.
func (t *messageStorage) Columns() []string {
	return []string{
		"id", "from_user_id", "to_user_id",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *messageStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *messageStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		-- Table: messages
		CREATE TABLE IF NOT EXISTS messages (
		id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
		from_user_id UUID NOT NULL,
		to_user_id UUID NOT NULL);
		-- Other entities
		CREATE UNIQUE INDEX IF NOT EXISTS messages_from_user_id_unique_idx ON messages USING btree (from_user_id);
		CREATE UNIQUE INDEX IF NOT EXISTS messages_to_user_id_unique_idx ON messages USING btree (to_user_id);
		CREATE INDEX IF NOT EXISTS messages_from_user_id_idx ON messages USING btree (from_user_id);
		CREATE INDEX IF NOT EXISTS messages_to_user_id_idx ON messages USING btree (to_user_id);
		-- Foreign keys for users
		ALTER TABLE messages
		ADD FOREIGN KEY (to_user_id) REFERENCES users(id);
		-- Foreign keys for users
		ALTER TABLE messages
		ADD FOREIGN KEY (to_user_id) REFERENCES users(id);
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *messageStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS messages;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *messageStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE messages;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *messageStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// LoadFromUser loads the FromUser relation.
func (t *messageStorage) LoadFromUser(ctx context.Context, model *Message, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Message is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s := NewUserStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(UserIdEq(model.FromUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find UserStorage: %w", err)
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
	s := NewUserStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(UserIdEq(model.ToUserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find UserStorage: %w", err)
	}

	model.ToUser = relationModel
	return nil
}

// LoadBatchFromUser loads the FromUser relation.
func (t *messageStorage) LoadBatchFromUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.FromUserId
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
		if v, ok := resultMap[item.FromUserId]; ok {
			item.FromUser = v
		}
	}

	return nil
}

// LoadBatchToUser loads the ToUser relation.
func (t *messageStorage) LoadBatchToUser(ctx context.Context, items []*Message, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.ToUserId
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
		if v, ok := resultMap[item.ToUserId]; ok {
			item.ToUser = v
		}
	}

	return nil
}

// Message is a struct for the "messages" table.
type Message struct {
	Id         string `db:"id"`
	FromUserId string `db:"from_user_id"`
	ToUserId   string `db:"to_user_id"`
	FromUser   *User
	ToUser     *User
}

// TableName returns the table name.
func (t *Message) TableName() string {
	return "messages"
}

// ScanRow scans a row into a Message.
func (t *Message) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.FromUserId, &t.ToUserId)
}

// ScanRows scans a single row into the Message.
func (t *Message) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.FromUserId,
		&t.ToUserId,
	)
}

// MessageFilters is a struct that holds filters for Message.
type MessageFilters struct {
	Id       *string
	ToUserId *string
}

// MessageIdEq returns a condition that checks if the field equals the value.
func MessageIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// MessageToUserIdEq returns a condition that checks if the field equals the value.
func MessageToUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "to_user_id", Value: value}
}

// MessageIdNotEq returns a condition that checks if the field equals the value.
func MessageIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// MessageToUserIdNotEq returns a condition that checks if the field equals the value.
func MessageToUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "to_user_id", Value: value}
}

// MessageIdGT greaterThanCondition than condition.
func MessageIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// MessageToUserIdGT greaterThanCondition than condition.
func MessageToUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "to_user_id", Value: value}
}

// MessageIdLT less than condition.
func MessageIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// MessageToUserIdLT less than condition.
func MessageToUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "to_user_id", Value: value}
}

// MessageIdGTE greater than or equal condition.
func MessageIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// MessageToUserIdGTE greater than or equal condition.
func MessageToUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "to_user_id", Value: value}
}

// MessageIdLTE less than or equal condition.
func MessageIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// MessageToUserIdLTE less than or equal condition.
func MessageToUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "to_user_id", Value: value}
}

// MessageIdBetween between condition.
func MessageIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// MessageToUserIdBetween between condition.
func MessageToUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "to_user_id", Min: min, Max: max}
}

// MessageIdLike like condition %
func MessageIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// MessageToUserIdLike like condition %
func MessageToUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "to_user_id", Value: value}
}

// MessageIdNotLike not like condition
func MessageIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// MessageToUserIdNotLike not like condition
func MessageToUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "to_user_id", Value: value}
}

// MessageIdIn condition
func MessageIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// MessageToUserIdIn condition
func MessageToUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "to_user_id", Values: values}
}

// MessageIdNotIn not in condition
func MessageIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// MessageToUserIdNotIn not in condition
func MessageToUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "to_user_id", Values: values}
}

// MessageIdOrderBy sorts the result in ascending order.
func MessageIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// MessageToUserIdOrderBy sorts the result in ascending order.
func MessageToUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("to_user_id", asc)
}

// Create creates a new Message.
func (t *messageStorage) Create(ctx context.Context, model *Message, opts ...Option) (*string, error) {
	if model == nil {
		return nil, errors.New("model is nil")
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
		).
		Values(
			model.FromUserId,
			model.ToUserId,
		)

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

		return nil, fmt.Errorf("failed to create Message: %w", err)
	}

	return &id, nil
}

// MessageUpdate is used to update an existing Message.
type MessageUpdate struct {
	FromUserId *string
	ToUserId   *string
}

// Update updates an existing Message based on non-nil fields.
func (t *messageStorage) Update(ctx context.Context, id string, updateData *MessageUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("messages")
	if updateData.FromUserId != nil {
		query = query.Set("from_user_id", updateData.FromUserId)
	}
	if updateData.ToUserId != nil {
		query = query.Set("to_user_id", updateData.ToUserId)
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
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

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
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
		return errors.New("filters are required for delete operation")
	}

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
		return nil, errors.Wrap(err, "find one Message: ")
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
		return nil, fmt.Errorf("failed to find Message: %w", err)
	}
	defer rows.Close()

	var results []*Message
	for rows.Next() {
		model := &Message{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Message: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Message based on the provided options.
func (t *messageStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Message, error) {
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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("failed to count Message: %w", err)
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
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var model Message
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan Message: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
func (t *messageStorage) Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *messageStorage) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *messageStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx).QueryContext(ctx, query, args...)
}
