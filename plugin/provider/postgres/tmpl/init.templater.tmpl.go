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
{{ range $key, $value := storages }}
{{ $key }} {{ $value }}{{ end }}
}

// {{ storageName }} is the interface for the {{ storageName }}.
type {{ storageName }} interface { {{ range $key, $value := storages }}
	// Get{{ $value }} returns the {{ $value }} store.
	Get{{ $value }}() {{ $value }}{{ end }}
	// TxManager returns the transaction manager.
	TxManager() *TxManager
	// CreateTables creates the tables for all the stores.
	CreateTables() error
	// DropTables drops the tables for all the stores.
	DropTables() error
	// TruncateTables truncates the tables for all the stores.
	TruncateTables() error
	// UpgradeTables upgrades the tables for all the stores.
	UpgradeTables() error
}

// New{{ storageName }} returns a new {{ storageName }}.
func New{{ storageName }}(db *sql.DB) {{ storageName }} {
	return &{{ storageName | lowerCamelCase }}{
		db: db,
		tx: NewTxManager(db),
{{ range $key, $value := storages }}
{{ $key }}: New{{ $value }}(db),{{ end }}
	}
}

// TxManager returns the transaction manager.
func (c *{{ storageName | lowerCamelCase }}) TxManager() *TxManager {
	return c.tx
}

{{ range $key, $value := storages }}
// Get{{ $value }} returns the {{ $value }} store.
func (c *{{ storageName | lowerCamelCase }}) Get{{ $value }}() {{ $value }} {
	return c.{{ $key }}
}
{{ end }}

// CreateTables creates the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) CreateTables() error {
	var err error
{{ range $key, $value := storages }}
	// create the {{ $value }} table.
	err = c.{{ $key }}.CreateTable()
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
{{ end }}
	return nil
}

// DropTables drops the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) DropTables() error {
	var err error
{{ range $key, $value := storages }}
	// drop the {{ $value }} table.
	err = c.{{ $key }}.DropTable()
	if err != nil {
		return fmt.Errorf("failed to drop table: %w", err)
	}
{{ end }}
	return nil
}

// TruncateTables truncates the tables for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) TruncateTables() error {
	var err error
{{ range $key, $value := storages }}
	// truncate the {{ $value }} table.
	err = c.{{ $key }}.TruncateTable()
	if err != nil {
		return fmt.Errorf("failed to truncate table: %w", err)
	}
{{ end }}
	return nil
}

// UpgradeTables runs the database upgrades for all the stores.
// This is idempotent and safe to run multiple times.
func (c *{{ storageName | lowerCamelCase }}) UpgradeTables() error {
	var err error
{{ range $key, $value := storages }}
	// run the {{ $value }} upgrade.
	err = c.{{ $key }}.UpgradeTable()
	if err != nil {
		return fmt.Errorf("failed to upgrade: %w", err)
	}
{{ end }}
	return nil
}
`

const TypesTemplate = `
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

	return fmt.Errorf("can't convert %T", src)
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
type {{ $field.FieldType }} struct {
	{{ $field.SourceName | camelCase }} {{ $field.Descriptor | fieldType }} ` + "`json:\"{{ $field.SourceName }}\"`" + `
}

// New{{ $field.SourceName | camelCase }}Field returns a new {{ $field.FieldType }}.
func New{{ $field.SourceName | camelCase }}Field (v {{ $field.Descriptor | fieldType }}) *{{ $field.FieldType }} {
	return &{{ $field.FieldType }}{
		{{ $field.SourceName | camelCase }}: v,
	}
}

// Scan implements the sql.Scanner interface for JSON.
func (m *{{ $field.FieldType }}) Scan(src interface{}) error  {
	if bytes, ok := src.([]byte); ok {
		return json.Unmarshal(bytes, m)
	}

	return fmt.Errorf("can't convert %T", src)
}

// Value implements the driver.Valuer interface for JSON.
func (m *{{ $field.FieldType }}) Value() (driver.Value, error) {
	if m == nil {
		m = &{{ $field.FieldType }}{}
	}
	return json.Marshal(m)
}

// Get returns the value of the field.
func (m *{{ $field.FieldType }}) Get() {{ $field.Descriptor | fieldType }} {
	return m.{{ $field.SourceName | camelCase }}
}

func (m *{{ $field.FieldType }}) String() string {
	return fmt.Sprintf("%v", m.Get())
}

{{ end }}
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
		return errors.New("transactions wasn't opened")
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
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
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
)
`
