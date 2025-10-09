package db

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	sq "github.com/Masterminds/squirrel"
	"strings"
	"time"
)

// userStorage is a struct for the "users" table.
type userStorage struct {
	config       *Config
	queryBuilder sq.StatementBuilderType
}

// UserCRUDOperations is an interface for managing the users table.
type UserCRUDOperations interface {
	Create(ctx context.Context, model *User, opts ...Option) error
	AsyncCreate(ctx context.Context, model *User, opts ...Option) error
	BatchCreate(ctx context.Context, models []*User, opts ...Option) error
	OriginalBatchCreate(ctx context.Context, models []*User, opts ...Option) error
}

// UserSearchOperations is an interface for searching the users table.
type UserSearchOperations interface {
	FindMany(ctx context.Context, builder ...*QueryBuilder) ([]*User, error)
	FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error)
}

type UserSettings interface {
	Conn() driver.Conn
	TableName() string
	SetConfig(config *Config) UserStorage
	SetQueryBuilder(builder sq.StatementBuilderType) UserStorage
	Columns() []string
	GetQueryBuilder() sq.StatementBuilderType
}

// UserRelationLoading is an interface for loading relations.
type UserRelationLoading interface {
	LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadPosts(ctx context.Context, model *User, builders ...*QueryBuilder) error
	LoadBatchDevice(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error
	LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error
}

// UserRawQueryOperations is an interface for executing raw queries.
type UserRawQueryOperations interface {
	Select(ctx context.Context, query string, dest any, args ...any) error
	Exec(ctx context.Context, query string, args ...interface{}) error
	QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row
	QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error)
}

// UserStorage is a struct for the "users" table.
type UserStorage interface {
	UserCRUDOperations
	UserSearchOperations
	UserRelationLoading
	UserRawQueryOperations
	UserSettings
}

// NewUserStorage returns a new userStorage.
func NewUserStorage(config *Config) (UserStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if config.DB == nil {
		return nil, fmt.Errorf("config.DB connection is nil")
	}

	return &userStorage{
		config:       config,
		queryBuilder: sq.StatementBuilder.PlaceholderFormat(sq.Question),
	}, nil
}

// logQuery logs the query if query logging is enabled.
func (t *userStorage) logQuery(ctx context.Context, query string, args ...interface{}) {
	if t.config.QueryLogMethod != nil {
		t.config.QueryLogMethod(ctx, t.TableName(), query, args...)
	}
}

// logError logs the error if error logging is enabled.
func (t *userStorage) logError(ctx context.Context, err error, message string) {
	if t.config.ErrorLogMethod != nil {
		t.config.ErrorLogMethod(ctx, err, message)
	}
}

// applyPrewhere applies ClickHouse PREWHERE conditions to the query.
// PREWHERE is executed before WHERE and reads only the specified columns,
// which can significantly improve query performance.
func (t *userStorage) applyPrewhere(query string, args []interface{}, conditions []FilterApplier) (string, []interface{}) {
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
func (t *userStorage) applySettings(query string, settings map[string]interface{}) string {
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
func (t *userStorage) TableName() string {
	return "users"
}

// GetQueryBuilder returns the query builder.
func (t *userStorage) GetQueryBuilder() sq.StatementBuilderType {
	return t.queryBuilder
}

// Columns returns the columns for the table.
func (t *userStorage) Columns() []string {
	return []string{
		"id", "name", "age", "email", "last_name", "created_at", "updated_at", "notification_settings", "phones", "balls", "numrs", "comments",
	}
}

// DB returns the underlying DB. This is useful for doing transactions.
func (t *userStorage) DB() QueryExecer {
	return t.config.DB
}

func (t *userStorage) SetConfig(config *Config) UserStorage {
	t.config = config
	return t
}

func (t *userStorage) SetQueryBuilder(builder sq.StatementBuilderType) UserStorage {
	t.queryBuilder = builder
	return t
}

// LoadDevice loads the Device relation.
func (t *userStorage) LoadDevice(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s, err := NewDeviceStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create DeviceStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(DeviceUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one DeviceStorage: %w", err)
	}

	model.Device = relationModel
	return nil
}

// LoadSettings loads the Settings relation.
func (t *userStorage) LoadSettings(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewSettingStorage creates a new SettingStorage.
	s, err := NewSettingStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create SettingStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
	builders = append(builders, FilterBuilder(SettingUserIdEq(model.Id)))
	relationModel, err := s.FindOne(ctx, builders...)
	if err != nil {
		return fmt.Errorf("failed to find one SettingStorage: %w", err)
	}

	model.Settings = relationModel
	return nil
}

// LoadAddresses loads the Addresses relation.
func (t *userStorage) LoadAddresses(ctx context.Context, model *User, builders ...*QueryBuilder) error {
	if model == nil {
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewAddressStorage creates a new AddressStorage.
	s, err := NewAddressStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create AddressStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
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
		return fmt.Errorf("model is nil: %w", ErrModelIsNil)
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create PostStorage: %w", err)
	}
	// Add the filter for the relation without dereferencing
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
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewDeviceStorage creates a new DeviceStorage.
	s, err := NewDeviceStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create DeviceStorage: %w", err)
	}

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
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Device = v
		}
	}

	return nil
}

