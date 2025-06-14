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
}

// WithRelations sets the relations flag.
// This is used to determine if the relations should be created or updated.
func WithRelations() Option {
	return func(o *Options) {
		o.relations = true
	}
}

// FilterApplier is a condition filters.
type FilterApplier interface {
	Apply(query sq.SelectBuilder) sq.SelectBuilder
	ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder
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
func Dsn(filepath string, mode string, cache string, timeout int) string {
	dsn := fmt.Sprintf("file:%s", filepath)

	// Mode is optional, common values are: ro, rw, rwc, memory
	if mode != "" {
		dsn += "?mode=" + mode
	}

	// Cache is optional, common values are: shared, private
	if cache != "" {
		if mode != "" {
			dsn += "&cache=" + cache
		} else {
			dsn += "?cache=" + cache
		}
	}

	// Timeout is optional, in milliseconds
	if timeout != 0 {
		if mode != "" || cache != "" {
			dsn += "&_timeout=" + strconv.Itoa(timeout)
		} else {
			dsn += "?_timeout=" + strconv.Itoa(timeout)
		}
	}

	return dsn
}

func Open(dsn string, opts ...{{ clientName }}Option) (*sql.DB, error) {
    options := &{{ clientName }}Options{
			MaxOpenConns: 10,
			MaxIdleConns: 5,
			ConnMaxIdleTime: time.Minute,
			MaxLifetime: time.Minute * 2,
	}

    for _, opt := range opts {
        opt(options)
    }

    db, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open database: %w", err)
    }

	// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
	if err = db.Ping(); err != nil {
		// If Ping fails, close the DB and return an error.
		db.Close() // Ignoring error from Close, as we already have a more significant error.
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Set the connection pool options.
	db.SetMaxOpenConns(options.MaxOpenConns)
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	db.SetMaxIdleConns(options.MaxIdleConns)
	// SetConnMaxIdleTime sets the maximum amount of time a connection may be idle.
	db.SetConnMaxIdleTime(options.ConnMaxIdleTime)
	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	db.SetConnMaxLifetime(options.MaxLifetime)

    return db, nil
}

// {{ clientName }}Options are the options for the {{ clientName }}.
type {{ clientName }}Options struct {
    MaxOpenConns int
	MaxIdleConns int
	ConnMaxIdleTime time.Duration
	MaxLifetime time.Duration
}

// Option is a function that configures the {{ clientName }}.
type {{ clientName }}Option func(*{{ clientName }}Options)

// WithMaxOpenConns sets the maximum number of open connections to the database.
func WithMaxOpenConns(maxOpenConns int) {{ clientName }}Option {
	return func(opts *{{ clientName }}Options) {
		opts.MaxOpenConns = maxOpenConns
	}
}

// WithMaxIdleConns sets the maximum number of idle connections to the database.
func WithMaxIdleConns(maxIdleConns int) {{ clientName }}Option {
	return func(opts *{{ clientName }}Options) {
		opts.MaxIdleConns = maxIdleConns
	}
}

// WithConnMaxIdleTime sets the maximum amount of time a connection may be idle.
func WithConnMaxIdleTime(connMaxIdleTime time.Duration) {{ clientName }}Option {
	return func(opts *{{ clientName }}Options) {
		opts.ConnMaxIdleTime = connMaxIdleTime
	}
}

// WithMaxLifetime sets the maximum amount of time a connection may be reused.
func WithMaxLifetime(maxLifetime time.Duration) {{ clientName }}Option {
	return func(opts *{{ clientName }}Options) {
		opts.MaxLifetime = maxLifetime
	}
}
`

// StorageTemplate is the template for the storage functions.
// This is included in the init template.
const StorageTemplate = `
// {{ storageName | lowerCamelCase }} is a map of provider to init function.
type {{ storageName | lowerCamelCase }} struct {
	db *sql.DB // The database connection.
	tx *TxManager // The transaction manager.
{{ range $value := storages }}
{{ $value.Key }} {{ $value.Value }}{{ end }}
}

// {{ storageName }} is the interface for the {{ storageName }}.
type {{ storageName }} interface { 
	{{- range $value := storages }}
	// Get{{ $value.Value }} returns the {{ $value.Value }} store.
	Get{{ $value.Value }}() {{ $value.Value }}
	{{- end }}
	// TxManager returns the transaction manager.
	TxManager() *TxManager
	// CreateTables creates the tables for all the stores.
	CreateTables(ctx context.Context) error
	// DropTables drops the tables for all the stores.
	DropTables(ctx context.Context) error
	// TruncateTables truncates the tables for all the stores.
	TruncateTables(ctx context.Context) error
	// UpgradeTables upgrades the tables for all the stores.
	UpgradeTables(ctx context.Context) error
}

// New{{ storageName }} returns a new {{ storageName }}.
func New{{ storageName }}(config *Config) ({{ storageName }}, error) {
    if config == nil {
        return nil, fmt.Errorf("config is required")
    }

    if config.DB == nil {
        return nil, fmt.Errorf("db is required")
    }

    if config.DB.DBRead == nil {
        return nil, fmt.Errorf("db read is required")
    }

    if config.DB.DBWrite == nil {
        return nil, fmt.Errorf("db write is required")
    }

    storages := &{{ storageName | lowerCamelCase }}Impl{
        config: config,
    }

    {{ range $key, $value := .Storages }}
    {{ $value.Key }}Impl, err := New{{ $value.Value }}(config)
    if err != nil {
        return nil, fmt.Errorf("failed to create {{ $value.Value }}: %w", err)
    }
    storages.{{ $value.Key }} = {{ $value.Key }}Impl
    {{ end }}

    return storages, nil
}

// TxManager returns the transaction manager.
func (c *{{ storageName | lowerCamelCase }}) TxManager() *TxManager {
	return c.tx
}

{{ range $value := storages }}
// Get{{ $value.Value }} returns the {{ $value.Value }} store.
func (c *{{ storageName | lowerCamelCase }}) Get{{ $value.Value }}() {{ $value.Value }} {
	return c.{{ $value.Key }}
}
{{ end }}

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
		return fmt.Errorf("failed to upgrade: %w", err)
	}
{{ end }}
	return nil
}
`

const TypesTemplate = `
{{ range $key, $field := nestedMessages }}
// {{ $field.StructureName }} is a JSON type nested in another message.
type {{ $field.StructureName }} struct {
	{{- range $nestedField := $field.Descriptor.GetField }}
	{{ $nestedField | fieldName }} {{ $nestedField | fieldType }}` + " `json:\"{{ $nestedField | sourceName }}\"`" + `
	{{- end }}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *{{ $field.StructureName }}) Scan(src interface{}) error  {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return fmt.Errorf(fmt.Sprintf("can't convert %T", src))
}

// Value implements the driver.Valuer interface for JSON.
func (m *{{ $field.StructureName }}) Value() (driver.Value, error) {
	if m == nil {
		m = &{{ $field.StructureName }}{}
	}
	return json.Marshal(m)
}
{{ end }}
`

const SingleRepeatedTypesTemplate = `
{{ range $field := singleTypes }}
// {{ $field.FieldType }} is a JSON type nested in another message.
type {{ $field.FieldType }} {{ $field.Descriptor | fieldType }}

// New{{ $field.SourceName | camelCase }}Field returns a new {{ $field.FieldType }}.
func New{{ $field.SourceName | camelCase }}Field (v {{ $field.Descriptor | fieldType }}) {{ $field.FieldType }} {
	return v
}

// Scan implements the sql.Scanner interface for JSON.
func (m *{{ $field.FieldType }}) Scan(src interface{}) error  {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return fmt.Errorf("can't convert %T", src)
}

// Value implements the driver.Valuer interface for JSON.
func (m {{ $field.FieldType }}) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m {{ $field.FieldType }}) Get() {{ $field.Descriptor | fieldType }} {
	return m
}

func (m *{{ $field.FieldType }}) String() string {
	return fmt.Sprintf("%v", m.Get())
}
{{ end }}

// Pagination is the pagination.
type Paginator struct {
	TotalCount int64
	Limit      int
	Page       int
	TotalPages int
}
`

const TransactionManagerTemplate = `
// txKey is the key used to store the transaction in the context.
type txKey struct{}

// TxFromContext returns the transaction from the context.
func TxFromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(*sql.Tx)
	return tx, ok
}

// TxManager is a transaction manager.
type TxManager struct {
	db *sql.DB
}

// NewTxManager creates a new transaction manager.
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{
		db: db,
	}
}

// Begin begins a transaction.
func (m *TxManager) Begin(ctx context.Context) (context.Context, error) {
	if _, ok := TxFromContext(ctx); ok {
		return ctx, nil
	}

	tx, err := m.db.Begin()
	if err != nil {
		return ctx, fmt.Errorf("could not begin transaction: %w", err)
	}

	// store the transaction in the context.
	return context.WithValue(ctx, txKey{}, tx), nil
}

// IsTxOpen returns true if a transaction is open.
func (m *TxManager) Commit(ctx context.Context) error {
	tx, ok := TxFromContext(ctx)
	if !ok {
		return fmt.Errorf("transactions wasn't opened")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("could not commit transaction: %w", err)
	}

	return nil
}

// Rollback rolls back a transaction.
func (m *TxManager) Rollback(ctx context.Context) error {
	if tx, ok := TxFromContext(ctx); ok {
		err := tx.Rollback()
		if err != nil && !errors.Is(err, sql.ErrTxDone) {
			return err
		}
	}

	return nil
}

// ExecFuncWithTx executes a function with a transaction.
func (m *TxManager) ExecFuncWithTx(ctx context.Context, f func(context.Context) error) error {
	// if a transaction is already open, just execute the function.
	if m.IsTxOpen(ctx) {
		return f(ctx)
	}

	ctx, err := m.Begin(ctx)
	if err != nil {
		return err
	}
	// rollback the transaction if there is an error.
	defer func() { _ = m.Rollback(ctx) }()

	if err := f(ctx); err != nil {
		return err
	}

	// commit the transaction.
	if err := m.Commit(ctx); err != nil {
		return err
	}

	return nil
}

// IsTxOpen returns true if a transaction is open.
func (m *TxManager) IsTxOpen(ctx context.Context) bool {
	_, ok := TxFromContext(ctx)
	return ok
}

// QueryExecer is an interface that can execute queries.
type QueryExecer interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
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
