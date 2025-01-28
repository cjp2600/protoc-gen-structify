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
// Dsn builds the DSN string for the database connection.
// See https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING
func Dsn(host string, port int, user string, password string, dbname string, sslMode string, timeout int) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s", host, port, user, password, dbname)

	// sslMode is optional. If not provided, it will be omitted.
 	// valid values are: disable, allow, prefer, require, verify-ca, verify-full
	if sslMode != "" {
		dsn += " sslmode=" + sslMode
	}

	if timeout != 0 {
		dsn += " connect_timeout=" + strconv.Itoa(timeout)
	}

	return dsn
}

// Open connects to the database and returns a *sql.DB.
// The caller is responsible for closing the database.
// The caller is responsible for calling db.Ping() to verify the connection.
// The caller is responsible for setting the connection pool options.
// 
// See https://golang.org/pkg/database/sql/#DB.SetMaxOpenConns
// See https://golang.org/pkg/database/sql/#DB.SetMaxIdleConns
// See https://golang.org/pkg/database/sql/#DB.SetConnMaxIdleTime
// See https://golang.org/pkg/database/sql/#DB.SetConnMaxLifetime
// See https://golang.org/pkg/database/sql/#DB.Ping
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

    db, err := sql.Open("postgres", dsn)
    if err != nil {
        return nil, errors.Wrap(err, "failed to open database")
    }

	// Ping verifies a connection to the database is still alive, establishing a connection if necessary.
	if err = db.Ping(); err != nil {
		// If Ping fails, close the DB and return an error.
		db.Close() // Ignoring error from Close, as we already have a more significant error.
		return nil, errors.Wrap(err, "failed to ping database")
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
	db *DB // The database connection.
	tx *TxManager // The transaction manager.
{{ range $value := storages }}
{{ $value.Key }} {{ $value.Value }}{{ end }}
}

type DB struct {
	DBRead *sql.DB
	DBWrite *sql.DB
}

// {{ storageName }} is the interface for the {{ storageName }}.
type {{ storageName }} interface { 
	{{- range $value := storages }}
	// Get{{ $value.Value }} returns the {{ $value.Value }} store.
	Get{{ $value.Value }}() {{ $value.Value }}
	{{- end }}
	// TxManager returns the transaction manager.
	TxManager() *TxManager

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
func New{{ storageName }}(db *DB) {{ storageName }} {
	if db == nil {
		panic("structify: db is required")
	}

	if db.DBRead == nil {
		panic("structify: dbRead is required")
	}

	if db.DBWrite == nil {
		db.DBWrite = db.DBRead
	}

	return &{{ storageName | lowerCamelCase }}{
		db: db,
		tx: NewTxManager(db.DBWrite),
{{ range $value := storages }}
{{ $value.Key }}: New{{ $value.Value }}(db),{{ end }}
	}
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
	Data T
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

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to convert value to []byte")
	}

	if err := json.Unmarshal(bytes, &n.Data); err != nil {
		n.Valid = false
		return errors.Wrap(err, "failed to unmarshal json")
	}

	n.Valid = true
	return nil
}

func (n *NullableJSON[T]) Value() (driver.Value, error) {
	if !n.Valid {
		return nil, nil
	}

	return json.Marshal(n.Data)
}

// ValueOrZero returns the value if valid, otherwise returns the zero value of type T.
func (n NullableJSON[T]) ValueOrZero() T {
	if !n.Valid {
		var zero T // This declares a variable of type T initialized to its zero value
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
func (m *{{ $field.StructureName }}) Scan(src interface{}) error  {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return errors.New(fmt.Sprintf("can't convert %T", src))
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

	return errors.New(fmt.Sprintf("can't convert %T", src))
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
		return ctx, errors.Wrap(err, "could not begin transaction")
	}

	// store the transaction in the context.
	return context.WithValue(ctx, txKey{}, tx), nil
}

// IsTxOpen returns true if a transaction is open.
func (m *TxManager) Commit(ctx context.Context) error {
	tx, ok := TxFromContext(ctx)
	if !ok {
		return errors.New("transactions wasn't opened")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "could not commit transaction")
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

// IsPgCheckViolation returns true if the error is a postgres check violation.
func IsPgUniqueViolation(err error) bool {
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}

	return pgErr.Code == errPgUniqueViolationError
}

// IsPgCheckViolation returns true if the error is a postgres check violation.
func IsPgViolationError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	if !ok {
		return false
	}
	
	return pgErr.Code == errPgCheckViolation ||
		pgErr.Code == errPgNotNullViolation ||
		pgErr.Code == errPgForeignKeyViolation ||
		pgErr.Code == errPgUniqueViolationError
}

// PgPrettyErr returns a pretty postgres error.
func PgPrettyErr(err error) error {
	if pgErr, ok := err.(*pq.Error); ok {
		return errors.New(pgErr.Detail)
	}
	return err
}

// errors for postgres.
// https://www.postgresql.org/docs/9.3/errcodes-appendix.html
const (
	errPgCheckViolation       = "23514"
	errPgNotNullViolation     = "23502"
	errPgForeignKeyViolation  = "23503"
	errPgUniqueViolationError = "23505"
)
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