// LoadBatchSettings loads the Settings relation.
func (t *userStorage) LoadBatchSettings(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewSettingStorage creates a new SettingStorage.
	s, err := NewSettingStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create SettingStorage: %w", err)
	}

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
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Settings = v
		}
	}

	return nil
}

// LoadBatchAddresses loads the Addresses relation.
func (t *userStorage) LoadBatchAddresses(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewAddressStorage creates a new AddressStorage.
	s, err := NewAddressStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create AddressStorage: %w", err)
	}

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
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Addresses = v
		}
	}

	return nil
}

// LoadBatchPosts loads the Posts relation.
func (t *userStorage) LoadBatchPosts(ctx context.Context, items []*User, builders ...*QueryBuilder) error {
	requestItems := make([]interface{}, 0, len(items))
	for _, item := range items {
		// Append the value directly for non-optional fields
		requestItems = append(requestItems, item.Id)
	}

	// NewPostStorage creates a new PostStorage.
	s, err := NewPostStorage(t.config)
	if err != nil {
		return fmt.Errorf("failed to create PostStorage: %w", err)
	}

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
		// Assign the relation directly for non-optional fields
		if v, ok := resultMap[item.Id]; ok {
			item.Posts = v
		}
	}

	return nil
}

// User is a struct for the "users" table.
type User struct {
	Id                   string
	Name                 string
	Age                  int32
	Email                string
	LastName             *string
	Device               *Device
	Settings             *Setting
	Addresses            []*Address
	Posts                []*Post
	CreatedAt            time.Time
	UpdatedAt            *time.Time
	NotificationSettings *UserNotificationSetting
	Phones               UserPhonesRepeated
	Balls                UserBallsRepeated
	Numrs                UserNumrsRepeated
	Comments             UserCommentsRepeated
}

// TableName returns the table name.
func (t *User) TableName() string {
	return "users"
}

// ScanRow scans a row into a User.
func (t *User) ScanRow(row driver.Row) error {
	return row.Scan(
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
	Name  *string
	Age   *int32
	Email *string
}

// UserIdEq returns a condition that checks if the field equals the value.
func UserIdEq(value string) FilterApplier {
	return EqualsCondition{Field: "id", Value: value}
}

// UserNameEq returns a condition that checks if the field equals the value.
func UserNameEq(value string) FilterApplier {
	return EqualsCondition{Field: "name", Value: value}
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

// UserNameNotEq returns a condition that checks if the field equals the value.
func UserNameNotEq(value string) FilterApplier {
	return NotEqualsCondition{Field: "name", Value: value}
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

// UserNameGT greaterThanCondition than condition.
func UserNameGT(value string) FilterApplier {
	return GreaterThanCondition{Field: "name", Value: value}
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

// UserNameLT less than condition.
func UserNameLT(value string) FilterApplier {
	return LessThanCondition{Field: "name", Value: value}
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

// UserNameGTE greater than or equal condition.
func UserNameGTE(value string) FilterApplier {
	return GreaterThanOrEqualCondition{Field: "name", Value: value}
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

// UserNameLTE less than or equal condition.
func UserNameLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "name", Value: value}
}

// UserAgeLTE less than or equal condition.
func UserAgeLTE(value int32) FilterApplier {
	return LessThanOrEqualCondition{Field: "age", Value: value}
}

// UserEmailLTE less than or equal condition.
func UserEmailLTE(value string) FilterApplier {
	return LessThanOrEqualCondition{Field: "email", Value: value}
}

// UserIdBetween between condition.
func UserIdBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "id", Min: min, Max: max}
}

// UserNameBetween between condition.
func UserNameBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "name", Min: min, Max: max}
}

// UserAgeBetween between condition.
func UserAgeBetween(min, max int32) FilterApplier {
	return BetweenCondition{Field: "age", Min: min, Max: max}
}

// UserEmailBetween between condition.
func UserEmailBetween(min, max string) FilterApplier {
	return BetweenCondition{Field: "email", Min: min, Max: max}
}

