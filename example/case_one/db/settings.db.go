package db

import (
	"context"
	"database/sql"
	sq "github.com/Masterminds/squirrel"
	_ "github.com/lib/pq"
	"github.com/pkg/errors"
	"math"
)

// settingStorage is a struct for the "settings" table.
type settingStorage struct {
	db           *DB                     // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

// SettingCRUDOperations is an interface for managing the settings table.
type SettingCRUDOperations interface {
	Create(ctx context.Context, model *Setting, opts ...Option) (*int32, error)
	BatchCreate(ctx context.Context, models []*Setting, opts ...Option) ([]string, error)
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
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// SettingStorage is a struct for the "settings" table.
type SettingStorage interface {
	SettingCRUDOperations
	SettingSearchOperations
	SettingPaginationOperations
	SettingRelationLoading
	SettingAdvancedDeletion
	SettingRawQueryOperations
}

// NewSettingStorage returns a new settingStorage.
func NewSettingStorage(db *DB) SettingStorage {
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

// DB returns the underlying DB. This is useful for doing transactions.
func (t *settingStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
	var db QueryExecer

	// Check if there is an active transaction in the context.
	if tx, ok := TxFromContext(ctx); ok {
		return tx
	}

	// Use the appropriate connection based on the operation type.
	if isWrite {
		db = t.db.DBWrite
	} else {
		db = t.db.DBRead
	}

	return db
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
		return errors.Wrap(err, "failed to find one UserStorage")
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
		return errors.Wrap(err, "failed to find many UserStorage")
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
		return nil, errors.Wrap(err, "failed to build query")
	}

	var id int32
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, errors.Wrap(err, "failed to create Setting")
	}

	return &id, nil
}

// BatchCreate creates multiple Setting records in a single batch.
func (t *settingStorage) BatchCreate(ctx context.Context, models []*Setting, opts ...Option) ([]string, error) {
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
			"name",
			"value",
			"user_id",
		)

	for _, model := range models {
		if model == nil {
			return nil, errors.New("one of the models is nil")
		}
		query = query.Values(
			model.Name,
			model.Value,
			model.UserId,
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

	rows, err := t.DB(ctx, true).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, errors.Wrap(ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}
		return nil, errors.Wrap(err, "failed to execute bulk insert")
	}
	defer rows.Close()

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
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to update Setting")
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
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete Setting")
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
		return errors.Wrap(err, "failed to build query")
	}

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to delete settings")
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

	rows, err := t.DB(ctx, false).QueryContext(ctx, sqlQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var results []*Setting
	for rows.Next() {
		model := &Setting{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan Setting")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single Setting based on the provided options.
func (t *settingStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Setting, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne Setting")
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

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "failed to build query")
	}

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, errors.Wrap(err, "failed to scan count")
	}

	return count, nil
}

// FindManyWithPagination finds multiple Setting with pagination support.
func (t *settingStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Setting, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to count Setting")
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
		return nil, nil, errors.Wrap(err, "failed to find Setting")
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

		// apply custom filters
		query = builder.ApplyCustomFilters(query)
	}

	// execute query
	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build query")
	}

	row := t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...)
	var model Setting
	if err := model.ScanRow(row); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRowNotFound
		}
		return nil, errors.Wrap(err, "failed to scan Setting")
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *settingStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *settingStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *settingStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
