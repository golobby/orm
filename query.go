package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	queryType_SELECT = iota + 1
	queryType_UPDATE
	queryType_Delete
)

//QueryBuilder is our query builder, almost all methods and functions in GolobbyORM
//create or configure instance of QueryBuilder.
type QueryBuilder[E Entity] struct {
	typ int
	// general parts
	where                *whereClause
	table                string
	placeholderGenerator func(n int) []string

	// select parts
	orderBy  *orderByClause
	groupBy  *GroupBy
	selected *selected
	subQuery *QueryBuilder[E]
	joins    []*Join
	limit    *Limit
	offset   *Offset

	// update parts
	sets [][2]interface{}

	//execution parts
	db *sql.DB
}

type raw struct {
	sql  string
	args []interface{}
}

//Raw creates a Raw sql query chunk that you can add to several components of QueryBuilder like
//Wheres.
func Raw(sql string, args ...interface{}) *raw {
	return &raw{sql: sql, args: args}
}

//All create the Select query based on QueryBuilder and scan results into
//slice of type parameter E.
func (q *QueryBuilder[E]) All() ([]E, error) {
	q.SetSelect()
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := getSchemaFor(*new(E)).getSQLDB().Query(query, args...)
	if err != nil {
		return nil, err
	}
	var output []E
	err = getSchemaFor(*new(E)).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

//One create the Select query based on QueryBuilder and scan results into
//object of type parameter E.
func (q *QueryBuilder[E]) One() (E, error) {
	q.Limit(1)
	query, args, err := q.ToSql()
	if err != nil {
		return *new(E), err
	}
	rows, err := getSchemaFor(*new(E)).getSQLDB().Query(query, args...)
	if err != nil {
		return *new(E), err
	}
	var output E
	err = getSchemaFor(*new(E)).bind(rows, &output)
	if err != nil {
		return *new(E), err
	}
	return output, nil
}

//Count creates and execute a select query from QueryBuilder and set it's column list of selection
//to COUNT(id).
func (q *QueryBuilder[E]) Count() (int64, error) {
	q.selected = &selected{Columns: []string{"COUNT(id)"}}
	q.SetSelect()
	query, args, err := q.ToSql()
	if err != nil {
		return 0, err
	}
	row := getSchemaFor(*new(E)).getSQLDB().QueryRow(query, args...)
	if row.Err() != nil {
		return 0, err
	}
	var counter int64
	err = row.Scan(&counter)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

//First is like One but it also do a OrderBy("id", OrderByASC)
func (q *QueryBuilder[E]) First() (E, error) {
	q.OrderBy("id", OrderByASC)
	return q.One()
}

//Latest is like One but it also do a OrderBy("id", OrderByDesc)
func (q *QueryBuilder[E]) Latest() (E, error) {
	q.OrderBy("id", OrderByDesc)
	return q.One()
}

//WherePK adds a where clause to QueryBuilder and also gets primary key name
//from type parameter schema.
func (q *QueryBuilder[E]) WherePK(value interface{}) *QueryBuilder[E] {
	return q.Where(getSchemaFor(*new(E)).pkName(), value)
}

//Execute executes QueryBuilder query, remember to use this when you have an Update
//or Delete Query.
func (q *QueryBuilder[E]) Execute() (sql.Result, error) {
	if q.typ == queryType_SELECT {
		return nil, fmt.Errorf("query type is SELECT")
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	return getSchemaFor(*new(E)).getSQLDB().Exec(query, args...)
}

//Delete sets QueryBuilder type to be delete and then Executes it.
func (q *QueryBuilder[E]) Delete() (sql.Result, error) {
	q.SetDelete()
	return q.Execute()
}

func asTuples(toUpdate map[string]interface{}) [][2]interface{} {
	var tuples [][2]interface{}
	for k, v := range toUpdate {
		tuples = append(tuples, [2]interface{}{k, v})
	}
	return tuples
}

type KV = map[string]interface{}

//Update creates an Update query from QueryBuilder and executes in into database, also adds all given key values in
//argument to list of key values of update query.
func (q *QueryBuilder[E]) Update(toUpdate map[string]interface{}) (sql.Result, error) {
	q.SetUpdate()
	q.sets = append(q.sets, asTuples(toUpdate)...)
	return q.Execute()
}

func (d *QueryBuilder[E]) toSqlDelete() (string, []interface{}, error) {
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

func (u *QueryBuilder[E]) kvString() string {
	phs := u.placeholderGenerator(len(u.sets))
	var sets []string
	for _, pair := range u.sets {
		sets = append(sets, fmt.Sprintf("%s=%s", pair[0], pop(&phs)))
	}
	return strings.Join(sets, ",")
}

func (u *QueryBuilder[E]) args() []interface{} {
	var values []interface{}
	for _, pair := range u.sets {
		values = append(values, pair[1])
	}
	return values
}

func (u *QueryBuilder[E]) toSqlUpdate() (string, []interface{}, error) {
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
func (s *QueryBuilder[E]) toSqlSelect() (string, []interface{}, error) {
	base := "SELECT"
	var args []interface{}
	//select
	if s.selected == nil {
		s.selected = &selected{
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

//ToSql creates sql query from QueryBuilder based on internal fields it would decide what kind
//of query to build.
func (q *QueryBuilder[E]) ToSql() (string, []interface{}, error) {
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
	JoinTypeFull  = "FULL OUTER"
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

type selected struct {
	Columns []string
}

func (s selected) String() string {
	return fmt.Sprintf("%s", strings.Join(s.Columns, ","))
}

//OrderBy adds an OrderBy section to QueryBuilder.
func (q *QueryBuilder[E]) OrderBy(column string, how string) *QueryBuilder[E] {
	q.SetSelect()
	if q.orderBy == nil {
		q.orderBy = &orderByClause{}
	}
	q.orderBy.Columns = append(q.orderBy.Columns, [2]string{column, how})
	return q
}

//LeftJoin adds a left join section to QueryBuilder.
func (q *QueryBuilder[E]) LeftJoin(table string, onLhs string, onRhs string) *QueryBuilder[E] {
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

//RightJoin adds a right join section to QueryBuilder.
func (q *QueryBuilder[E]) RightJoin(table string, onLhs string, onRhs string) *QueryBuilder[E] {
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

//InnerJoin adds a inner join section to QueryBuilder.
func (q *QueryBuilder[E]) InnerJoin(table string, onLhs string, onRhs string) *QueryBuilder[E] {
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

//Join adds a inner join section to QueryBuilder.
func (q *QueryBuilder[E]) Join(table string, onLhs string, onRhs string) *QueryBuilder[E] {
	return q.InnerJoin(table, onLhs, onRhs)
}

//FullOuterJoin adds a full outer join section to QueryBuilder.
func (q *QueryBuilder[E]) FullOuterJoin(table string, onLhs string, onRhs string) *QueryBuilder[E] {
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

//Where Adds a where clause to query, if already have where clause append to it
//as AndWhere.
func (q *QueryBuilder[E]) Where(parts ...interface{}) *QueryBuilder[E] {
	if q.where != nil {
		return q.addWhere("AND", parts...)
	}
	if len(parts) == 1 {
		if r, isRaw := parts[0].(*raw); isRaw {
			q.where = &whereClause{raw: r.sql, args: r.args, PlaceHolderGenerator: q.placeholderGenerator}
			return q
		} else {
			panic("when you have one argument passed to where, it should be *raw")
		}

	} else if len(parts) == 2 {
		if strings.Index(parts[0].(string), " ") == -1 {
			// Equal mode
			q.where = &whereClause{Cond: Cond{Lhs: parts[0].(string), Op: Eq, Rhs: parts[1]}, PlaceHolderGenerator: q.placeholderGenerator}
		}
		return q
	} else if len(parts) == 3 {
		// operator mode
		q.where = &whereClause{Cond: Cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2]}, PlaceHolderGenerator: q.placeholderGenerator}
		return q
	} else {
		panic("wrong number of arguments passed to Where")
	}
}

type binaryOp string

const (
	Eq      = "="
	GT      = ">"
	LT      = "<"
	GE      = ">="
	LE      = "<="
	NE      = "!="
	Between = "BETWEEN"
	Like    = "LIKE"
	In      = "IN"
)

type Cond struct {
	PlaceHolderGenerator func(n int) []string

	Lhs string
	Op  binaryOp
	Rhs interface{}
}

func (b Cond) ToSql() (string, []interface{}) {
	var phs []string
	if b.Op == In {
		rhs, isInterfaceSlice := b.Rhs.([]interface{})
		if isInterfaceSlice {
			phs = b.PlaceHolderGenerator(len(rhs))
			return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), rhs
		} else if rawThing, isRaw := b.Rhs.(*raw); isRaw {
			return fmt.Sprintf("%s IN (%s)", b.Lhs, rawThing.sql), rawThing.args
		} else {
			panic("Right hand side of Cond when operator is IN should be either a interface{} slice or *raw")
		}

	} else {
		phs = b.PlaceHolderGenerator(1)
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, pop(&phs)), []interface{}{b.Rhs}
	}
}

const (
	nextType_AND = "AND"
	nextType_OR  = "OR"
)

type whereClause struct {
	PlaceHolderGenerator func(n int) []string
	nextTyp              string
	next                 *whereClause
	Cond
	raw  string
	args []interface{}
}

func (w whereClause) ToSql() (string, []interface{}) {
	var base string
	var args []interface{}
	if w.raw != "" {
		base = w.raw
		args = w.args
	} else {
		w.Cond.PlaceHolderGenerator = w.PlaceHolderGenerator
		base, args = w.Cond.ToSql()
	}
	if w.next == nil {
		return base, args
	}
	if w.next != nil {
		next, nextArgs := w.next.ToSql()
		base += " " + w.nextTyp + " " + next
		args = append(args, nextArgs...)
		return base, args
	}

	return base, args
}

//WhereIn adds a where clause to QueryBuilder using In operator.
func (q *QueryBuilder[E]) WhereIn(column string, values ...interface{}) *QueryBuilder[E] {
	return q.Where(append([]interface{}{column, In}, values...)...)
}

//AndWhere appends a where clause to query builder as And where clause.
func (q *QueryBuilder[E]) AndWhere(parts ...interface{}) *QueryBuilder[E] {
	return q.addWhere(nextType_AND, parts...)
}

//OrWhere appends a where clause to query builder as Or where clause.
func (q *QueryBuilder[E]) OrWhere(parts ...interface{}) *QueryBuilder[E] {
	return q.addWhere(nextType_OR, parts...)
}

func (q *QueryBuilder[E]) addWhere(typ string, parts ...interface{}) *QueryBuilder[E] {
	w := q.where
	for {
		if w == nil {
			break
		} else if w.next == nil {
			w.next = &whereClause{PlaceHolderGenerator: q.placeholderGenerator}
			w.nextTyp = typ
			w = w.next
			break
		} else {
			w = w.next
		}
	}
	if w == nil {
		w = &whereClause{PlaceHolderGenerator: q.placeholderGenerator}
	}
	if len(parts) == 1 {
		w.raw = parts[0].(*raw).sql
		w.args = parts[0].(*raw).args
		return q
	} else if len(parts) == 2 {
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

//Offset adds offset section to query builder.
func (q *QueryBuilder[E]) Offset(n int) *QueryBuilder[E] {
	q.SetSelect()
	q.offset = &Offset{N: n}
	return q
}

//Limit adds limit section to query builder.
func (q *QueryBuilder[E]) Limit(n int) *QueryBuilder[E] {
	q.SetSelect()
	q.limit = &Limit{N: n}
	return q
}

//Table sets table of QueryBuilder.
func (q *QueryBuilder[E]) Table(t string) *QueryBuilder[E] {
	q.table = t
	return q
}

//SetSelect sets query type of QueryBuilder to Select.
func (q *QueryBuilder[E]) SetSelect() *QueryBuilder[E] {
	q.typ = queryType_SELECT
	return q
}

//GroupBy adds a group by section to QueryBuilder.
func (q *QueryBuilder[E]) GroupBy(columns ...string) *QueryBuilder[E] {
	q.SetSelect()
	if q.groupBy == nil {
		q.groupBy = &GroupBy{}
	}
	q.groupBy.Columns = append(q.groupBy.Columns, columns...)
	return q
}

//Select adds columns to QueryBuilder select column list.
func (q *QueryBuilder[E]) Select(columns ...string) *QueryBuilder[E] {
	q.SetSelect()
	if q.selected == nil {
		q.selected = &selected{}
	}
	q.selected.Columns = append(q.selected.Columns, columns...)
	return q
}

//FromQuery sets subquery of QueryBuilder to be given subquery so
//when doing select instead of from table we do from(subquery).
func (q *QueryBuilder[E]) FromQuery(subQuery *QueryBuilder[E]) *QueryBuilder[E] {
	q.SetSelect()
	q.subQuery = subQuery
	q.subQuery.SetSelect()
	return q
}

func (q *QueryBuilder[E]) SetUpdate() *QueryBuilder[E] {
	q.typ = queryType_UPDATE
	return q
}

func (q *QueryBuilder[E]) Set(name string, value interface{}) *QueryBuilder[E] {
	q.SetUpdate()
	q.sets = append(q.sets, [2]interface{}{name, value})
	return q
}

func (q *QueryBuilder[E]) Sets(tuples ...[2]interface{}) *QueryBuilder[E] {
	q.SetUpdate()
	q.sets = append(q.sets, tuples...)
	return q
}
func (q *QueryBuilder[E]) SetDialect(dialect *Dialect) *QueryBuilder[E] {
	q.placeholderGenerator = dialect.PlaceHolderGenerator
	return q
}
func (q *QueryBuilder[E]) SetDelete() *QueryBuilder[E] {
	q.typ = queryType_Delete
	return q
}

func NewQueryBuilder[E Entity]() *QueryBuilder[E] {
	return &QueryBuilder[E]{}
}
