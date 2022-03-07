package orm

import (
	"fmt"
	"strings"
)

const (
	queryType_SELECT = iota + 1
	queryType_UPDATE
	queryType_Delete
)

type QueryBuilder struct {
	typ int
	// general parts
	where                *whereClause
	table                string
	placeholderGenerator func(n int) []string

	// select parts
	orderBy  *orderByClause
	groupBy  *GroupBy
	selected *Selected
	subQuery *QueryBuilder
	joins    []*Join
	limit    *Limit
	offset   *Offset

	// update parts
	sets [][2]interface{}
}

func (d *QueryBuilder) toSqlDelete() (string, []interface{}, error) {
	base := fmt.Sprintf("DELETE FROM %s", d.table)
	var args []interface{}
	if d.where != nil {
		d.where.PlaceHolderGenerator = d.placeholderGenerator
		where, whereArgs := d.where.ToSql()
		base += " WHERE " + where
		args = append(args, whereArgs...)
	}
	return base, args, nil
}
func pop(phs *[]string) string {
	top := (*phs)[len(*phs)-1]
	*phs = (*phs)[:len(*phs)-1]
	return top
}

func (u *QueryBuilder) kvString() string {
	phs := u.placeholderGenerator(len(u.sets))
	var sets []string
	for _, pair := range u.sets {
		sets = append(sets, fmt.Sprintf("%s=%s", pair[0], pop(&phs)))
	}
	return strings.Join(sets, ",")
}

func (u *QueryBuilder) args() []interface{} {
	var values []interface{}
	for _, pair := range u.sets {
		values = append(values, pair[1])
	}
	return values
}

func (u *QueryBuilder) toSqlUpdate() (string, []interface{}, error) {
	if u.table == "" {
		return "", nil, fmt.Errorf("table cannot be empty")
	}
	base := fmt.Sprintf("UPDATE %s SET %s", u.table, u.kvString())
	args := u.args()
	if u.where != nil {
		u.where.PlaceHolderGenerator = u.placeholderGenerator
		where, whereArgs := u.where.ToSql()
		args = append(args, whereArgs...)
		base += " WHERE " + where
	}
	return base, args, nil
}
func (s *QueryBuilder) toSqlSelect() (string, []interface{}, error) {
	base := "SELECT"
	var args []interface{}
	//select
	if s.selected == nil {
		s.selected = &Selected{
			Columns: []string{"*"},
		}
	}
	base += " " + s.selected.String()
	// from
	if s.table == "" && s.subQuery == nil {
		panic("Table name cannot be empty")
	} else if s.table != "" && s.subQuery != nil {
		panic("cannot have both Table and subquery")
	}
	if s.table != "" {
		base += " " + "FROM " + s.table
	}
	if s.subQuery != nil {
		s.subQuery.placeholderGenerator = s.placeholderGenerator
		subQuery, subArgs, err := s.subQuery.ToSql()
		if err != nil {

			return "", nil, fmt.Errorf("SubQuery: %w", err)
		}
		base += " " + "FROM (" + subQuery + " )"
		args = append(args, subArgs...)
	}
	// Joins
	if s.joins != nil {
		for _, join := range s.joins {
			base += " " + join.String()
		}
	}
	// whereClause
	if s.where != nil {
		s.where.PlaceHolderGenerator = s.placeholderGenerator
		where, whereArgs := s.where.ToSql()
		base += " WHERE " + where
		args = append(args, whereArgs...)
	}

	// orderByClause
	if s.orderBy != nil {
		base += " " + s.orderBy.String()
	}

	// GroupBy
	if s.groupBy != nil {
		base += " " + s.groupBy.String()
	}

	// Limit
	if s.limit != nil {
		base += " " + s.limit.String()
	}

	// Offset
	if s.offset != nil {
		base += " " + s.offset.String()
	}

	return base, args, nil
}
func (q *QueryBuilder) ToSql() (string, []interface{}, error) {
	if q.typ == queryType_SELECT {
		return q.toSqlSelect()
	} else if q.typ == queryType_Delete {
		return q.toSqlDelete()
	} else if q.typ == queryType_UPDATE {
		return q.toSqlUpdate()
	} else {
		return "", nil, fmt.Errorf("no sql type matched")
	}
}

