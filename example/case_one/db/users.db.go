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

// userStorage is a struct for the "users" table.
type userStorage struct {
	db           *sql.DB                 // The database connection.
	queryBuilder sq.StatementBuilderType // queryBuilder is used to build queries.
}

type UserStorage interface {
	CreateTable(ctx context.Context) error
	DropTable(ctx context.Context) error
	TruncateTable(ctx context.Context) error
	UpgradeTable(ctx context.Context) error
	Create(ctx context.Context, model *User, opts ...Option) (*string, error)
	Update(ctx context.Context, id string, updateData *UserUpdate) error
	DeleteById(ctx context.Context, id string, opts ...Option) error
	FindById(ctx context.Context, id string, opts ...Option) (*User, error)
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*User, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error)
	Count(ctx context.Context, builders ...*QueryBuilder) (int64, error)
	LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error
	FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*User, *Paginator, error)
	LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadPosts(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadBatchDevice(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error
}

// NewUserStorage returns a new userStorage.
func NewUserStorage(db *sql.DB) UserStorage {
	return &userStorage{
		db:           db,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

// TableName returns the table name.
func (t *userStorage) TableName() string {
	return "users"
}

// Columns returns the columns for the table.
func (t *userStorage) Columns() []string {
	return []string{
		"id", "name", "age", "email", "last_name", "created_at", "updated_at", "notification_settings", "phones", "balls", "numrs", "comments",
	}
}

// DB returns the underlying sql.DB. This is useful for doing transactions.
func (t *userStorage) DB(ctx context.Context) QueryExecer {
	var db QueryExecer = t.db
	if tx, ok := TxFromContext(ctx); ok {
		db = tx
	}

	return db
}

// createTable creates the table.
func (t *userStorage) CreateTable(ctx context.Context) error {
	sqlQuery := `
		CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
		-- Table: users
		CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY NOT NULL DEFAULT uuid_generate_v4(),
		name TEXT NOT NULL,
		age INTEGER NOT NULL,
		email TEXT NOT NULL,
		last_name TEXT,
		created_at TIMESTAMP NOT NULL DEFAULT now(),
		updated_at TIMESTAMP,
		notification_settings JSONB,
		phones JSONB,
		balls JSONB,
		numrs JSONB,
		comments JSONB);
		-- Other entities
		COMMENT ON TABLE users IS 'This is a comment of User';
		CREATE UNIQUE INDEX IF NOT EXISTS users_email_unique_idx ON users USING btree (email);
		CREATE UNIQUE INDEX IF NOT EXISTS users_unique_idx_name_email ON users USING btree (
            name, 
            email
    	);
		CREATE INDEX IF NOT EXISTS users_name_idx ON users USING btree (name);
		-- Foreign keys for devices
		ALTER TABLE users
		ADD FOREIGN KEY (id) REFERENCES devices(user_id)
		ON DELETE CASCADE;
		-- Foreign keys for settings
		ALTER TABLE users
		ADD FOREIGN KEY (id) REFERENCES settings(user_id)
		ON DELETE CASCADE;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// DropTable drops the table.
func (t *userStorage) DropTable(ctx context.Context) error {
	sqlQuery := `
		DROP TABLE IF EXISTS users;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// TruncateTable truncates the table.
func (t *userStorage) TruncateTable(ctx context.Context) error {
	sqlQuery := `
		TRUNCATE TABLE users;
	`

	_, err := t.db.ExecContext(ctx, sqlQuery)
	return err
}

// UpgradeTable upgrades the table.
// todo: delete this method
func (t *userStorage) UpgradeTable(ctx context.Context) error {
	return nil
}

// LoadDevice loads the Device relation.
func (t *userStorage) LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s := NewDeviceStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(DeviceUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find DeviceStorage: %w", err)
	}

	model.Device = relationModel
	return nil
}

// LoadSettings loads the Settings relation.
func (t *userStorage) LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewSettingStorage creates a new SettingStorage.
	s := NewSettingStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(SettingUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find SettingStorage: %w", err)
	}

	model.Settings = relationModel
	return nil
}

// LoadAddresses loads the Addresses relation.
func (t *userStorage) LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewAddressStorage creates a new AddressStorage.
	s := NewAddressStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(AddressUserIdEq(model.Id)))
	relationModels, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many AddressStorage: %w", err)
	}

	model.Addresses = relationModels
	return nil
}

