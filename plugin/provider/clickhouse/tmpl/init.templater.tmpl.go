package tmpl

// InitStatementTemplate is the template for the init functions.
const InitStatementTemplate = `
{{ if .IncludeConnection }}
// 
// Database connection.
//

{{ template "connection" . }}
{{ end }}

// 
// storages.
//

{{ template "storages" . }}

// 
// Json types.
//

{{ template "types" . }}

// 
// Single repeated types.
//

{{ template "repeatedTypes" . }}

// 
// errors.
//

{{ template "errors" . }}

//
// Transaction manager.
//

{{ template "transaction" . }}

//
// Options.
// 
{{ template "options" . }}

// 
// Conditions for query builder.
// 
{{ template "conditions" . }}
`

const OptionsTemplate = `
// Option is a function that configures the {{ storageName }}.
type Option func(*Options)

// Options are the options for the {{ storageName }}.
type Options struct {
	// if true, then method was create/update relations
	relations bool
	// uniqField is the unique field.
	uniqField string
}

// WithRelations sets the relations flag.
// This is used to determine if the relations should be created or updated.
func WithRelations() Option {
	return func(o *Options) {
		o.relations = true
	}
}

// WithUniqField sets the unique field.
func WithUniqField(field string) Option {
	return func(o *Options) {
		o.uniqField = field
	}
}

// FilterApplier is a condition filters.
type FilterApplier interface {
	Apply(query sq.SelectBuilder) sq.SelectBuilder
	ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder
}

// CustomFilter is a custom filter.
type CustomFilter interface {
	ApplyFilter(query sq.SelectBuilder, params any) sq.SelectBuilder
}

// QueryBuilder is a query builder.
type QueryBuilder struct {
	// additional options for the query.
	options       []Option
	// filterOptions are the filter options.
	filterOptions []FilterApplier
	// orderOptions are the order options.
	sortOptions  []FilterApplier
	// pagination is the pagination.
	pagination    *Pagination
	// customFilters are the custom filters.
	customFilters []struct {
		filter CustomFilter
		params any
	}
	// customTableName is the custom table name.
	customTableName string
}

// NewQueryBuilder returns a new query builder.
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}

// WithOptions sets the options for the query.
func (b *QueryBuilder) WithOptions(options ...Option) *QueryBuilder {
	b.options = options
	return b
}

// WithCustomFilter sets a custom filter for the query.
func (qb *QueryBuilder) WithCustomFilter(filter CustomFilter, params any) *QueryBuilder {
	qb.customFilters = append(qb.customFilters, struct {
		filter CustomFilter
		params any
	}{
		filter: filter,
		params: params,
	})
	return qb
}

// WithCustomTableName sets a custom table name for the query.
func (qb *QueryBuilder) WithCustomTableName(tableName string) *QueryBuilder {
	qb.customTableName = tableName
	return qb
}

// nullValue returns the null value.
func nullValue[T any](v *T) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

// Apply customTableName to the query.
func (qb *QueryBuilder) ApplyCustomTableName(query sq.SelectBuilder) sq.SelectBuilder {
	if qb.customTableName != "" {
		query = query.From(qb.customTableName)
	}
	return query
}

// ApplyCustomFilters applies the custom filters to the query.
func (qb *QueryBuilder) ApplyCustomFilters(query sq.SelectBuilder) sq.SelectBuilder {
	for _, cf := range qb.customFilters {
		query = cf.filter.ApplyFilter(query, cf.params)
	}
	return query
}

// WithFilterOptions sets the filter options for the query.
func (b *QueryBuilder) WithFilter(filterOptions ...FilterApplier) *QueryBuilder {
	b.filterOptions = filterOptions
	return b
}

// WithSort sets the sort options for the query.
func (b *QueryBuilder) WithSort(sortOptions ...FilterApplier) *QueryBuilder {
	b.sortOptions = sortOptions
	return b
}

// WithPagination sets the pagination for the query.
func (b *QueryBuilder) WithPagination(pagination *Pagination) *QueryBuilder {
	b.pagination = pagination
	return b
}

// Filter is a helper function to create a new query builder with filter options.
func FilterBuilder(filterOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithFilter(filterOptions...)
}

// SortBuilder is a helper function to create a new query builder with sort options.
func SortBuilder(sortOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithSort(sortOptions...)
}

// Options is a helper function to create a new query builder with options.
func LimitBuilder(limit uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(&Pagination{
		limit: &limit,
	})
}

// Offset is a helper function to create a new query builder with options.
func OffsetBuilder(offset uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(&Pagination{
		offset: &offset,
	})
}

// Paginate is a helper function to create a new query builder with options.
func PaginateBuilder(limit, offset uint64) *QueryBuilder {
	return NewQueryBuilder().WithPagination(NewPagination(limit, offset))
}

// Pagination is the pagination.
type Pagination struct {
	// limit is the limit.
	limit *uint64
	// offset is the offset.
	offset *uint64
}

// NewPagination returns a new pagination.
// If limit or offset are nil, then they will be omitted.
func NewPagination(limit, offset uint64) *Pagination {
	return &Pagination{
		limit: &limit,
		offset: &offset,
	}
}

// Limit is a helper function to create a new pagination.
func Limit(limit uint64) *Pagination {
	return &Pagination{
		limit: &limit,
	}
}

// Offset is a helper function to create a new pagination.
func Offset(offset uint64) *Pagination {
	return &Pagination{
		offset: &offset,
	}
}
`