// UserIdILike iLike condition %
func UserIdILike(value string) FilterApplier {
	return ILikeCondition{Field: "id", Value: value}
}

// UserNameILike iLike condition %
func UserNameILike(value string) FilterApplier {
	return ILikeCondition{Field: "name", Value: value}
}

// UserEmailILike iLike condition %
func UserEmailILike(value string) FilterApplier {
	return ILikeCondition{Field: "email", Value: value}
}

// UserIdLike like condition %
func UserIdLike(value string) FilterApplier {
	return LikeCondition{Field: "id", Value: value}
}

// UserNameLike like condition %
func UserNameLike(value string) FilterApplier {
	return LikeCondition{Field: "name", Value: value}
}

// UserEmailLike like condition %
func UserEmailLike(value string) FilterApplier {
	return LikeCondition{Field: "email", Value: value}
}

// UserIdNotLike not like condition
func UserIdNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "id", Value: value}
}

// UserNameNotLike not like condition
func UserNameNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "name", Value: value}
}

// UserEmailNotLike not like condition
func UserEmailNotLike(value string) FilterApplier {
	return NotLikeCondition{Field: "email", Value: value}
}

// UserIdIn condition
func UserIdIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "id", Values: values}
}

// UserNameIn condition
func UserNameIn(values ...interface{}) FilterApplier {
	return InCondition{Field: "name", Values: values}
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

// UserNameNotIn not in condition
func UserNameNotIn(values ...interface{}) FilterApplier {
	return NotInCondition{Field: "name", Values: values}
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

// UserNameOrderBy sorts the result in ascending order.
func UserNameOrderBy(asc bool) FilterApplier {
	return OrderBy("name", asc)
}

// UserAgeOrderBy sorts the result in ascending order.
func UserAgeOrderBy(asc bool) FilterApplier {
	return OrderBy("age", asc)
}

// UserEmailOrderBy sorts the result in ascending order.
func UserEmailOrderBy(asc bool) FilterApplier {
	return OrderBy("email", asc)
}

// AsyncCreate asynchronously inserts a new User.
func (t *userStorage) AsyncCreate(ctx context.Context, model *User, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// Set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	// Get value of phones
	phones, err := model.Phones.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Phones: %w", err)
	}
	// Get value of balls
	balls, err := model.Balls.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Balls: %w", err)
	}
	// Get value of numrs
	numrs, err := model.Numrs.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Numrs: %w", err)
	}
	// Get value of comments
	comments, err := model.Comments.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Comments: %w", err)
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
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	if err := t.DB().AsyncInsert(ctx, sqlQuery, options.waitAsyncInsert, args...); err != nil {
		return fmt.Errorf("failed to asynchronously create User: %w", err)
	}
	if options.relations && model.Device != nil {
		s, err := NewDeviceStorage(t.config)
		if err != nil {
			return fmt.Errorf("failed to create Device: %w", err)
		}

		err = s.AsyncCreate(ctx, model.Device)
		if err != nil {
			return fmt.Errorf("failed to asynchronously create Device: %w", err)
		}
	}
	if options.relations && model.Settings != nil {
		s, err := NewSettingStorage(t.config)
		if err != nil {
			return fmt.Errorf("failed to create Settings: %w", err)
		}

		err = s.AsyncCreate(ctx, model.Settings)
		if err != nil {
			return fmt.Errorf("failed to asynchronously create Settings: %w", err)
		}
	}
	if options.relations && model.Addresses != nil {
		for _, item := range model.Addresses {
			s, err := NewAddressStorage(t.config)
			if err != nil {
				return fmt.Errorf("failed to create Addresses: %w", err)
			}

			err = s.AsyncCreate(ctx, item)
			if err != nil {
				return fmt.Errorf("failed to asynchronously create Addresses: %w", err)
			}
		}
	}

	return nil
}