// LoadPosts loads the Posts relation.
func (t *userStorage) LoadPosts(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return errors.Wrap(ErrModelIsNil, "User is nil")
	}

	// NewPostStorage creates a new PostStorage.
	s := NewPostStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(PostAuthorIdEq(model.Id)))
	relationModels, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many PostStorage: %w", err)
	}

	model.Posts = relationModels
	return nil
}

// LoadBatchDevice loads the Device relation.
func (t *userStorage) LoadBatchDevice(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.Id
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s := NewDeviceStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(DeviceUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many DeviceStorage: %w", err)
	}
	resultMap := make(map[interface{}]*Device)
	for _, result := range results {
		resultMap[result.UserId] = result
	}

	// Assign Device to items
	for _, item := range items {
		if v, ok := resultMap[item.Id]; ok {
			item.Device = v
		}
	}

	return nil
}

// LoadBatchSettings loads the Settings relation.
func (t *userStorage) LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.Id
	}

	// NewSettingStorage creates a new SettingStorage.
	s := NewSettingStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(SettingUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many SettingStorage: %w", err)
	}
	resultMap := make(map[interface{}]*Setting)
	for _, result := range results {
		resultMap[result.UserId] = result
	}

	// Assign Setting to items
	for _, item := range items {
		if v, ok := resultMap[item.Id]; ok {
			item.Settings = v
		}
	}

	return nil
}

// LoadBatchAddresses loads the Addresses relation.
func (t *userStorage) LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.Id
	}

	// NewAddressStorage creates a new AddressStorage.
	s := NewAddressStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(AddressUserIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many AddressStorage: %w", err)
	}
	resultMap := make(map[interface{}][]*Address)
	for _, result := range results {
		resultMap[result.UserId] = append(resultMap[result.UserId], result)
	}

	// Assign Address to items
	for _, item := range items {
		if v, ok := resultMap[item.Id]; ok {
			item.Addresses = v
		}
	}

	return nil
}

// LoadBatchPosts loads the Posts relation.
func (t *userStorage) LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, len(items))
	for i, item := range items {
		requestItems[i] = item.Id
	}

	// NewPostStorage creates a new PostStorage.
	s := NewPostStorage(t.db)

	// Add the filter for the relation
	builders = append(builders, FilterBuilder(PostAuthorIdIn(requestItems...)))

	results, err := s.FindMany(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find many PostStorage: %w", err)
	}
	resultMap := make(map[interface{}][]*Post)
	for _, result := range results {
		resultMap[result.AuthorId] = append(resultMap[result.AuthorId], result)
	}

	// Assign Post to items
	for _, item := range items {
		if v, ok := resultMap[item.Id]; ok {
			item.Posts = v
		}
	}

	return nil
}

// User is a struct for the "users" table.
type User struct {
	Id                   string  `db:"id"`
	Name                 string  `db:"name"`
	Age                  int32   `db:"age"`
	Email                string  `db:"email"`
	LastName             *string `db:"last_name"`
	Device               *Device
	Settings             *Setting
	Addresses            []*Address
	Posts                []*Post
	CreatedAt            time.Time                `db:"created_at"`
	UpdatedAt            *time.Time               `db:"updated_at"`
	NotificationSettings *UserNotificationSetting `db:"notification_settings"`
	Phones               UserPhonesRepeated       `db:"phones"`
	Balls                UserBallsRepeated        `db:"balls"`
	Numrs                UserNumrsRepeated        `db:"numrs"`
	Comments             UserCommentsRepeated     `db:"comments"`
}

// TableName returns the table name.
func (t *User) TableName() string {
	return "users"
}