// ConnectionTemplate is the template for the connection functions.
// This is included in the init template.
const ConnectionTemplate = `
func Open(ctx context.Context, dsn string) (driver.Conn, error) {
	parsedOptions, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "parse dsn")
	}

	conn, err := clickhouse.Open(parsedOptions)
	if err != nil {
		return nil, errors.Wrap(err, "open clickhouse connection")
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, errors.Wrap(err, "ping clickhouse instance")
	}

	return conn, nil
}
`

// StorageTemplate is the template for the storage functions.
// This is included in the init template.
const StorageTemplate = `
// {{ storageName | lowerCamelCase }} is a map of provider to init function.
type {{ storageName | lowerCamelCase }} struct {
	config *Config // configuration for the {{ storageName }}.
{{ range $value := storages }}
{{ $value.Key }} {{ $value.Value }}{{ end }}
}

// configuration for the {{ storageName }}.
type Config struct {
	DB driver.Conn

	QueryLogMethod    func(ctx context.Context, table string, query string, args ...interface{})
	ErrorLogMethod    func(ctx context.Context, err error, message string)
}

// {{ storageName }} is the interface for the {{ storageName }}.
type {{ storageName }} interface { 
	{{- range $value := storages }}
	// Get{{ $value.Value }} returns the {{ $value.Value }} store.
	Get{{ $value.Value }}() {{ $value.Value }}
	{{- end }}

{{ if .CRUDSchemas }}
	// CreateTables creates the tables for all the stores.
	CreateTables(ctx context.Context) error
	// DropTables drops the tables for all the stores.
	DropTables(ctx context.Context) error
	// TruncateTables truncates the tables for all the stores.
	TruncateTables(ctx context.Context) error
	// UpgradeTables upgrades the tables for all the stores.
	UpgradeTables(ctx context.Context) error
{{ end }}
}

// New{{ storageName }} returns a new {{ storageName }}.
func New{{ storageName }}(config *Config) ({{ storageName }}, error) {
	if config == nil {
		return nil, errors.New("config is required")
	}

	if config.DB == nil {
		return nil, errors.New("db is required")
	}
	
	var storages = {{ storageName | lowerCamelCase }}{
		config: config,
	}
{{ range $value := storages }}
	{{ $value.Key }}Impl, err := New{{ $value.Value }}(config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create {{ $value.Value }}")
	}
	storages.{{ $value.Key }} = {{ $value.Key }}Impl
{{ end }}

	return &storages, nil
}

{{ range $value := storages }}
// Get{{ $value.Value }} returns the {{ $value.Value }} store.
func (c *{{ storageName | lowerCamelCase }}) Get{{ $value.Value }}() {{ $value.Value }} {
	return c.{{ $value.Key }}
}
{{ end }}

{{ if .CRUDSchemas }}
// CreateTables creates the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) CreateTables(ctx context.Context) error {
	var err error
{{ range $value := storages }}
	// create the {{ $value.Value }} table.
	err = c.{{ $value.Key }}.CreateTable(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to create table")
	}
{{ end }}
	return nil
}

// DropTables drops the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) DropTables(ctx context.Context) error {
	var err error
{{ range $value := storages }}
	// drop the {{ $value.Value }} table.
	err = c.{{ $value.Key }}.DropTable(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to drop table")
	}
{{ end }}
	return nil
}

// TruncateTables truncates the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) TruncateTables(ctx context.Context) error {
	var err error
{{ range  $value := storages }}
	// truncate the {{ $value.Value }} table.
	err = c.{{ $value.Key }}.TruncateTable(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to truncate table")
	}
{{ end }}
	return nil
}

// UpgradeTables runs the database upgrades for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) UpgradeTables(ctx context.Context) error {
	var err error
{{ range $value := storages }}
	// run the {{ $value.Value }} upgrade.
	err = c.{{ $value.Key }}.UpgradeTable(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade table")
	}
{{ end }}
	return nil
}
{{ end }}
`

