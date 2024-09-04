package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math"
	"time"
)

// botStorage is a struct for the "bots" table.
type botStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

// BotTableManager is an interface for managing the bots table.
type BotTableManager interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
}

// BotCRUDOperations is an interface for managing the bots table.
type BotCRUDOperations interface {
	Create(ctx context.Context, model *Bot, opts ...Option) (*string, error)
	Update(ctx context.Context, id string, updateData *BotUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*Bot, error)
}

// BotSearchOperations is an interface for searching the bots table.
type BotSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Bot, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Bot, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Bot, error)
}

// BotPaginationOperations is an interface for pagination operations.
type BotPaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Bot, *Paginator, error)
}

// BotRelationLoading is an interface for loading relations.
type BotRelationLoading interface {
	LoadUser(ctx context.Context, model *Bot, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Bot, builders ...*QueryBuilder) error
}

// BotAdvancedDeletion is an interface for advanced deletion operations.
type BotAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// BotRawQueryOperations is an interface for executing raw queries.
type BotRawQueryOperations interface {
	Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// BotStorage is a struct for the "bots" table.
type BotStorage interface {
	BotTableManager
	BotCRUDOperations
	BotSearchOperations
	BotPaginationOperations
	BotRelationLoading
	BotAdvancedDeletion
	BotRawQueryOperations
}

// NewBotStorage returns a new botStorage.
func NewBotStorage(db *sql.DB) BotStorage {
	return &botStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *botStorage) TableName() string {
	return "bots"
}

// Columns returns the columns for the table.
func (t *botStorage) Columns() []string {
	return []string{
		"id", "user_id", "name", "token", "is_publish", "created_at", "updated_at", "deleted_at",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *botStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *botStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		-- Table: bots
		CREATE TABLE IF NOT EXISTS bots (
		id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
		user_id UUID NOT NULL,
		name TEXT,
		token TEXT NOT NULL,
		is_publish BOOLEAN,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP,
		deleted_at TIMESTAMP);
		-- Other entities
		CREATE UNIQUE INDEX IF NOT EXISTS bots_token_unique_idx ON bots USING btree (token);
		-- Foreign keys for users
		ALTER TABLE bots
		ADD FOREIGN KEY (user_id) REFERENCES users(id)
		ON DELETE CASCADE;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *botStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS bots;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *botStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE bots;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *botStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// LoadUser loads the User relation.
func (t *botStorage) LoadUser(ctx context.Context, model *Bot, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Bot is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s := NewUserStorage(t.db)
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(UserIdEq(model.UserId)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find UserStorage: %w", err)
	}

	model.User = relationModel
	return nil
}

// LoadBatchUser loads the User relation.
func (t *botStorage) LoadBatchUser(ctx context.Context, items []*Bot, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.UserId)
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
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.UserId]; ok {
			item.User = v
		}
	}

	return nil
}

