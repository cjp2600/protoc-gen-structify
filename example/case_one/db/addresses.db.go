package db

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"gopkg.in/guregu/null.v4"
	"math"
	"strings"
	"time"
)

// addressStorage is a struct for the "addresses" table.
type addressStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// AddressCRUDOperations is an interface for managing the addresses table.
type AddressCRUDOperations interface {
	Create(ctx context.Context, model *Address, opts ...Option) (*string, error)
	Upsert(ctx context.Context, model *Address, updateFields []string, opts ...Option) (*string, error)
	BatchCreate(ctx context.Context, models []*Address, opts ...Option) ([]string, error)
	Update(ctx context.Context, id string, updateData *AddressUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*Address, error)
	GetIdField(ctx context.Context, id string, field string) (interface{}, error)
}

// AddressSearchOperations is an interface for searching the addresses table.
type AddressSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Address, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Address, error)
}

// AddressPaginationOperations is an interface for pagination operations.
type AddressPaginationOperations interface {
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Address, *Paginator, error)
}

// AddressRelationLoading is an interface for loading relations.
type AddressRelationLoading interface {
	LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error
}

// AddressAdvancedDeletion is an interface for advanced deletion operations.
type AddressAdvancedDeletion interface {
	DeleteMany(ctx context.Context, builders ...*QueryBuilder) error
}

// AddressRawQueryOperations is an interface for executing raw queries.
type AddressRawQueryOperations interface {
	Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error)
	QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row
	QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error)
}

// AddressStorage is a struct for the "addresses" table.
type AddressStorage interface {
	AddressCRUDOperations
	AddressSearchOperations
	AddressPaginationOperations
	AddressRelationLoading
	AddressAdvancedDeletion
	AddressRawQueryOperations
}