type orderByOrder string

const (
	OrderByASC  = "ASC"
	OrderByDesc = "DESC"
)

type orderByClause struct {
	Columns [][2]string
}

func (o orderByClause) String() string {
	var tuples []string
	for _, pair := range o.Columns {
		tuples = append(tuples, fmt.Sprintf("%s %s", pair[0], pair[1]))
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(tuples, ","))
}

type GroupBy struct {
	Columns []string
}

func (g GroupBy) String() string {
	return fmt.Sprintf("GROUP BY %s", strings.Join(g.Columns, ","))
}

type joinType string

const (
	JoinTypeInner = "INNER"
	JoinTypeLeft  = "LEFT"
	JoinTypeRight = "RIGHT"
	JoinTypeFull  = "FULL"
	JoinTypeSelf  = "SELF"
)

type JoinOn struct {
	Lhs string
	Rhs string
}

func (j JoinOn) String() string {
	return fmt.Sprintf("%s = %s", j.Lhs, j.Rhs)
}

type Join struct {
	Type  joinType
	Table string
	On    JoinOn
}

func (j Join) String() string {
	return fmt.Sprintf("%s JOIN %s ON %s", j.Type, j.Table, j.On.String())
}

type Limit struct {
	N int
}

func (l Limit) String() string {
	return fmt.Sprintf("LIMIT %d", l.N)
}

type Offset struct {
	N int
}

func (o Offset) String() string {
	return fmt.Sprintf("OFFSET %d", o.N)
}

type Having struct {
	PlaceHolderGenerator func(n int) []string
	Cond                 Cond
}

func (h Having) ToSql() (string, []interface{}) {
	h.Cond.PlaceHolderGenerator = h.PlaceHolderGenerator
	cond, condArgs := h.Cond.ToSql()
	return fmt.Sprintf("HAVING %s", cond), condArgs
}

type Selected struct {
	Columns []string
}

func (s Selected) String() string {
	return fmt.Sprintf("%s", strings.Join(s.Columns, ","))
}

func (q *QueryBuilder) WhereIn(column string, values ...interface{}) *QueryBuilder {
	if q.where == nil {
		q.where = &whereClause{
			Cond: Cond{
				Lhs: column,
				Op:  In,
				Rhs: values,
			},
		}
		return q
	} else {
		return q.addWhere("AND", append([]interface{}{column, In}, values...))
	}
}

func (q *QueryBuilder) OrderBy(column string, how string) *QueryBuilder {
	q.SetSelect()
	if q.orderBy == nil {
		q.orderBy = &orderByClause{}
	}
	q.orderBy.Columns = append(q.orderBy.Columns, [2]string{column, how})
	return q
}

func (q *QueryBuilder) LeftJoin(table string, onLhs string, onRhs string) *QueryBuilder {
	q.SetSelect()
	q.joins = append(q.joins, &Join{
		Type:  JoinTypeLeft,
		Table: table,
		On: JoinOn{
			Lhs: onLhs,
			Rhs: onRhs,
		},
	})
	return q
}
func (q *QueryBuilder) RightJoin(table string, onLhs string, onRhs string) *QueryBuilder {
	q.SetSelect()
	q.joins = append(q.joins, &Join{
		Type:  JoinTypeRight,
		Table: table,
		On: JoinOn{
			Lhs: onLhs,
			Rhs: onRhs,
		},
	})
	return q
}
func (q *QueryBuilder) InnerJoin(table string, onLhs string, onRhs string) *QueryBuilder {
	q.SetSelect()
	q.joins = append(q.joins, &Join{
		Type:  JoinTypeInner,
		Table: table,
		On: JoinOn{
			Lhs: onLhs,
			Rhs: onRhs,
		},
	})
	return q
}
func (q *QueryBuilder) FullOuterJoin(table string, onLhs string, onRhs string) *QueryBuilder {
	q.SetSelect()
	q.joins = append(q.joins, &Join{
		Type:  JoinTypeFull,
		Table: table,
		On: JoinOn{
			Lhs: onLhs,
			Rhs: onRhs,
		},
	})
	return q
}