// Bot is a struct for the "bots" table.
type Bot struct {
	Id        string     `db:"id"`
	UserId    string     `db:"user_id"`
	Name      string     `db:"name"`
	Token     string     `db:"token"`
	IsPublish bool       `db:"is_publish"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
	User      *User
}

// TableName returns the table name.
func (t *Bot) TableName() string {
	return "bots"
}

// ScanRow scans a row into a Bot.
func (t *Bot) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.UserId, &t.Name, &t.Token, &t.IsPublish, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)
}

// ScanRows scans a single row into the Bot.
func (t *Bot) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.UserId,
		&t.Name,
		&t.Token,
		&t.IsPublish,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.DeletedAt,
	)
}

// BotFilters is a struct that holds filters for Bot.
type BotFilters struct {
	Id        *string
	UserId    *string
	CreatedAt *time.Time
}

// BotIdEq returns a condition that checks if the field equals the value.
func BotIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// BotUserIdEq returns a condition that checks if the field equals the value.
func BotUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// BotCreatedAtEq returns a condition that checks if the field equals the value.
func BotCreatedAtEq(value time.Time) FilterApplier {
	return EqualsCondition{Field: "created_at", Value: value}
}

// BotIdNotEq returns a condition that checks if the field equals the value.
func BotIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// BotUserIdNotEq returns a condition that checks if the field equals the value.
func BotUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// BotCreatedAtNotEq returns a condition that checks if the field equals the value.
func BotCreatedAtNotEq(value time.Time) FilterApplier {
	return NotEqualsCondition{Field: "created_at", Value: value}
}

// BotIdGT greaterThanCondition than condition.
func BotIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// BotUserIdGT greaterThanCondition than condition.
func BotUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// BotCreatedAtGT greaterThanCondition than condition.
func BotCreatedAtGT(value time.Time) FilterApplier {
	return GreaterThanCondition{Field: "created_at", Value: value}
}

// BotIdLT less than condition.
func BotIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// BotUserIdLT less than condition.
func BotUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// BotCreatedAtLT less than condition.
func BotCreatedAtLT(value time.Time) FilterApplier {
	return LessThanCondition{Field: "created_at", Value: value}
}

// BotIdGTE greater than or equal condition.
func BotIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// BotUserIdGTE greater than or equal condition.
func BotUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotCreatedAtGTE greater than or equal condition.
func BotCreatedAtGTE(value time.Time) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotIdLTE less than or equal condition.
func BotIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// BotUserIdLTE less than or equal condition.
func BotUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// BotCreatedAtLTE less than or equal condition.
func BotCreatedAtLTE(value time.Time) FilterApplier {
	return LessThanOrEqualCondition{Field: "created_at", Value: value}
}

// BotIdBetween between condition.
func BotIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// BotUserIdBetween between condition.
func BotUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// BotCreatedAtBetween between condition.
func BotCreatedAtBetween(min, max time.Time) FilterApplier {
	return BetweenCondition{Field: "created_at", Min: min, Max: max}
}

// BotIdLike like condition %
func BotIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// BotUserIdLike like condition %
func BotUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// BotIdNotLike not like condition
func BotIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// BotUserIdNotLike not like condition
func BotUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// BotIdIn condition
func BotIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// BotUserIdIn condition
func BotUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// BotCreatedAtIn condition
func BotCreatedAtIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "created_at", Values: values}
}

// BotIdNotIn not in condition
func BotIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// BotUserIdNotIn not in condition
func BotUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// BotCreatedAtNotIn not in condition
func BotCreatedAtNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "created_at", Values: values}
}

// BotIdOrderBy sorts the result in ascending order.
func BotIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// BotUserIdOrderBy sorts the result in ascending order.
func BotUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// BotCreatedAtOrderBy sorts the result in ascending order.
func BotCreatedAtOrderBy(asc bool) FilterApplier {
	return OrderBy("created_at", asc)
}

// Create creates a new Bot.
func (t *botStorage) Create(ctx context.Context, model *Bot, opts ...Option) (*string, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("bots").
		Columns(
			"user_id",
			"name",
			"token",
			"is_publish",
			"created_at",
			"updated_at",
			"deleted_at",
		).
		Values(
			model.UserId,
			model.Name,
			model.Token,
			model.IsPublish,
			model.CreatedAt,
			model.UpdatedAt,
			model.DeletedAt,
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

		return nil, fmt.Errorf("failed to create Bot: %w", err)
	}

	return &id, nil
}

// BotUpdate is used to update an existing Bot.
type BotUpdate struct {
	UserId    *string
	Name      *string
	Token     *string
	IsPublish *bool
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

// Update updates an existing Bot based on non-nil fields.
func (t *botStorage) Update(ctx context.Context, id string, updateData *BotUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("bots")
	if updateData.UserId != nil {
		query = query.Set("user_id", updateData.UserId)
	}
	if updateData.Name != nil {
		query = query.Set("name", updateData.Name)
	}
	if updateData.Token != nil {
		query = query.Set("token", updateData.Token)
	}
	if updateData.IsPublish != nil {
		query = query.Set("is_publish", updateData.IsPublish)
	}
	if updateData.CreatedAt != nil {
		query = query.Set("created_at", updateData.CreatedAt)
	}
	if updateData.UpdatedAt != nil {
		query = query.Set("updated_at", updateData.UpdatedAt)
	}
	if updateData.DeletedAt != nil {
		query = query.Set("deleted_at", updateData.DeletedAt)
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Bot: %w", err)
	}

	return nil
}

// DeleteById - deletes a Bot by its id.
func (t *botStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("bots").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Bot: %w", err)
	}

	return nil
}

// DeleteMany removes entries from the bots table using the provided filters
func (t *botStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("bots")

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

// FindById retrieves a Bot by its id.
func (t *botStorage) FindById(ctx context.Context, id string, opts ...Option) (*Bot, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(BotIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one Bot: ")
	}

	return model, nil
}

// FindMany finds multiple Bot based on the provided options.
func (t *botStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Bot, error) {
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
		return nil, fmt.Errorf("failed to find Bot: %w", err)
	}
	defer rows.Close()

	var results []*Bot
	for rows.Next() {
		model := &Bot{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Bot: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Bot based on the provided options.
func (t *botStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Bot, error) {
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

// Count counts Bot based on the provided options.
func (t *botStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count Bot: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Bot with pagination support.
func (t *botStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Bot, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Bot: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Bot: %w", err)
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the Bot for the given ID.
func (t *botStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Bot, error) {
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
	var model Bot
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan Bot: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
func (t *botStorage) Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *botStorage) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *botStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx).QueryContext(ctx, query, args...)
}