// ScanRow scans a row into a User.
func (t *User) ScanRow(r *sql.Row) error {
	return r.Scan(&t.Id, &t.Name, &t.Age, &t.Email, &t.LastName, &t.CreatedAt, &t.UpdatedAt, &t.NotificationSettings, &t.Phones, &t.Balls, &t.Numrs, &t.Comments)
}

// ScanRows scans a single row into the User.
func (t *User) ScanRows(r *sql.Rows) error {
	return r.Scan(
		&t.Id,
		&t.Name,
		&t.Age,
		&t.Email,
		&t.LastName,
		&t.CreatedAt,
		&t.UpdatedAt,
		&t.NotificationSettings,
		&t.Phones,
		&t.Balls,
		&t.Numrs,
		&t.Comments,
	)
}

// UserFilters is a struct that holds filters for User.
type UserFilters struct {
	Id    *string
	Age   *int32
	Email *string
}

// UserIdEq returns a condition that checks if the field equals the value.
func UserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// UserAgeEq returns a condition that checks if the field equals the value.
func UserAgeEq(value int32) FilterApplier {
	return EqualsCondition{Field: "age", Value: value}
}

// UserEmailEq returns a condition that checks if the field equals the value.
func UserEmailEq(value string) FilterApplier {
	return EqualsCondition{Field: "email", Value: value}
}

// UserIdNotEq returns a condition that checks if the field equals the value.
func UserIdNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "id", Value: value}
}

// UserAgeNotEq returns a condition that checks if the field equals the value.
func UserAgeNotEq(value int32) FilterApplier {
	return NotEqualsCondition{Field: "age", Value: value}
}

// UserEmailNotEq returns a condition that checks if the field equals the value.
func UserEmailNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "email", Value: value}
}

// UserIdGT greaterThanCondition than condition.
func UserIdGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "id", Value: value}
}

// UserAgeGT greaterThanCondition than condition.
func UserAgeGT(value int32) FilterApplier {
	return GreaterThanCondition{Field: "age", Value: value}
}

// UserEmailGT greaterThanCondition than condition.
func UserEmailGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "email", Value: value}
}

// UserIdLT less than condition.
func UserIdLT(value string) FilterApplier {
	return LessThanCondition{Field: "id", Value: value}
}

// UserAgeLT less than condition.
func UserAgeLT(value int32) FilterApplier {
	return LessThanCondition{Field: "age", Value: value}
}

// UserEmailLT less than condition.
func UserEmailLT(value string) FilterApplier {
	return LessThanCondition{Field: "email", Value: value}
}

// UserIdGTE greater than or equal condition.
func UserIdGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "id", Value: value}
}

// UserAgeGTE greater than or equal condition.
func UserAgeGTE(value int32) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "age", Value: value}
}

// UserEmailGTE greater than or equal condition.
func UserEmailGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "email", Value: value}
}

// UserIdLTE less than or equal condition.
func UserIdLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "id", Value: value}
}

// UserAgeLTE less than or equal condition.
func UserAgeLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "age", Value: value}
}

// UserEmailLTE less than or equal condition.
func UserEmailLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "email", Value: value}
}

// UserIdLike like condition %
func UserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// UserAgeLike like condition %
func UserAgeLike(value int32) FilterApplier {
	return LikeCondition{Field: "age", Value: value}
}

// UserEmailLike like condition %
func UserEmailLike(value string) FilterApplier {
	return LikeCondition{Field: "email", Value: value}
}

// UserIdNotLike not like condition
func UserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// UserAgeNotLike not like condition
func UserAgeNotLike(value int32) FilterApplier {
	return NotLikeCondition{Field: "age", Value: value}
}

// UserEmailNotLike not like condition
func UserEmailNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "email", Value: value}
}

// UserIdIsNull is null condition
func UserIdIsNull() FilterApplier {
	return IsNullCondition{Field: "id"}
}

// UserAgeIsNull is null condition
func UserAgeIsNull() FilterApplier {
	return IsNullCondition{Field: "age"}
}

