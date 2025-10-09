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
	// waitAsyncInsert is the wait flag. wait_for_async_insert = 1
	waitAsyncInsert bool
}

// WithWaitAsyncInsert sets the waitAsyncInsert flag.
func WithWaitAsyncInsert() Option {
	return func(o *Options) {
		o.waitAsyncInsert = true
	}
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
	// prewhereOptions are the PREWHERE filter options (ClickHouse specific).
	prewhereOptions []FilterApplier
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
	// settings are the ClickHouse query settings.
	settings map[string]interface{}
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

// WithPrewhere sets the PREWHERE filter options for the query (ClickHouse specific).
// PREWHERE is executed before reading all columns, which can significantly improve performance.
func (b *QueryBuilder) WithPrewhere(prewhereOptions ...FilterApplier) *QueryBuilder {
	b.prewhereOptions = prewhereOptions
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

// WithSettings sets the ClickHouse query settings.
func (b *QueryBuilder) WithSettings(settings map[string]interface{}) *QueryBuilder {
	if b.settings == nil {
		b.settings = make(map[string]interface{})
	}
	for k, v := range settings {
		b.settings[k] = v
	}
	return b
}

// WithSetting sets a single ClickHouse query setting.
func (b *QueryBuilder) WithSetting(key string, value interface{}) *QueryBuilder {
	if b.settings == nil {
		b.settings = make(map[string]interface{})
	}
	b.settings[key] = value
	return b
}

// Filter is a helper function to create a new query builder with filter options.
func FilterBuilder(filterOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithFilter(filterOptions...)
}

// PrewhereBuilder is a helper function to create a new query builder with PREWHERE options (ClickHouse specific).
// PREWHERE filters are applied before reading all columns, improving performance for queries with selective filters.
func PrewhereBuilder(prewhereOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithPrewhere(prewhereOptions...)
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

// SettingsBuilder is a helper function to create a new query builder with settings.
func SettingsBuilder(settings map[string]interface{}) *QueryBuilder {
	return NewQueryBuilder().WithSettings(settings)
}

// SettingBuilder is a helper function to create a new query builder with a single setting.
func SettingBuilder(key string, value interface{}) *QueryBuilder {
	return NewQueryBuilder().WithSetting(key, value)
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

// ClickHouse query settings constants
const (
	// Query execution settings
	SettingMaxThreads                           = "max_threads"
	SettingMaxMemoryUsage                       = "max_memory_usage"
	SettingMaxBytesToRead                       = "max_bytes_to_read"
	SettingMaxExecutionTime                     = "max_execution_time"
	SettingMaxRowsToRead                        = "max_rows_to_read"
	
	// Optimization settings
	SettingOptimizeReadInOrder                  = "optimize_read_in_order"
	SettingForcePrimaryKey                      = "force_primary_key"
	SettingUseUncompressedCache                 = "use_uncompressed_cache"
	SettingMaxColumnsToRead                     = "max_columns_to_read"
	
	// MergeTree settings
	SettingMergeTreeMinRowsForConcurrentRead    = "merge_tree_min_rows_for_concurrent_read"
	SettingMergeTreeMinBytesForConcurrentRead   = "merge_tree_min_bytes_for_concurrent_read"
	SettingMergeTreeCoarseIndexGranularity      = "merge_tree_coarse_index_granularity"
	
	// Load balancing settings
	SettingLoadBalancing                        = "load_balancing"
	
	// Distributed query settings
	SettingMaxParallelReplicas                  = "max_parallel_replicas"
	SettingDistributedAggregationMemoryEfficient = "distributed_aggregation_memory_efficient"
	
	// Insert settings
	SettingAsyncInsert                          = "async_insert"
	SettingWaitForAsyncInsert                   = "wait_for_async_insert"
	SettingAsyncInsertThreads                   = "async_insert_threads"
	
	// Query result settings
	SettingMaxResultRows                        = "max_result_rows"
	SettingMaxResultBytes                       = "max_result_bytes"
	SettingResultOverflowMode                   = "result_overflow_mode"
	
	// Connection settings
	SettingConnectTimeout                       = "connect_timeout"
	SettingReceiveTimeout                       = "receive_timeout"
	SettingSendTimeout                          = "send_timeout"
	
	// Additional optimization settings
	SettingOptimizeAggregationInOrder           = "optimize_aggregation_in_order"
	SettingOptimizeMoveToPrewhere               = "optimize_move_to_prewhere"
	SettingAllowSuspiciousLowCardinalityTypes   = "allow_suspicious_low_cardinality_types"
)
`

// ConnectionTemplate is the template for the connection functions.
// This is included in the init template.
const ConnectionTemplate = `
func Open(ctx context.Context, dsn string) (driver.Conn, error) {
	parsedOptions, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	conn, err := clickhouse.Open(parsedOptions)
	if err != nil {
		return nil, fmt.Errorf("open clickhouse connection: %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		if exception, ok := err.(*clickhouse.Exception); ok {
			fmt.Printf("Exception [%d] %s \n%s\n", exception.Code, exception.Message, exception.StackTrace)
		}
		return nil, fmt.Errorf("ping clickhouse instance: %w", err)
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
		return nil, fmt.Errorf("config is required")
	}

	if config.DB == nil {
		return nil, fmt.Errorf("db is required")
	}
	
	var storages = {{ storageName | lowerCamelCase }}{
		config: config,
	}
{{ range $value := storages }}
	{{ $value.Key }}Impl, err := New{{ $value.Value }}(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create {{ $value.Value }}: %w", err)
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
		return fmt.Errorf("failed to create table: %w", err)
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
		return fmt.Errorf("failed to drop table: %w", err)
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
		return fmt.Errorf("failed to truncate table: %w", err)
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
		return fmt.Errorf("failed to upgrade table: %w", err)
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
	ErrRowNotFound = fmt.Errorf("row not found")
	// ErrNoTransaction is returned when a transaction is not provided.
	ErrNoTransaction = fmt.Errorf("no transaction provided")
	// ErrRowAlreadyExist is returned when a row already exist.
	ErrRowAlreadyExist    = fmt.Errorf("row already exist")
	// ErrModelIsNil is returned when a relation model is nil.
	ErrModelIsNil = fmt.Errorf("model is nil")
)
`