func (q *QueryBuilder) Where(parts ...interface{}) *QueryBuilder {
	if len(parts) == 2 {
		// Equal mode
		q.where = &whereClause{Cond: Cond{Lhs: parts[0].(string), Op: Eq, Rhs: parts[1]}}
		return q
	} else if len(parts) == 3 {
		// operator mode
		q.where = &whereClause{Cond: Cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2]}}
		return q
	} else {
		panic("wrong number of arguments passed to Where")
	}
}

func (q *QueryBuilder) AndWhere(parts ...interface{}) *QueryBuilder {
	return q.addWhere("AND", parts...)
}

func (q *QueryBuilder) OrWhere(parts ...interface{}) *QueryBuilder {
	return q.addWhere("OR", parts...)
}
func (q *QueryBuilder) addWhere(typ string, parts ...interface{}) *QueryBuilder {
	w := q.where
	for {
		if w == nil {
			break
		} else if w.next == nil {
			w.next = &whereClause{}
			w.nextTyp = typ
			w = w.next
			break
		} else {
			w = w.next
		}
	}
	if w == nil {
		w = &whereClause{}
	}
	if len(parts) == 2 {
		// Equal mode
		w.Cond = Cond{Lhs: parts[0].(string), Op: Eq, Rhs: parts[1]}
		return q
	} else if len(parts) == 3 {
		// operator mode
		w.Cond = Cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2]}
		return q
	} else {
		panic("wrong number of arguments passed to Where")
	}
}

func (q *QueryBuilder) Offset(n int) *QueryBuilder {
	q.SetSelect()
	q.offset = &Offset{N: n}
	return q
}

func (q *QueryBuilder) Limit(n int) *QueryBuilder {
	q.SetSelect()
	q.limit = &Limit{N: n}
	return q
}

func (q *QueryBuilder) Table(t string) *QueryBuilder {
	q.table = t
	return q
}

func (q *QueryBuilder) SetSelect() *QueryBuilder {
	q.typ = queryType_SELECT
	return q
}

func (q *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	q.SetSelect()
	if q.groupBy == nil {
		q.groupBy = &GroupBy{}
	}
	q.groupBy.Columns = append(q.groupBy.Columns, columns...)
	return q
}
func (q *QueryBuilder) Select(columns ...string) *QueryBuilder {
	q.SetSelect()
	if q.selected == nil {
		q.selected = &Selected{}
	}
	q.selected.Columns = append(q.selected.Columns, columns...)
	return q
}

func (q *QueryBuilder) FromQuery(subQuery *QueryBuilder) *QueryBuilder {
	q.SetSelect()
	q.subQuery = subQuery
	q.subQuery.SetSelect()
	return q
}

func (q *QueryBuilder) SetUpdate() *QueryBuilder {
	q.typ = queryType_UPDATE
	return q
}

func (q *QueryBuilder) Set(name string, value interface{}) *QueryBuilder {
	q.SetUpdate()
	q.sets = append(q.sets, [2]interface{}{name, value})
	return q
}
func (q *QueryBuilder) Sets(tuples ...[2]interface{}) *QueryBuilder {
	q.SetUpdate()
	q.sets = append(q.sets, tuples...)
	return q
}
func (q *QueryBuilder) SetDialect(dialect *Dialect) *QueryBuilder {
	q.placeholderGenerator = dialect.PlaceHolderGenerator
	return q
}
func (q *QueryBuilder) SetDelete() *QueryBuilder {
	q.typ = queryType_Delete
	return q
}
func (q *QueryBuilder) Delete() *QueryBuilder {
	q.SetDelete()
	return q
}

func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{}
}