// Create creates a new User.
func (t *userStorage) Create(ctx context.Context, model *User, opts ...Option) error {
	if model == nil {
		return fmt.Errorf("model is nil")
	}

	// set default options
	options := &Options{}
	for _, o := range opts {
		o(options)
	}
	// get value of phones
	phones, err := model.Phones.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Phones: %w", err)
	}
	// get value of balls
	balls, err := model.Balls.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Balls: %w", err)
	}
	// get value of numrs
	numrs, err := model.Numrs.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Numrs: %w", err)
	}
	// get value of comments
	comments, err := model.Comments.Value()
	if err != nil {
		return fmt.Errorf("failed to get value of Comments: %w", err)
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
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
		)

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build query: %w", err)
	}
	t.logQuery(ctx, sqlQuery, args...)

	err = t.DB().Exec(ctx, sqlQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to create User: %w", err)
	}
	if options.relations && model.Device != nil {
		s, err := NewDeviceStorage(t.config)
		if err != nil {
			return fmt.Errorf("failed to create Device: %w", err)
		}

		err = s.Create(ctx, model.Device)
		if err != nil {
			return fmt.Errorf("failed to create Device: %w", err)
		}
	}
	if options.relations && model.Settings != nil {
		s, err := NewSettingStorage(t.config)
		if err != nil {
			return fmt.Errorf("failed to create Settings: %w", err)
		}

		err = s.Create(ctx, model.Settings)
		if err != nil {
			return fmt.Errorf("failed to create Settings: %w", err)
		}
	}
	if options.relations && model.Addresses != nil {
		for _, item := range model.Addresses {
			s, err := NewAddressStorage(t.config)
			if err != nil {
				return fmt.Errorf("failed to create Addresses: %w", err)
			}

			err = s.Create(ctx, item)
			if err != nil {
				return fmt.Errorf("failed to create Addresses: %w", err)
			}
		}
	}

	return nil
}

// BatchCreate creates multiple User records in a single batch.
func (t *userStorage) BatchCreate(ctx context.Context, models []*User, opts ...Option) error {
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
		// Get value of phones
		phones, err := model.Phones.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Phones: %w", err)
		}
		// Get value of balls
		balls, err := model.Balls.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Balls: %w", err)
		}
		// Get value of numrs
		numrs, err := model.Numrs.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Numrs: %w", err)
		}
		// Get value of comments
		comments, err := model.Comments.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Comments: %w", err)
		}

		err = batch.Append(
			model.Name,
			model.Age,
			model.Email,
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
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

// OriginalBatchCreate creates multiple User records in a single batch.
func (t *userStorage) OriginalBatchCreate(ctx context.Context, models []*User, opts ...Option) error {
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
		)

	for _, model := range models {
		if model == nil {
			return fmt.Errorf("model is nil: %w", ErrModelIsNil)
		}
		// Get value of phones
		phones, err := model.Phones.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Phones: %w", err)
		}
		// Get value of balls
		balls, err := model.Balls.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Balls: %w", err)
		}
		// Get value of numrs
		numrs, err := model.Numrs.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Numrs: %w", err)
		}
		// Get value of comments
		comments, err := model.Comments.Value()
		if err != nil {
			return fmt.Errorf("failed to get value of Comments: %w", err)
		}

		query = query.Values(
			model.Name,
			model.Age,
			model.Email,
			nullValue(model.LastName),
			model.CreatedAt,
			nullValue(model.UpdatedAt),
			nullValue(model.NotificationSettings),
			phones,
			balls,
			numrs,
			comments,
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

// FindMany finds multiple User based on the provided options.
func (t *userStorage) FindMany(ctx context.Context, builders ...*QueryBuilder) ([]*User, error) {
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

	var results []*User
	for rows.Next() {
		model := &User{}
		if err := model.ScanRow(rows); err != nil { // Используем ScanRow вместо ScanRows
			return nil, fmt.Errorf("failed to scan User: %w", err)
		}
		results = append(results, model)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over rows: %w", err)
	}

	return results, nil
}

// FindOne finds a single User based on the provided options.
func (t *userStorage) FindOne(ctx context.Context, builders ...*QueryBuilder) (*User, error) {
	// Use findMany but limit the results to 1
	builders = append(builders, LimitBuilder(1))
	results, err := t.FindMany(ctx, builders...)
	if err != nil {
		return nil, fmt.Errorf("failed to findOne User: %w", err)
	}

	if len(results) == 0 {
		return nil, ErrRowNotFound
	}

	return results[0], nil
}

// Select executes a raw query and returns the result.
func (t *userStorage) Select(ctx context.Context, query string, dest any, args ...any) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Select(ctx, dest, query, args...)
}

// Exec executes a raw query and returns the result.
func (t *userStorage) Exec(ctx context.Context, query string, args ...interface{}) error {
	t.logQuery(ctx, query, args...)
	return t.DB().Exec(ctx, query, args...)
}

// QueryRow executes a raw query and returns the result.
func (t *userStorage) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	t.logQuery(ctx, query, args...)
	return t.DB().QueryRow(ctx, query, args...)
}

// QueryRows executes a raw query and returns the result.
func (t *userStorage) QueryRows(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	t.logQuery(ctx, query, args...)
	return t.DB().Query(ctx, query, args...)
}

// Conn returns the connection.
func (t *userStorage) Conn() driver.Conn {
	return t.DB()
}
