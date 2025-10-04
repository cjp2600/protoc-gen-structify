package tmpl

const TableConditionsTemplate = `
type Table interface {
	TableName() string
}

type JoinType string

const (
	LeftJoin  JoinType = "LEFT"
	InnerJoin JoinType = "INNER"
	RightJoin JoinType = "RIGHT"
)

type JoinCondition struct {
	Type  JoinType
	Table Table
	On    FilterApplier
}

func Join(joinType JoinType, table Table, on FilterApplier) FilterApplier {
	return JoinCondition{Type: joinType, Table: table, On: on}
}

func toInterface[T any](s []T) []interface{} {
	result := make([]interface{}, len(s))
	for i, v := range s {
		result[i] = v
	}
	return result
}

func (c JoinCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	onQuery := c.On.Apply(sq.Select("*"))
	onClause, args, _ := onQuery.ToSql()
	onClause = strings.TrimPrefix(onClause, "SELECT * WHERE ")
	joinExpr := fmt.Sprintf("%s JOIN %s ON %s", c.Type, c.Table.TableName(), onClause)
	return query.JoinClause(sq.Expr(joinExpr, args...))
}

func (c JoinCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}

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

// BetweenCondition 
type BetweenCondition struct {
	Field string
	Min   interface{}
	Max   interface{}
}

// Apply applies the condition to the query.
func (c BetweenCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.Min, c.Max))
}

// ApplyDelete applies the condition to the query.
func (c BetweenCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.Min, c.Max))
}

// Between returns a condition that checks if the field is between the min and max values.
func Between(field string, min, max interface{}) FilterApplier {
	return BetweenCondition{Field: field, Min: min, Max: max}
}

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

// ILikeCondition ilike condition.
type ILikeCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c ILikeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.ILike{c.Field: c.Value})
}

// ApplyDelete applies the condition to the query.
func (c ILikeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.ILike{c.Field: c.Value})
}

// ILike returns a condition that checks if the field equals the value.
func ILike(field string, value interface{}) FilterApplier {
	return ILikeCondition{Field: field, Value: value}
}

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

// OrderCondition represents the ORDER BY condition.
type OrderCondition struct {
	Column string
	Asc    bool
}

// Apply applies the condition to the query.
func OrderBy(column string, asc bool) FilterApplier {
	return OrderCondition{Column: column, Asc: asc}
}

// Apply applies the condition to the query.
func (c OrderCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if c.Asc {
		return query.OrderBy(c.Column + " ASC")
	}

	// default to descending.
	return query.OrderBy(c.Column + " DESC")
}

// ApplyDelete applies the condition to the query.
func (c OrderCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}

// ArrayOverlapCondition represents the array overlap condition (&&).
type ArrayOverlapCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c ArrayOverlapCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s && ?", c.Field), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c ArrayOverlapCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s && ?", c.Field), c.Value))
}

// ArrayOverlap returns a condition that checks if the array field overlaps with the given value.
func ArrayOverlap(field string, value interface{}) FilterApplier {
	return ArrayOverlapCondition{Field: field, Value: value}
}

// ArrayContainsCondition represents the array contains condition (@>).
type ArrayContainsCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c ArrayContainsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s @> ?", c.Field), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c ArrayContainsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s @> ?", c.Field), c.Value))
}

// ArrayContains returns a condition that checks if the array field contains the given value.
func ArrayContains(field string, value interface{}) FilterApplier {
	return ArrayContainsCondition{Field: field, Value: value}
}

// ArrayContainedByCondition represents the array contained by condition (<@).
type ArrayContainedByCondition struct {
	Field string
	Value interface{}
}

// Apply applies the condition to the query.
func (c ArrayContainedByCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s <@ ?", c.Field), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c ArrayContainedByCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s <@ ?", c.Field), c.Value))
}

// ArrayContainedBy returns a condition that checks if the array field is contained by the given value.
func ArrayContainedBy(field string, value interface{}) FilterApplier {
	return ArrayContainedByCondition{Field: field, Value: value}
}

// CursorPaginationCondition represents cursor-based pagination condition.
type CursorPaginationCondition struct {
	Fields []string
	Values []interface{}
}

// Apply applies the condition to the query.
func (c CursorPaginationCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if len(c.Fields) == 0 || len(c.Values) == 0 {
		return query
	}
	
	// Build tuple comparison: (field1, field2, ...) < (value1, value2, ...)
	tupleFields := "(" + strings.Join(c.Fields, ", ") + ")"
	tupleValues := "(" + strings.Repeat("?,", len(c.Values))
	tupleValues = tupleValues[:len(tupleValues)-1] + ")" // Remove last comma
	
	return query.Where(sq.Expr(fmt.Sprintf("%s < %s", tupleFields, tupleValues), c.Values...))
}

// ApplyDelete applies the condition to the query.
func (c CursorPaginationCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}

// CursorPagination returns a condition for cursor-based pagination.
func CursorPagination(fields []string, values []interface{}) FilterApplier {
	return CursorPaginationCondition{Fields: fields, Values: values}
}

`