const TypesTemplate = `
// NullableJSON represents a JSON field that can be null.
type NullableJSON[T any] struct {
	Data  T
	Valid bool // Valid is true if the field is not NULL
}

// NewNullableJSON creates a new NullableJSON with a value.
func NewNullableJSON[T any](v T) NullableJSON[T] {
	return NullableJSON[T]{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface.
func (n *NullableJSON[T]) Scan(value interface{}) error {
	if value == nil {
		n.Valid = false
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &n.Data); err != nil {
		n.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	n.Valid = true
	return nil
}

// Value converts NullableJSON to a string representation for ClickHouse.
func (n *NullableJSON[T]) Value() (string, error) {
	if !n.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(n.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// ValueOrZero returns the value if valid, otherwise returns the zero value of type T.
func (n NullableJSON[T]) ValueOrZero() T {
	if !n.Valid {
		var zero T
		return zero
	}
	return n.Data
}

{{ range $key, $field := nestedMessages }}
// {{ $key }} is a JSON type nested in another message.
type {{ $field.StructureName }} struct {
	{{- range $nestedField := $field.Descriptor.GetField }}
	{{ $nestedField | fieldName }} {{ $nestedField | fieldType }}` + " `json:\"{{ $nestedField | sourceName }}\"`" + `
	{{- end }}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *{{ $field.StructureName }}) Scan(src interface{}) error {
	if str, ok := src.(string); ok {
		return json.Unmarshal([]byte(str), m)
	}
	return fmt.Errorf("can't convert %T to string", src)
}

// Value converts the struct to a JSON string.
func (m *{{ $field.StructureName }}) Value() (string, error) {
	if m == nil {
		m = &{{ $field.StructureName }}{}
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}
{{ end }}`

const SingleRepeatedTypesTemplate = `
{{ range $field := singleTypes }}
// {{ $field.FieldType }} is a JSON type nested in another message.
type {{ $field.FieldType }} struct {
	Data  {{ $field.Descriptor | fieldType }}
	Valid bool // Valid is true if the field is not NULL
}

// New{{ $field.SourceName | camelCase }}Field creates a new {{ $field.FieldType }}.
func New{{ $field.SourceName | camelCase }}Field(v {{ $field.Descriptor | fieldType }}) {{ $field.FieldType }} {
	return {{ $field.FieldType }}{Data: v, Valid: true}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *{{ $field.FieldType }}) Scan(src interface{}) error {
	if src == nil {
		m.Valid = false
		return nil
	}

	str, ok := src.(string)
	if !ok {
		return fmt.Errorf("failed to convert value to string")
	}

	if err := json.Unmarshal([]byte(str), &m.Data); err != nil {
		m.Valid = false
		return fmt.Errorf("failed to unmarshal json: %w", err)
	}

	m.Valid = true
	return nil
}

// Value converts the struct to a JSON string.
func (m {{ $field.FieldType }}) Value() (string, error) {
	if !m.Valid {
		return "", nil
	}

	bytes, err := json.Marshal(m.Data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal json: %w", err)
	}
	return string(bytes), nil
}

// Get returns the value of the field or the zero value if invalid.
func (m {{ $field.FieldType }}) Get() {{ $field.Descriptor | fieldType }} {
	if !m.Valid {
		var zero {{ $field.Descriptor | fieldType }}
		return zero
	}
	return m.Data
}

func (m {{ $field.FieldType }}) String() string {
	return fmt.Sprintf("%v", m.Get())
}
{{ end }}`

const TransactionManagerTemplate = `
// QueryExecer is an interface that can execute queries.
type QueryExecer interface {
	driver.Conn
}
`

// ErrorsTemplate is the template for the errors.
// This is included in the init template.
const ErrorsTemplate = `
var (
	// ErrNotFound is returned when a record is not found.
	ErrRowNotFound = errors.New("row not found")
	// ErrNoTransaction is returned when a transaction is not provided.
	ErrNoTransaction = errors.New("no transaction provided")
	// ErrRowAlreadyExist is returned when a row already exist.
	ErrRowAlreadyExist    = errors.New("row already exist")
	// ErrModelIsNil is returned when a relation model is nil.
	ErrModelIsNil = errors.New("model is nil")
)
`