// UserEmailIsNull is null condition
func UserEmailIsNull() FilterApplier {
	return IsNullCondition{Field: "email"}
}

// UserIdIsNotNull is not null condition
func UserIdIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "id"}
}

// UserAgeIsNotNull is not null condition
func UserAgeIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "age"}
}

// UserEmailIsNotNull is not null condition
func UserEmailIsNotNull() FilterApplier {
	return IsNotNullCondition{Field: "email"}
}

// UserIdIn condition
func UserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// UserAgeIn condition
func UserAgeIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "age", Values: values}
}

// UserEmailIn condition
func UserEmailIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "email", Values: values}
}

// UserIdNotIn not in condition
func UserIdNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "id", Values: values}
}

// UserAgeNotIn not in condition
func UserAgeNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "age", Values: values}
}

// UserEmailNotIn not in condition
func UserEmailNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "email", Values: values}
}

// UserIdOrderBy sorts the result in ascending order.
func UserIdOrderBy(asc bool) FilterApplier {
	return OrderBy("id", asc)
}

// UserAgeOrderBy sorts the result in ascending order.
func UserAgeOrderBy(asc bool) FilterApplier {
	return OrderBy("age", asc)
}

// UserEmailOrderBy sorts the result in ascending order.
func UserEmailOrderBy(asc bool) FilterApplier {
	return OrderBy("email", asc)
}

// Create creates a new User.
func (t *userStorage) Create(ctx context.Context, model *User, opts ...Option) (*string, error) {
	if model == nil {
		return nil, errors.New("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	// get value of phones
	phones, err := model.Phones.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get phones value: %w", err)
	}
	// get value of balls
	balls, err := model.Balls.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get balls value: %w", err)
	}
	// get value of numrs
	numrs, err := model.Numrs.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get numrs value: %w", err)
	}
	// get value of comments
	comments, err := model.Comments.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to get comments value: %w", err)
	}

	query := t.queryBuilder.Insert("users").
		Columns(
			"name",
			"age",
			"email",
			"last_name",
			"created_at",
			"updated_at",
			"notification_settings",
			"phones",
			"balls",
			"numrs",
			"comments",
		).
		Values(
			model.Name,
			model.Age,
			model.Email,
			model.LastName,
			model.CreatedAt,
			model.UpdatedAt,
			model.NotificationSettings,
			phones,
			balls,
			numrs,
			comments,
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

		return nil, fmt.Errorf("failed to create User: %w", err)
	}

	if options.relations && model.Device != nil {
		s := NewDeviceStorage(t.db)
		model.Device.UserId = id
		err := s.Create(ctx, model.Device)
		if err != nil {
			return nil, fmt.Errorf("failed to create Device: %w", err)
		}
	}
	if options.relations && model.Settings != nil {
		s := NewSettingStorage(t.db)
		model.Settings.UserId = id
		_, err := s.Create(ctx, model.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to create Settings: %w", err)
		}
	}
	if options.relations && model.Addresses != nil {
		for _, item := range model.Addresses {
			item.UserId = id
			s := NewAddressStorage(t.db)
			_, err := s.Create(ctx, item)
			if err != nil {
				return nil, fmt.Errorf("failed to create Addresses: %w", err)
			}
		}
	}

	return &id, nil
}

// UserUpdate is used to update an existing User.
type UserUpdate struct {
	Name                 *string
	Age                  *int32
	Email                *string
	LastName             *string
	CreatedAt            *time.Time
	UpdatedAt            *time.Time
	NotificationSettings *UserNotificationSetting
	Phones               *UserPhonesRepeated
	Balls                *UserBallsRepeated
	Numrs                *UserNumrsRepeated
	Comments             *UserCommentsRepeated
}

