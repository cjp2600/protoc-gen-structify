package db

import (
	"context"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"github.com/pkg/errors"
	"time"
)

// addressStorage is a struct for the "addresses" table.
type addressStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// AddressCRUDOperations is an interface for managing the addresses table.
type AddressCRUDOperations interface {
	Create(ctx context.Context, model *Address, opts ...Option) error
	AsyncCreate(ctx context.Context, model *Address, opts ...Option) error
	BatchCreate(ctx context.Context, models []*Address, opts ...Option) error
}

// AddressSearchOperations is an interface for searching the addresses table.
type AddressSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*Address, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error)
}

type AddressSettings interface {
	Conn() driver.Conn
	SetConfig(config *Config) AddressStorage
	SetQueryBuilder(builder sq.StatementBuilderType) AddressStorage
}

// AddressRelationLoading is an interface for loading relations.
type AddressRelationLoading interface {
	LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error
	LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error
}

// AddressRawQueryOperations is an interface for executing raw queries.
type AddressRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// AddressStorage is a struct for the "addresses" table.
type AddressStorage interface {
	AddressCRUDOperations
	AddressSearchOperations
	AddressRelationLoading
	AddressRawQueryOperations
	AddressSettings
}

// NewAddressStorage returns a new addressStorage.
func NewAddressStorage(config *Config) (AddressStorage, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}
	if config.DB == nil {
		return nil, errors.New("config.DB connection is nil")
	}

	return &addressStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
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
func (t *addressStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *addressStorage) SetConfig(config *Config) AddressStorage {
	t.config = config
	return t
}

func (t *addressStorage) SetQueryBuilder(builder sq.StatementBuilderType) AddressStorage {
	t.queryBuilder = builder
	return t
}

// LoadUser loads the User relation.
func (t *addressStorage) LoadUser(ctx context.Context, model *Address, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "Address is nil")
	}

	// NewUserStorage creates a new UserStorage.
	s, err := NewUserStorage(t.config)
	if err != nil {
		return errors.Wrap(err, "failed to create UserStorage")
	}
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
func (t *addressStorage) LoadBatchUser(ctx context.Context, items []*Address, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.UserId)
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
		if v, ok := resultMap[item.UserId]; ok {
			item.User = v
		}
	}

	return nil
}

// Address is a struct for the "addresses" table.
type Address struct {
	Id        string
	Street    string
	City      string
	State     int32
	Zip       int64
	User      *User
	UserId    string
	CreatedAt time.Time
	UpdatedAt *time.Time
}

// TableName returns the table name.
func (t *Address) TableName() string {
	return "addresses"
}

// ScanRow scans a row into a Address.
func (t *Address) ScanRow(row driver.Row) error {
	return row.Scan(
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

// ScanRows scans multiple rows into the struct Address.
func (t *Address) ScanRows(rows driver.Rows) error {
	for rows.Next() {
		if err := rows.Scan(
			&t.Id,
			&t.Street,
			&t.City,
			&t.State,
			&t.Zip,
			&t.UserId,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return err
		}
	}
	return rows.Err()
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

// AsyncCreate asynchronously inserts a new Address.
func (t *addressStorage) AsyncCreate(ctx context.Context, model *Address, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
	}

	// Set default options
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, false, args...); err != nil {
		return errors.Wrap(err, "failed to asynchronously create Address")
	}

	return nil
}

// Create creates a new Address.
func (t *addressStorage) Create(ctx context.Context, model *Address, opts ...Option) error {
	if model == nil {
		return errors.New("model is nil")
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

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "failed to build query")
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return errors.Wrap(err, "failed to create Address")
	}

	return nil
}

// BatchCreate creates multiple Address records in a single batch.
func (t *addressStorage) BatchCreate(ctx context.Context, models []*Address, opts ...Option) error {
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

	batch, err := t.DB().PrepareBatch(ctx, "INSERT INTO "+t.TableName())
	if err != nil {
		return errors.Wrap(err, "failed to prepare batch")
	}

	for _, model := range models {
		if model == nil {
			return errors.New("one of the models is nil")
		}

		err := batch.Append(
			model.Street,
			model.City,
			model.State,
			model.Zip,
			model.UserId,
			model.CreatedAt,
			nullValue(model.UpdatedAt),
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

	var results []*Address
	for rows.Next() {
		model := &Address{}
		if err := model.ScanRows(rows); err != nil {
			return nil, errors.Wrap(err, "failed to scan Address")
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate over rows")
	}

	return results, nil
}

// FindOne finds a single Address based on the provided options.
func (t *addressStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*Address, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to findOne Address")
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *addressStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *addressStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *addressStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *addressStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *addressStorage) Conn() driver.Conn {
	return t.DB()
}