// NewAddressStorage returns a new addressStorage.
func NewAddressStorage(config *Config) (AddressStorage, error) {
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

	return &addressStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *addressStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *addressStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// TableName returns the table name.
func (t *addressStorage) TableName() string {
	return "addresses"
}

// Columns returns the columns for the table.
func (t *addressStorage) Columns() []string {
	return []string{
		"id", "street", "city", "state", "zip", "user_id", "created_at", "updated_at",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *addressStorage) DB(ctx context.Context, isWrite bool) QueryExecer {
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

// LoadUser loads the User relation.
func (t *addressStorage) LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("Address is nil")
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

// LoadBatchUser loads the User relation.
func (t *addressStorage) LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error {
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

// Address is a struct for the "addresses" table.
type Address struct {
	Id        string `db:"id"`
	Street    string `db:"street"`
	City      string `db:"city"`
	State     int32  `db:"state"`
	Zip       int64  `db:"zip"`
	User      *User
	UserId    string     `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
}

// TableName returns the table name.
func (t *Address) TableName() string {
	return "addresses"
}

// ScanRow scans a row into a Address.
func (t *Address) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.Street, &t.City, &t.State, &t.Zip, &t.UserId, &t.CreatedAt, &t.UpdatedAt)
}

// ScanRows scans a single row into the Address.
func (t *Address) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.Street,
		&t.City,
		&t.State,
		&t.Zip,
		&t.UserId,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
}

// AddressFilters is a struct that holds filters for Address.
type AddressFilters struct {
	Id     *string
	UserId *string
}

// AddressIdEq returns a condition that checks if the field equals the value.
func AddressIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// AddressUserIdEq returns a condition that checks if the field equals the value.
func AddressUserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "user_id", Value: value}
}

// AddressIdNotEq returns a condition that checks if the field equals the value.
func AddressIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// AddressUserIdNotEq returns a condition that checks if the field equals the value.
func AddressUserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "user_id", Value: value}
}

// AddressIdGT greaterThanCondition than condition.
func AddressIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// AddressUserIdGT greaterThanCondition than condition.
func AddressUserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "user_id", Value: value}
}

// AddressIdLT less than condition.
func AddressIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// AddressUserIdLT less than condition.
func AddressUserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "user_id", Value: value}
}

// AddressIdGTE greater than or equal condition.
func AddressIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// AddressUserIdGTE greater than or equal condition.
func AddressUserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "user_id", Value: value}
}

// AddressIdLTE less than or equal condition.
func AddressIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// AddressUserIdLTE less than or equal condition.
func AddressUserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "user_id", Value: value}
}

// AddressIdBetween between condition.
func AddressIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// AddressUserIdBetween between condition.
func AddressUserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "user_id", Min: min, Max: max}
}

// AddressIdILike iLike condition %
func AddressIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// AddressUserIdILike iLike condition %
func AddressUserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "user_id", Value: value}
}

// AddressIdLike like condition %
func AddressIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// AddressUserIdLike like condition %
func AddressUserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "user_id", Value: value}
}

// AddressIdNotLike not like condition
func AddressIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// AddressUserIdNotLike not like condition
func AddressUserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "user_id", Value: value}
}

// AddressIdIn condition
func AddressIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// AddressUserIdIn condition
func AddressUserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "user_id", Values: values}
}

// AddressIdNotIn not in condition
func AddressIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// AddressUserIdNotIn not in condition
func AddressUserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "user_id", Values: values}
}

// AddressIdOrderBy sorts the result in ascending order.
func AddressIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// AddressUserIdOrderBy sorts the result in ascending order.
func AddressUserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("user_id", asc)
}

// Create creates a new Address.
func (t *addressStorage) Create(ctx context.Context, model *Address, opts ...Option) (*string, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Insert("addresses").
		Columns(
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

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

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		if IsPgUniqueViolation(err) {
			return nil, fmt.Errorf("%w: %s", ErrRowAlreadyExist, PgPrettyErr(err).Error())
		}

		return nil, fmt.Errorf("failed to create Address: %w", err)
	}

	return &id, nil
}

// Upsert creates a new Address or updates existing one on conflict.
func (t *addressStorage) Upsert(ctx context.Context, model *Address, updateFields []string, opts ...Option) (*string, error) {
	if model == nil {
		return nil, fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	// Build INSERT query
	query := t.queryBuilder.Insert("addresses").
		Columns(
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		).
		Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
		)

	// Build UPDATE SET clause based on updateFields
	updateSet := make([]string, 0, len(updateFields))
	for _, field := range updateFields {
		if field == "street" {
			updateSet = append(updateSet, "street = EXCLUDED.street")
		}
		if field == "city" {
			updateSet = append(updateSet, "city = EXCLUDED.city")
		}
		if field == "state" {
			updateSet = append(updateSet, "state = EXCLUDED.state")
		}
		if field == "zip" {
			updateSet = append(updateSet, "zip = EXCLUDED.zip")
		}
		if field == "user_id" {
			updateSet = append(updateSet, "user_id = EXCLUDED.user_id")
		}
		if field == "created_at" {
			updateSet = append(updateSet, "created_at = EXCLUDED.created_at")
		}
		if field == "updated_at" {
			updateSet = append(updateSet, "updated_at = EXCLUDED.updated_at")
		}
	}

	// Note: You can manually add updated_at to updateFields if needed

	// Build the complete suffix with ON CONFLICT, UPDATE SET, and RETURNING in one string
	var suffixBuilder strings.Builder

	// Add ON CONFLICT clause
	if options.ignoreConflictField != "" {
		suffixBuilder.WriteString("ON CONFLICT (")
		suffixBuilder.WriteString(options.ignoreConflictField)
		suffixBuilder.WriteString(") DO UPDATE SET ")
	} else {
		suffixBuilder.WriteString("ON CONFLICT (id) DO UPDATE SET ")
	}

	// Add UPDATE SET fields
	if len(updateSet) > 0 {
		suffixBuilder.WriteString(strings.Join(updateSet, ", "))
	}

	// Add RETURNING clause
	suffixBuilder.WriteString(" RETURNING \"id\"")

	// Add the complete suffix once
	query = query.Suffix(suffixBuilder.String())

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	var id string
	err = t.DB(ctx, true).QueryRowContext(ctx, sqlQuery, args...).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert Address: %w", err)
	}

	return &id, nil
}

// BatchCreate creates multiple Address records in a single batch.
func (t *addressStorage) BatchCreate(ctx context.Context, models []*Address, opts ...Option) ([]string, error) {
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
			"street",
			"city",
			"state",
			"zip",
			"user_id",
			"created_at",
			"updated_at",
		)

	for _, model := range models {
		if model == nil {
			return nil, fmt.Errorf("one of the models is nil")
		}

		query = query.Values(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

// AddressUpdate is used to update an existing Address.
type AddressUpdate struct {
	// Use regular pointer types for non-optional fields
	Street *string
	// Use regular pointer types for non-optional fields
	City *string
	// Use regular pointer types for non-optional fields
	State *int32
	// Use regular pointer types for non-optional fields
	Zip *int64
	// Use regular pointer types for non-optional fields
	UserId *string
	// Use regular pointer types for non-optional fields
	CreatedAt *time.Time
	// Use null types for optional fields
	UpdatedAt null.Time
}

// Update updates an existing Address based on non-nil fields.
func (t *addressStorage) Update(ctx context.Context, id string, updateData *AddressUpdate) error {
	if updateData == nil {
		return fmt.Errorf("update data is nil")
	}

	query := t.queryBuilder.Update("addresses")
	// Handle fields that are not optional using a nil check
	if updateData.Street != nil {
		query = query.Set("street", *updateData.Street) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.City != nil {
		query = query.Set("city", *updateData.City) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.State != nil {
		query = query.Set("state", *updateData.State) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.Zip != nil {
		query = query.Set("zip", *updateData.Zip) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.UserId != nil {
		query = query.Set("user_id", *updateData.UserId) // Dereference pointer value
	}
	// Handle fields that are not optional using a nil check
	if updateData.CreatedAt != nil {
		query = query.Set("created_at", *updateData.CreatedAt) // Dereference pointer value
	}
	// Handle fields that are optional and can be explicitly set to NULL
	if updateData.UpdatedAt.Valid {
		// Handle null.Time specifically
		if updateData.UpdatedAt.Time.IsZero() {
			query = query.Set("updated_at", nil) // Explicitly set NULL if time is zero
		} else {
			query = query.Set("updated_at", updateData.UpdatedAt.Time)
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
		return fmt.Errorf("failed to update Address: %w", err)
	}

	return nil
}

// DeleteById - deletes a Address by its id.
func (t *addressStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("addresses").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	_, err = t.DB(ctx, true).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete Address: %w", err)
	}

	return nil
}

// DeleteMany removes entries from the addresses table using the provided filters
func (t *addressStorage) DeleteMany(ctx context.Context, builders ...*QueryBuilder) error {
	// build query
	query := t.queryBuilder.Delete("addresses")

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
		return fmt.Errorf("failed to delete addresses: %w", err)
	}

	return nil
}

// FindById retrieves a Address by its id.
func (t *addressStorage) FindById(ctx context.Context, id string, opts ...Option) (*Address, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(AddressIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, fmt.Errorf("find one Address: %w", err)
	}

	return model, nil
}

// GetIdField retrieves a specific field value by id.
func (t *addressStorage) GetIdField(ctx context.Context, id string, field string) (interface{}, error) {
	query := t.queryBuilder.Select(field).From(t.TableName()).Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	row := t.DB(ctx, false).QueryRowContext(ctx, sqlQuery, args...)
	var value interface{}
	if err := row.Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan field value: %w", err)
	}

	return value, nil
}

// FindMany finds multiple Address based on the provided options.
func (t *addressStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*Address, error) {
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

	var results []*Address
	for rows.Next() {
		model := &Address{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan Address: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single Address based on the provided options.
func (t *addressStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne Address: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Count counts Address based on the provided options.
func (t *addressStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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

// FindManyWithPagination finds multiple Address with pagination support.
func (t *addressStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*Address, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count Address: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find Address: %w", err)
	}

	return records, paginator, nil
}

// SelectForUpdate lock locks the Address for the given ID.
func (t *addressStorage) SelectForUpdate(ctx context.Context, builders ...*QueryBuilder) (*Address, error) {
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
	var model Address
	if err := model.ScanRow(row); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRowNotFound
		}
		return nil, fmt.Errorf("failed to scan Address: %w", err)
	}

	return &model, nil
}

// Query executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *addressStorage) Query(ctx context.Context, isWrite bool, query string, args ...interface{}) (sql.Result, error) {
	return t.DB(ctx, isWrite).ExecContext(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *addressStorage) QueryRow(ctx context.Context, isWrite bool, query string, args ...interface{}) *sql.Row {
	return t.DB(ctx, isWrite).QueryRowContext(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
// isWrite is used to determine if the query is a write operation.
func (t *addressStorage) QueryRows(ctx context.Context, isWrite bool, query string, args ...interface{}) (*sql.Rows, error) {
	return t.DB(ctx, isWrite).QueryContext(ctx, query, args...)
}