// Update updates an existing User based on non-nil fields.
func (t *userStorage) Update(ctx context.Context, id string, updateData *UserUpdate) error {
	if updateData == nil {
		return errors.New("update data is nil")
	}

	query := t.queryBuilder.Update("users")
	if updateData.Name != nil {
		query = query.Set("name", updateData.Name)
	}
	if updateData.Age != nil {
		query = query.Set("age", updateData.Age)
	}
	if updateData.Email != nil {
		query = query.Set("email", updateData.Email)
	}
	if updateData.LastName != nil {
		query = query.Set("last_name", updateData.LastName)
	}
	if updateData.CreatedAt != nil {
		query = query.Set("created_at", updateData.CreatedAt)
	}
	if updateData.UpdatedAt != nil {
		query = query.Set("updated_at", updateData.UpdatedAt)
	}
	if updateData.NotificationSettings != nil {
		query = query.Set("notification_settings", updateData.NotificationSettings)
	}
	if updateData.Phones != nil {
		value, err := updateData.Phones.Value()
		if err != nil {
			return fmt.Errorf("failed to get phones value: %w", err)
		}
		query = query.Set("phones", value)
	}
	if updateData.Balls != nil {
		value, err := updateData.Balls.Value()
		if err != nil {
			return fmt.Errorf("failed to get balls value: %w", err)
		}
		query = query.Set("balls", value)
	}
	if updateData.Numrs != nil {
		value, err := updateData.Numrs.Value()
		if err != nil {
			return fmt.Errorf("failed to get numrs value: %w", err)
		}
		query = query.Set("numrs", value)
	}
	if updateData.Comments != nil {
		value, err := updateData.Comments.Value()
		if err != nil {
			return fmt.Errorf("failed to get comments value: %w", err)
		}
		query = query.Set("comments", value)
	}

	query = query.Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to update User: %w", err)
	}

	return nil
}

// DeleteById - deletes a User by its id.
func (t *userStorage) DeleteById(ctx context.Context, id string, opts ...Option) error {
	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}

	query := t.queryBuilder.Delete("users").Where("id = ?", id)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}

	_, err = t.DB(ctx).ExecContext(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete User: %w", err)
	}

	return nil
}

// FindById retrieves a User by its id.
func (t *userStorage) FindById(ctx context.Context, id string, opts ...Option) (*User, error) {
	builder := NewQueryBuilder()
	{
		builder.WithFilter(UserIdEq(id))
		builder.WithOptions(opts...)
	}

	// Use FindOne to get a single result
	model, err := t.FindOne(ctx, builder)
	if err != nil {
		return nil, errors.Wrap(err, "find one User: ")
	}

	return model, nil
}

// FindMany finds multiple User based on the provided options.
func (t *userStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*User, error) {
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
		return nil, fmt.Errorf("failed to find User: %w", err)
	}
	defer rows.Close()

	var results []*User
	for rows.Next() {
		model := &User{}
		if err := model.ScanRows(rows); err != nil {
			return nil, fmt.Errorf("failed to scan User: %w", err)
		}
		results = append(results, model)
	}

	return results, nil
}

// FindOne finds a single User based on the provided options.
func (t *userStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error) {
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

// Count counts User based on the provided options.
func (t *userStorage) Count(ctx context.Context, builders ...*QueryBuilder) (int64, error) {
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
		return 0, fmt.Errorf("failed to count User: %w", err)
	}

	return count, nil
}

// FindManyWithPagination finds multiple User with pagination support.
func (t *userStorage) FindManyWithPagination(ctx context.Context, limit int, page int, builders ...*QueryBuilder) ([]*User, *Paginator, error) {
	// Count the total number of records
	totalCount, err := t.Count(ctx, builders...)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to count User: %w", err)
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
		return nil, nil, fmt.Errorf("failed to find User: %w", err)
	}

	return records, paginator, nil
}

// LockForUpdate lock locks the User for the given ID.
func (t *userStorage) LockForUpdate(ctx context.Context, builders ...*QueryBuilder) error {
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
		return fmt.Errorf("failed to build query: %w", err)
	}

	row := t.DB(ctx).QueryRowContext(ctx, sqlQuery, args...)
	var model User
	if err := model.ScanRow(row); err != nil {
		return fmt.Errorf("failed to scan User: %w", err)
	}

	return nil
}
