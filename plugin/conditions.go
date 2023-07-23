package plugin

import (
	"bytes"
	"fmt"
	"text/template"
)

const ConditionTemplate = `
// Condition is a condition filters.
type Condition interface {
	Apply(query sq.SelectBuilder) sq.SelectBuilder
	ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder
}

// PageCondition is a condition that limits the number of rows returned based on the page number.
type PageCondition struct {
	PageSize uint64
	Page     uint64
}

// Page returns a condition that limits the number of rows returned based on the page number.
func Page(pageSize uint64, page uint64) Condition {
	return PageCondition{PageSize: pageSize, Page: page}
}

// Apply applies the condition to the query.
func (c PageCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	// Calculate offset based on the page number
	offset := c.PageSize * (c.Page - 1)
	return query.Limit(c.PageSize).Offset(offset)
}

// ApplyDelete applies the condition to the query.
func (c PageCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	// Calculate offset based on the page number
	offset := c.PageSize * (c.Page - 1)
	return query.Limit(c.PageSize).Offset(offset)
}

// PaginateCondition is a condition that limits the number of rows returned.
type PaginateCondition struct {
	Limit  uint64
	Offset uint64
}

// Paginate returns a condition that limits the number of rows returned.
func Paginate(limit uint64, offset uint64) Condition {
	return PaginateCondition{Limit: limit, Offset: offset}
}

// Apply applies the condition to the query.
func (c PaginateCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Limit(c.Limit).Offset(c.Offset)
}

// ApplyDelete applies the condition to the query.
func (c PaginateCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Limit(c.Limit).Offset(c.Offset)
}

// LimitCondition is a condition that limits the number of rows returned.
type LimitCondition struct {
	Limit uint64
}

// Limit returns a condition that limits the number of rows returned.
func Limit(limit uint64) Condition {
	return LimitCondition{Limit: limit}
}

// Apply applies the condition to the query.
func (c LimitCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Limit(c.Limit)
}

// ApplyDelete applies the condition to the query.
func (c LimitCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Limit(c.Limit)
}

// OffsetCondition is a condition that skips the first n rows.
type OffsetCondition struct {
	Offset uint64
}

// Offset returns a condition that skips the first n rows.
func Offset(offset uint64) Condition {
	return OffsetCondition{Offset: offset}
}

// Apply applies the condition to the query.
func (c OffsetCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Offset(c.Offset)
}

// ApplyDelete applies the condition to the query.
func (c OffsetCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Offset(c.Offset)
}

// And returns a condition that combines the given conditions with AND.
type AndCondition struct {
	Where []Condition
}

// And returns a condition that combines the given conditions with AND.
func And(conditions ...Condition) Condition {
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

// Or returns a condition that checks if any of the conditions are true.
type OrCondition struct {
	Conditions []Condition
}

func Or(conditions ...Condition) Condition {
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

// WhereEq returns a condition that checks if the field equals the value.
func WhereEq(field string, value interface{}) Condition {
	return EqualsCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}Eq returns a condition that checks if the field equals the value.
    func Where{{ $msg }}{{ $field.Name | sToCml }}Eq(value interface{}) Condition {
      return EqualsCondition{Field: "{{ $field.Name }}", Value: value}
    }
  {{ end }}
{{ end }}

// ------------------------------

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

// WhereNotEq returns a condition that checks if the field equals the value.
func WhereNotEq(field string, value interface{}) Condition {
	return NotEqualsCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}NotEq returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}NotEq(value interface{}) Condition {
	  return NotEqualsCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereGreaterThan returns a condition that checks if the field equals the value.
func WhereGreaterThan(field string, value interface{}) Condition {
	return GreaterThanCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}GreaterThan returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}GreaterThan(value interface{}) Condition {
	  return GreaterThanCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereLessThan returns a condition that checks if the field equals the value.
func WhereLessThan(field string, value interface{}) Condition {
	return LessThanCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}LessThan returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}LessThan(value interface{}) Condition {
	  return LessThanCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereGreaterThanOrEqual returns a condition that checks if the field equals the value.
func WhereGreaterThanOrEqual(field string, value interface{}) Condition {
	return GreaterThanOrEqualCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}GreaterThanOrEqual returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}GreaterThanOrEqual(value interface{}) Condition {
	  return GreaterThanOrEqualCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereLessThanOrEqual returns a condition that checks if the field equals the value.
func WhereLessThanOrEqual(field string, value interface{}) Condition {
	return LessThanOrEqualCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}LessThanOrEqual returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}LessThanOrEqual(value interface{}) Condition {
	  return LessThanOrEqualCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereLike returns a condition that checks if the field equals the value.
func WhereLike(field string, value interface{}) Condition {
	return LikeCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}Like returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}Like(value interface{}) Condition {
	  return LikeCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereNotLike returns a condition that checks if the field equals the value.
func WhereNotLike(field string, value interface{}) Condition {
	return NotLikeCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}NotLike returns a condition that checks if the field equals the value.
	func Where{{ $msg }}{{ $field.Name | sToCml }}NotLike(value interface{}) Condition {
	  return NotLikeCondition{Field: "{{ $field.Name }}", Value: value}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereIsNull returns a condition that checks if the field is null.
func WhereIsNull(field string) Condition {
	return IsNullCondition{Field: field}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}IsNull returns a condition that checks if the field is null.
	func Where{{ $msg }}{{ $field.Name | sToCml }}IsNull() Condition {
	  return IsNullCondition{Field: "{{ $field.Name }}"}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereIsNotNull returns a condition that checks if the field is not null.
func WhereIsNotNull(field string) Condition {
	return IsNotNullCondition{Field: field}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}IsNotNull returns a condition that checks if the field is not null.
	func Where{{ $msg }}{{ $field.Name | sToCml }}IsNotNull() Condition {
	  return IsNotNullCondition{Field: "{{ $field.Name }}"}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereIn returns a condition that checks if the field is in the given values.
func WhereIn(field string, values ...interface{}) Condition {
	return InCondition{Field: field, Values: values}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}In returns a condition that checks if the field is in the given values.
	func Where{{ $msg }}{{ $field.Name | sToCml }}In(values ...interface{}) Condition {
	  return InCondition{Field: "{{ $field.Name }}", Values: values}
	}
  {{ end }}
{{ end }}

// --------------------------------

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

// WhereNotIn returns a condition that checks if the field is not in the given values.
func WhereNotIn(field string, values ...interface{}) Condition {
	return NotInCondition{Field: field, Values: values}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}NotIn returns a condition that checks if the field is not in the given values.
	func Where{{ $msg }}{{ $field.Name | sToCml }}NotIn(values ...interface{}) Condition {
	  return NotInCondition{Field: "{{ $field.Name }}", Values: values}
	}
  {{ end }}
{{ end }}

// --------------------------------

// BetweenCondition represents the BETWEEN condition.
type BetweenCondition struct {
	Field string
	From  interface{}
	To    interface{}
}

// Apply applies the condition to the query.
func (c BetweenCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.From, c.To))
}

// ApplyDelete applies the condition to the query.
func (c BetweenCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s BETWEEN ? AND ?", c.Field), c.From, c.To))
}

// WhereBetween returns a condition that checks if the field is between the given values.
func WhereBetween(field string, from, to interface{}) Condition {
	return BetweenCondition{Field: field, From: from, To: to}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}Between returns a condition that checks if the field is between the given values.
	func Where{{ $msg }}{{ $field.Name | sToCml }}Between(from, to interface{}) Condition {
	  return BetweenCondition{Field: "{{ $field.Name }}", From: from, To: to}
	}
  {{ end }}
{{ end }}

// --------------------------------

// OrderCondition represents the ORDER BY condition.
type OrderCondition struct {
	Column string
	Asc    bool
}

// WhereOrderBy applies the condition to the query.
func WhereOrderBy(column string, asc bool) Condition {
	return OrderCondition{Column: column, Asc: asc}
}

// Apply applies the condition to the query.
func (c OrderCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	if c.Asc {
		return query.OrderBy(c.Column + " ASC")
	}
	return query.OrderBy(c.Column + " DESC")
}

func (c OrderCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// Where{{ $msg }}{{ $field.Name | sToCml }}OrderBy returns a condition that orders the query by the given column.
	func Where{{ $msg }}{{ $field.Name | sToCml }}OrderBy(asc bool) Condition {
	  return OrderCondition{Column: "{{ $field.Name }}", Asc: asc}
	}
  {{ end }}
{{ end }}

// --------------------------------

// DateAfterCondition represents the '>' condition for dates.
type DateAfterCondition struct {
	Field string
	Date  time.Time
}

// Apply applies the condition to the query.
func (c DateAfterCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s > $1", c.Field), c.Date))
}

// ApplyDelete applies the condition to the query.
func (c DateAfterCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s > $1", c.Field), c.Date))
}

// WhereDateAfter returns a condition that checks if the field is after the given date.
func WhereDateAfter(field string, date time.Time) Condition {
	return DateAfterCondition{Field: field, Date: date}
}

// DateBeforeCondition represents the '<' condition for dates.
type DateBeforeCondition struct {
	Field string
	Date  time.Time
}

// Apply applies the condition to the query.
func (c DateBeforeCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s < $1", c.Field), c.Date))
}

// ApplyDelete applies the condition to the query.
func (c DateBeforeCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s < $1", c.Field), c.Date))
}

// WhereDateBefore returns a condition that checks if the field is before the given date.
func WhereDateBefore(field string, date time.Time) Condition {
	return DateBeforeCondition{Field: field, Date: date}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	{{ if or (eq $field.Type "time.Time") (eq $field.Type "*time.Time") }}
		// Where{{ $msg }}{{ $field.Name | sToCml }}After returns a condition that checks if the field is after the given date.
		func Where{{ $msg }}{{ $field.Name | sToCml }}After(date time.Time) Condition {
			return DateAfterCondition{Field: "{{ $field.Name }}", Date: date}
		}
		// Where{{ $msg }}{{ $field.Name | sToCml }}Before returns a condition that checks if the field is before the given date.
		func Where{{ $msg }}{{ $field.Name | sToCml }}Before(date time.Time) Condition {
			return DateBeforeCondition{Field: "{{ $field.Name }}", Date: date}
		}
	{{ end }}
  {{ end }}
{{ end }}

// --------------------------------
// JSON
// --------------------------------

// JSONExistsCondition	exists condition.
type JSONExistsCondition struct {
	Field string
	Key   string
}

// Apply applies the condition to the query.
func (c JSONExistsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s -> '%s' IS NOT NULL", c.Field, c.Key)))
}

// ApplyDelete applies the condition to the query.
func (c JSONExistsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s -> '%s' IS NOT NULL", c.Field, c.Key)))
}

// JSONEqualsCondition equals condition.
type JSONEqualsCondition struct {
	Field string
	Key   string
	Value interface{}
}

// Apply applies the condition to the query.
func (c JSONEqualsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s ->> '%s' = ?", c.Field, c.Key), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c JSONEqualsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s ->> '%s' = ?", c.Field, c.Key), c.Value))
}

// JSONContainsCondition contains condition.
type JSONContainsCondition struct {
	Field string
	Value string // This should be a JSON string
}

// Apply applies the condition to the query.
func (c JSONContainsCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s @> ?", c.Field), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c JSONContainsCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s @> ?", c.Field), c.Value))
}

// JSONContainedInCondition contained in condition.
type JSONContainedInCondition struct {
	Field string
	Value string // This should be a JSON string
}

// Apply applies the condition to the query.
func (c JSONContainedInCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s <@ ?", c.Field), c.Value))
}

// ApplyDelete applies the condition to the query.
func (c JSONContainedInCondition) ApplyDelete(query sq.DeleteBuilder) sq.DeleteBuilder {
	return query.Where(sq.Expr(fmt.Sprintf("%s <@ ?", c.Field), c.Value))
}

// WhereJSONExists returns a condition that checks if the JSON field contains the given key.
func WhereJSONExists(field string, key string) Condition {
	return JSONExistsCondition{Field: field, Key: key}
}

// WhereJSONEquals returns a condition that checks if the JSON field's key equals to the given value.
func WhereJSONEquals(field string, key string, value interface{}) Condition {
	return JSONEqualsCondition{Field: field, Key: key, Value: value}
}

// WhereJSONContains returns a condition that checks if the JSON field contains the given JSON value.
func WhereJSONContains(field string, value string) Condition {
	return JSONContainsCondition{Field: field, Value: value}
}

// WhereJSONContainedIn returns a condition that checks if the JSON field is contained in the given JSON value.
func WhereJSONContainedIn(field string, value string) Condition {
	return JSONContainedInCondition{Field: field, Value: value}
}
`

