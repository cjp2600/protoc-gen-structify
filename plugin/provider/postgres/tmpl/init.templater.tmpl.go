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

// WithPagination sets the pagination for the query.
func (b *QueryBuilder) WithPagination(pagination *Pagination) *QueryBuilder {
	b.pagination = pagination
	return b
}

// Filter is a helper function to create a new query builder with filter options.
func FilterBuilder(filterOptions ...FilterApplier) *QueryBuilder {
	return NewQueryBuilder().WithFilter(filterOptions...)
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
	CreateTables(ctx context.Context) error
	// DropTables drops the tables for all the stores.
	DropTables(ctx context.Context) error
	// TruncateTables truncates the tables for all the stores.
	TruncateTables(ctx context.Context) error
	// UpgradeTables upgrades the tables for all the stores.
	UpgradeTables(ctx context.Context) error
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
func (c *{{ storageName | lowerCamelCase }}) CreateTables(ctx context.Context) error {
	var err error
{{ range $key, $value := storages }}
	// create the {{ $value }} table.
	err = c.{{ $key }}.CreateTable(ctx)
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
{{ range $key, $value := storages }}
	// drop the {{ $value }} table.
	err = c.{{ $key }}.DropTable(ctx)
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
{{ range $key, $value := storages }}
	// truncate the {{ $value }} table.
	err = c.{{ $key }}.TruncateTable(ctx)
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
{{ range $key, $value := storages }}
	// run the {{ $value }} upgrade.
	err = c.{{ $key }}.UpgradeTable(ctx)
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
)
`

const TableConditionsTemplate = `
// And returns a condition that combines the given conditions with AND.
type AndCondition struct {
	Where []FilterApplier
}

// And returns a condition that combines the given conditions with AND.
func And(conditions ...FilterApplier) FilterApplier {
	return AndCondition{Where: conditions}
}

// And returns a condition that combines the given conditions with AND.
func (c AndCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	for _, condition := range c.Where {
		query = condition.Apply(query)
	}
	return query
}

// And returns a condition that combines the given conditions with AND.
func (c AndCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	for _, condition := range c.Where {
		query = condition.ApplyDelete(query)
	}
	return query
}

//
// Or returns a condition that checks if any of the conditions are true.
//

// Or returns a condition that checks if any of the conditions are true.
type OrCondition struct {
	Conditions []FilterApplier
}
// Or returns a condition that checks if any of the conditions are true.
func Or(conditions ...FilterApplier) FilterApplier {
	return OrCondition{Conditions: conditions}
}

// Apply applies the condition to the query.
func (c OrCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	or := sq.Or{}
	for _, condition := range c.Conditions {
		subQuery := condition.Apply(sq.Select("*"))
		// Extract WHERE clause from the subquery
		whereParts, args, _ := subQuery.ToSql()
		whereParts = strings.TrimPrefix(whereParts, "SELECT * WHERE ")
		// Append the WHERE clause to the OR condition
		or = append(or, sq.Expr(whereParts, args...))
	}
	return query.Where(or)
}

// Apply applies the condition to the query.
func (c OrCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	or := sq.Or{}
	for _, condition := range c.Conditions {
		subQuery := condition.Apply(sq.Select("*"))
		// Extract WHERE clause from the subquery
		whereParts, args, _ := subQuery.ToSql()
		whereParts = strings.TrimPrefix(whereParts, "SELECT * WHERE ")
		// Append the WHERE clause to the OR condition
		or = append(or, sq.Expr(whereParts, args...))
	}
	return query.Where(or)
}

// EqualsCondition equals condition.
type EqualsCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c EqualsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Eq{c.Field: c.Value})
}

func (c EqualsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Eq{c.Field: c.Value})
}

// Eq returns a condition that checks if the field equals the value.
func Eq(field string, value interface{}) FilterApplier {
	return EqualsCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Eq returns a condition that checks if the field equals the value.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Eq(value {{ $field | fieldType }}) FilterApplier {
      return EqualsCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// NotEqualsCondition not equals condition.
type NotEqualsCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c NotEqualsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c NotEqualsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotEq{c.Field: c.Value})
}

