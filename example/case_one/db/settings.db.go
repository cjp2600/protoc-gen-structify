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

// settingStorage is a struct for the "settings" table.
type settingStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

// SettingTableManager is an interface for managing the settings table.
type SettingTableManager interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
}

// SettingCRUDOperations is an interface for managing the settings table.
type SettingCRUDOperations interface {
	Create(ctx context.Context, model *Setting, opts ...Option) (*int32, error)
	Update(ctx context.Context, id int32, updateData *SettingUpdate) error
	DeleteById(ctx context.Context, id int32, opts ...Option) error
	FindById(ctx context.Context, id int32, opts ...Option) (*Setting, error)
}

// SettingSearchOperations is an interface for searching the settings table.
type SettingSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Setting, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Setting, error)
}

// SettingPaginationOperations is an interface for pagination operations.
type SettingPaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Setting, *Paginator, error)
}

// SettingRelationLoading is an interface for loading relations.
type SettingRelationLoading interface {
	LoadUser(ctx context.Context, model *Setting, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Setting, builders ...*QueryBuilder) error
}

// SettingAdvancedDeletion is an interface for advanced deletion operations.
type SettingAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// SettingRawQueryOperations is an interface for executing raw queries.
type SettingRawQueryOperations interface {
	Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// SettingStorage is a struct for the "settings" table.
type SettingStorage interface {
	SettingTableManager
	SettingCRUDOperations
	SettingSearchOperations
	SettingPaginationOperations
	SettingRelationLoading
	SettingAdvancedDeletion
	SettingRawQueryOperations
}

// NewSettingStorage returns a new settingStorage.
func NewSettingStorage(db *sql.DB) SettingStorage {
	return &settingStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *settingStorage) TableName() string {
	return "settings"
}

// Columns returns the columns for the table.
func (t *settingStorage) Columns() []string {
	return []string{
		"id", "name", "value", "user_id",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *settingStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *settingStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		-- Table: settings
		CREATE TABLE IF NOT EXISTS settings (
		id  SERIAL PRIMARY KEY,
		name TEXT NOT NULL,
		value TEXT,
		user_id UUID NOT NULL);
		-- Other entities
		CREATE UNIQUE INDEX IF NOT EXISTS settings_user_id_unique_idx ON settings USING btree (user_id);
		CREATE INDEX IF NOT EXISTS settings_name_idx ON settings USING btree (name);
		CREATE INDEX IF NOT EXISTS settings_user_id_idx ON settings USING btree (user_id);
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *settingStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS settings;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *settingStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE settings;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *settingStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// LoadUser loads the User relation.
func (t *settingStorage) LoadUser(ctx context.Context, model *Setting, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Setting is nil")
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
func (t *settingStorage) LoadBatchUser(ctx context.Context, items []*Setting, builders ...*QueryBuilder) error {
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

// Setting is a struct for the "settings" table.
type Setting struct {
	Id     int32  `db:"id"`
	Name   string `db:"name"`
	Value  string `db:"value"`
	User   *User
	UserId string `db:"user_id"`
}

// TableName returns the table name.
func (t *Setting) TableName() string {
	return "settings"
}

// ScanRow scans a row into a Setting.
func (t *Setting) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.Name, &t.Value, &t.UserId)
}

// ScanRows scans a single row into the Setting.
func (t *Setting) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.Name,
		&t.Value,
		&t.UserId,
	)
}

// SettingFilters is a struct that holds filters for Setting.
type SettingFilters struct {
	Id     *int32
	UserId *string
}

// SettingIdEq returns a condition that checks if the field equals the value.
func SettingIdEq(value int32) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// SettingUserIdEq returns a condition that checks if the field equals the value.
func SettingUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// SettingIdNotEq returns a condition that checks if the field equals the value.
func SettingIdNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// SettingUserIdNotEq returns a condition that checks if the field equals the value.
func SettingUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// SettingIdGT greaterThanCondition than condition.
func SettingIdGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// SettingUserIdGT greaterThanCondition than condition.
func SettingUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// SettingIdLT less than condition.
func SettingIdLT(value int32) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// SettingUserIdLT less than condition.
func SettingUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// SettingIdGTE greater than or equal condition.
func SettingIdGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// SettingUserIdGTE greater than or equal condition.
func SettingUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// SettingIdLTE less than or equal condition.
func SettingIdLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// SettingUserIdLTE less than or equal condition.
func SettingUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// SettingIdBetween between condition.
func SettingIdBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// SettingUserIdBetween between condition.
func SettingUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// SettingUserIdILike iLike condition %
func SettingUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// SettingUserIdLike like condition %
func SettingUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// SettingUserIdNotLike not like condition
func SettingUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// SettingIdIn condition
func SettingIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// SettingUserIdIn condition
func SettingUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// SettingIdNotIn not in condition
func SettingIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// SettingUserIdNotIn not in condition
func SettingUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// SettingIdOrderBy sorts the result in ascending order.
func SettingIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// SettingUserIdOrderBy sorts the result in ascending order.
func SettingUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// Create creates a new Setting.
func (t *settingStorage) Create(ctx context.Context, model *Setting, opts ...Option) (*int32, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("settings").
		Columns(
			"name",
			"value",
			"user_id",
		).
		Values(
			model.Name,
			model.Value,
			model.UserId,
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

		return nil, fmt.Errorf("failed to create Setting: %w", err)
	}

	return &id, nil
}

// SettingUpdate is used to update an existing Setting.
type SettingUpdate struct {
	// Use regular pointer types for non-optional fields
	Name *string
	// Use regular pointer types for non-optional fields
	Value *string
	// Use regular pointer types for non-optional fields
	UserId *string
}

// Update updates an existing Setting based on non-nil fields.
func (t *settingStorage) Update(ctx context.Context, id int32, updateData *SettingUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("settings")
	// Handle fields that are not optional using a nil check
	if updateData.Name != nil {
		query = query.Set("name", *updateData.Name) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Value != nil {
		query = query.Set("value", *updateData.Value) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.UserId != nil {
		query = query.Set("user_id", *updateData.UserId) // Dereference pointer value
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update Setting: %w", err)
	}

	return nil
}

// DeleteById - deletes a Setting by its id.
func (t *settingStorage) DeleteById(ctx context.Context, id int32, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("settings").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Setting: %w", err)
	}

	return nil
}

// DeleteMany removes entries from the settings table using the provided filters
func (t *settingStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("settings")

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

// FindById retrieves a Setting by its id.
func (t *settingStorage) FindById(ctx context.Context, id int32, opts ...Option) (*Setting, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(SettingIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one Setting: ")
	}

	return model, nil
}

// FindMany finds multiple Setting based on the provided options.
func (t *settingStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Setting, error) {
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
		return nil, fmt.Errorf("failed to find Setting: %w", err)
	}
	defer rows.Close()

	var results []*Setting
	for rows.Next() {
		model := &Setting{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Setting: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single Setting based on the provided options.
func (t *settingStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error) {
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

// Count counts Setting based on the provided options.
func (t *settingStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count Setting: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple Setting with pagination support.
func (t *settingStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Setting, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Setting: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Setting: %w", err)
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the Setting for the given ID.
func (t *settingStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Setting, error) {
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
	var model Setting
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan Setting: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
func (t *settingStorage) Query(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *settingStorage) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *settingStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx).QueryContext(ctx, query, args...)
}