// BuildConditionsTemplate builds the conditions template.
func (p *Plugin) BuildConditionsTemplate() string {
	type field struct {
		Name string
		Type string
	}

	type TemplateData struct {
		Plugin   *Plugin
		Messages map[string][]field
	}

	data := TemplateData{
		Plugin:   p,
		Messages: make(map[string][]field),
	}

	for _, m := range p.state.Tables {
		for _, f := range m.GetField() {
			if !checkIsRelation(f) {
				data.Messages[m.GetName()] = append(data.Messages[m.GetName()], field{
					Name: f.GetName(),
					Type: convertType(f),
				})
			}
		}
	}

	var output bytes.Buffer

	funcs := template.FuncMap{
		"upperClientName": upperClientName,
		"lowerClientName": lowerClientName,
		"sToCml":          sToCml,
		"sToLowerCamel":   sToLowerCamel,
	}

	tmpl, err := template.New("goFile").Funcs(funcs).Parse(ConditionTemplate)
	if err != nil {
		fmt.Println(err)
		return ""
	}

	if err = tmpl.Execute(&output, data); err != nil {
		fmt.Println(err)
		return ""
	}

	// enable imports
	p.state.Imports.Enable(ImportSquirrel, ImportFMT, ImportStrings)

	return output.String()
}