// NotEq returns a condition that checks if the field equals the value.
func NotEq(field string, value interface{}) FilterApplier {
	return NotEqualsCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotEq returns a condition that checks if the field equals the value.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotEq(value {{ $field | fieldType }}) FilterApplier {
      return NotEqualsCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// GreaterThanCondition greaterThanCondition than condition.
type GreaterThanCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c GreaterThanCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Gt{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c GreaterThanCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Gt{c.Field: c.Value})
}

// GreaterThan returns a condition that checks if the field equals the value.
func GreaterThan(field string, value interface{}) FilterApplier {
	return GreaterThanCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GreaterThan greaterThanCondition than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GreaterThan(value {{ $field | fieldType }}) FilterApplier {
      return GreaterThanCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// LessThanCondition less than condition.
type LessThanCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LessThanCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Lt{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LessThanCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Lt{c.Field: c.Value})
}

// LessThan returns a condition that checks if the field equals the value.
func LessThan(field string, value interface{}) FilterApplier {
	return LessThanCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LessThan less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LessThan(value {{ $field | fieldType }}) FilterApplier {
      return LessThanCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// LessThanOrEqualCondition less than or equal condition.
type GreaterThanOrEqualCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c GreaterThanOrEqualCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.GtOrEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c GreaterThanOrEqualCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.GtOrEq{c.Field: c.Value})
}

// GreaterThanOrEqual returns a condition that checks if the field equals the value.
func GreaterThanOrEq(field string, value interface{}) FilterApplier {
	return GreaterThanOrEqualCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GreaterThanOrEq less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}GreaterThanOrEq(value {{ $field | fieldType }}) FilterApplier {
      return GreaterThanOrEqualCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// LessThanOrEqualCondition less than or equal condition.
type LessThanOrEqualCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LessThanOrEqualCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.LtOrEq{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LessThanOrEqualCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.LtOrEq{c.Field: c.Value})
}

func LessThanOrEq(field string, value interface{}) FilterApplier {
	return LessThanOrEqualCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LessThanOrEq less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}LessThanOrEq(value {{ $field | fieldType }}) FilterApplier {
      return LessThanOrEqualCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// LikeCondition like condition.
type LikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c LikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Like{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c LikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Like{c.Field: c.Value})
}

// Like returns a condition that checks if the field equals the value.
func Like(field string, value interface{}) FilterApplier {
	return LikeCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Like less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}Like(value {{ $field | fieldType }}) FilterApplier {
      return LikeCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// NotLikeCondition not like condition.
type NotLikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c NotLikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotLike{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c NotLikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotLike{c.Field: c.Value})
}

// NotLike returns a condition that checks if the field equals the value.
func NotLike(field string, value interface{}) FilterApplier {
	return NotLikeCondition{Field: field, Value: value}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotLike less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotLike(value {{ $field | fieldType }}) FilterApplier {
      return NotLikeCondition{Field: "{{ $field.GetName }}", Value: value}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// IsNullCondition represents the IS NULL condition.
type IsNullCondition struct {
	Field string
}

// Apply applies the condition to the query.
func (c IsNullCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(c.Field + " IS NULL"))
}

// ApplyDelete applies the condition to the query.
func (c IsNullCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(c.Field + " IS NULL"))
}

// IsNull returns a condition that checks if the field is null.
func IsNull(field string) FilterApplier {
	return IsNullCondition{Field: field}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNull() FilterApplier {
      return IsNullCondition{Field: "{{ $field.GetName }}"}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// IsNotNullCondition represents the IS NOT NULL condition.
type IsNotNullCondition struct {
	Field string
}

// Apply applies the condition to the query.
func (c IsNotNullCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(c.Field + " IS NOT NULL"))
}

// ApplyDelete applies the condition to the query.
func (c IsNotNullCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(c.Field + " IS NOT NULL"))
}

// IsNotNull returns a condition that checks if the field is not null.
func IsNotNull(field string) FilterApplier {
	return IsNotNullCondition{Field: field}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}IsNotNull() FilterApplier {
      return IsNotNullCondition{Field: "{{ $field.GetName }}"}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// InCondition represents the IN condition.
type InCondition struct {
	Field  string
	Values []interface{}
}

// Apply applies the condition to the query.
func (c InCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Eq{c.Field: c.Values})
}

// ApplyDelete applies the condition to the query.
func (c InCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Eq{c.Field: c.Values})
}

// In returns a condition that checks if the field is in the given values.
func In(field string, values ...interface{}) FilterApplier {
	return InCondition{Field: field, Values: values}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}In less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}In(values ...interface{}) FilterApplier {
      return InCondition{Field: "{{ $field.GetName }}", Values: values}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}

// NotInCondition represents the NOT IN condition.
type NotInCondition struct {
	Field  string
	Values []interface{}
}

// Apply applies the condition to the query.
func (c NotInCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.NotEq{c.Field: c.Values})
}

// ApplyDelete applies the condition to the query.
func (c NotInCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.NotEq{c.Field: c.Values})
}

// NotIn returns a condition that checks if the field is not in the given values.
func NotIn(field string, values ...interface{}) FilterApplier {
	return NotInCondition{Field: field, Values: values}
}

{{ range $key, $fieldMess := messages }}
  {{ range $field := $fieldMess.GetField }}
   {{- if not ($field | isRelation) }}
   {{- if not ($field | isJSON) }}
	// {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotIn less than condition.
    func {{ $fieldMess.GetName | camelCase }}{{ $field.GetName | camelCase }}NotIn(values ...interface{}) FilterApplier {
      return NotInCondition{Field: "{{ $field.GetName }}", Values: values}
    }
  {{ end }}
  {{ end }}
  {{ end }}
{{ end }}
`
