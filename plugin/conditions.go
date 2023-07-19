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

// Or returns a condition that checks if any of the conditions are true.
type OrCondition struct {
	Where []Condition
}

// Apply applies the condition to the query.
func Or(conditions ...Condition) Condition {
	return OrCondition{Where: conditions}
}

// Apply applies the condition to the query.
func (c OrCondition) Apply(query sq.SelectBuilder) sq.SelectBuilder {
	or := sq.Or{}
	for _, condition := range c.Where {
		condQuery := condition.Apply(query)
		sql, args, _ := condQuery.ToSql()
		{
			or = append(or, sq.Expr(sql, args...))
		}
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

// Eq returns a condition that checks if the field equals the value.
func Eq(field string, value interface{}) Condition {
	return EqualsCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}Eq returns a condition that checks if the field equals the value.
    func {{ $msg }}{{ $field | sToCml }}Eq(value interface{}) Condition {
      return EqualsCondition{Field: "{{ $field }}", Value: value}
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

// NotEq returns a condition that checks if the field equals the value.
func NotEq(field string, value interface{}) Condition {
	return NotEqualsCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}NotEq returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}NotEq(value interface{}) Condition {
	  return NotEqualsCondition{Field: "{{ $field }}", Value: value}
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

// GreaterThan returns a condition that checks if the field equals the value.
func GreaterThan(field string, value interface{}) Condition {
	return GreaterThanCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}GreaterThan returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}GreaterThan(value interface{}) Condition {
	  return GreaterThanCondition{Field: "{{ $field }}", Value: value}
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

// LessThan returns a condition that checks if the field equals the value.
func LessThan(field string, value interface{}) Condition {
	return LessThanCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}LessThan returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}LessThan(value interface{}) Condition {
	  return LessThanCondition{Field: "{{ $field }}", Value: value}
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

// GreaterThanOrEqual returns a condition that checks if the field equals the value.
func GreaterThanOrEqual(field string, value interface{}) Condition {
	return GreaterThanOrEqualCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}GreaterThanOrEqual returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}GreaterThanOrEqual(value interface{}) Condition {
	  return GreaterThanOrEqualCondition{Field: "{{ $field }}", Value: value}
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

func LessThanOrEqual(field string, value interface{}) Condition {
	return LessThanOrEqualCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}LessThanOrEqual returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}LessThanOrEqual(value interface{}) Condition {
	  return LessThanOrEqualCondition{Field: "{{ $field }}", Value: value}
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

// Like returns a condition that checks if the field equals the value.
func Like(field string, value interface{}) Condition {
	return LikeCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}Like returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}Like(value interface{}) Condition {
	  return LikeCondition{Field: "{{ $field }}", Value: value}
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

// NotLike returns a condition that checks if the field equals the value.
func NotLike(field string, value interface{}) Condition {
	return NotLikeCondition{Field: field, Value: value}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}NotLike returns a condition that checks if the field equals the value.
	func {{ $msg }}{{ $field | sToCml }}NotLike(value interface{}) Condition {
	  return NotLikeCondition{Field: "{{ $field }}", Value: value}
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

// IsNull returns a condition that checks if the field is null.
func IsNull(field string) Condition {
	return IsNullCondition{Field: field}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}IsNull returns a condition that checks if the field is null.
	func {{ $msg }}{{ $field | sToCml }}IsNull() Condition {
	  return IsNullCondition{Field: "{{ $field }}"}
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

// IsNotNull returns a condition that checks if the field is not null.
func IsNotNull(field string) Condition {
	return IsNotNullCondition{Field: field}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}IsNotNull returns a condition that checks if the field is not null.
	func {{ $msg }}{{ $field | sToCml }}IsNotNull() Condition {
	  return IsNotNullCondition{Field: "{{ $field }}"}
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

// In returns a condition that checks if the field is in the given values.
func In(field string, values ...interface{}) Condition {
	return InCondition{Field: field, Values: values}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}In returns a condition that checks if the field is in the given values.
	func {{ $msg }}{{ $field | sToCml }}In(values ...interface{}) Condition {
	  return InCondition{Field: "{{ $field }}", Values: values}
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

// NotIn returns a condition that checks if the field is not in the given values.
func NotIn(field string, values ...interface{}) Condition {
	return NotInCondition{Field: field, Values: values}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}NotIn returns a condition that checks if the field is not in the given values.
	func {{ $msg }}{{ $field | sToCml }}NotIn(values ...interface{}) Condition {
	  return NotInCondition{Field: "{{ $field }}", Values: values}
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

// Between returns a condition that checks if the field is between the given values.
func Between(field string, from, to interface{}) Condition {
	return BetweenCondition{Field: field, From: from, To: to}
}

{{ range $msg, $fields := .Messages }}
  {{ range $field := $fields }}
	// {{ $msg }}{{ $field | sToCml }}Between returns a condition that checks if the field is between the given values.
	func {{ $msg }}{{ $field | sToCml }}Between(from, to interface{}) Condition {
	  return BetweenCondition{Field: "{{ $field }}", From: from, To: to}
	}
  {{ end }}
{{ end }}
`

func (p *Plugin) BuildConditionsTemplate() string {
	type TemplateData struct {
		Plugin   *Plugin
		Messages map[string][]string
	}

	data := TemplateData{
		Plugin:   p,
		Messages: map[string][]string{},
	}

	for _, m := range getMessages(p.req) {
		for _, f := range m.GetField() {
			data.Messages[m.GetName()] = append(data.Messages[m.GetName()], f.GetName())
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
	p.imports.Enable(ImportSquirrel, ImportFMT)

	return output.String()
}
