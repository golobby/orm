package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	queryTypeSELECT = iota + 1
	queryTypeUPDATE
	queryTypeDelete
)

// QueryBuilder is our query builder, almost all methods and functions in GoLobby ORM
// create or configure instance of QueryBuilder.
type QueryBuilder[OUTPUT any] struct {
	typ    int
	schema *schema
	// general parts
	where                *whereClause
	table                string
	placeholderGenerator func(n int) []string

	// select parts
	orderBy  *orderByClause
	groupBy  *GroupBy
	selected *selected
	subQuery *struct {
		q                    string
		args                 []interface{}
		placeholderGenerator func(n int) []string
	}
	joins  []*Join
	limit  *Limit
	offset *Offset

	// update parts
	sets [][2]interface{}

	// execution parts
	db  *sql.DB
	err error
}

// Finisher APIs

// execute is a finisher executes QueryBuilder query, remember to use this when you have an Update
// or Delete Query.
func (q *QueryBuilder[OUTPUT]) execute() (sql.Result, error) {
	if q.err != nil {
		return nil, q.err
	}
	if q.typ == queryTypeSELECT {
		return nil, fmt.Errorf("query type is SELECT")
	}
	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	return q.schema.getConnection().exec(query, args...)
}

// Get limit results to 1, runs query generated by query builder, scans result into OUTPUT.
func (q *QueryBuilder[OUTPUT]) Get() (OUTPUT, error) {
	if q.err != nil {
		return *new(OUTPUT), q.err
	}
	queryString, args, err := q.ToSql()
	if err != nil {
		return *new(OUTPUT), err
	}
	rows, err := q.schema.getConnection().query(queryString, args...)
	if err != nil {
		return *new(OUTPUT), err
	}
	var output OUTPUT
	err = newBinder(q.schema).bind(rows, &output)
	if err != nil {
		return *new(OUTPUT), err
	}
	return output, nil
}

// All is a finisher, create the Select query based on QueryBuilder and scan results into
// slice of type parameter E.
func (q *QueryBuilder[OUTPUT]) All() ([]OUTPUT, error) {
	if q.err != nil {
		return nil, q.err
	}
	q.SetSelect()
	queryString, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := q.schema.getConnection().query(queryString, args...)
	if err != nil {
		return nil, err
	}
	var output []OUTPUT
	err = newBinder(q.schema).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

// Delete is a finisher, creates a delete query from query builder and executes it.
func (q *QueryBuilder[OUTPUT]) Delete() (rowsAffected int64, err error) {
	if q.err != nil {
		return 0, q.err
	}
	q.SetDelete()
	res, err := q.execute()
	if err != nil {
		return 0, q.err
	}
	return res.RowsAffected()
}

// Update is a finisher, creates an Update query from QueryBuilder and executes in into database, returns
func (q *QueryBuilder[OUTPUT]) Update() (rowsAffected int64, err error) {
	if q.err != nil {
		return 0, q.err
	}
	q.SetUpdate()
	res, err := q.execute()
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func copyQueryBuilder[T1 any, T2 any](q *QueryBuilder[T1], q2 *QueryBuilder[T2]) {
	q2.db = q.db
	q2.err = q.err
	q2.groupBy = q.groupBy
	q2.joins = q.joins
	q2.limit = q.limit
	q2.offset = q.offset
	q2.orderBy = q.orderBy
	q2.placeholderGenerator = q.placeholderGenerator
	q2.schema = q.schema
	q2.selected = q.selected
	q2.sets = q.sets

	q2.subQuery = q.subQuery
	q2.table = q.table
	q2.typ = q.typ
	q2.where = q.where
}

// Count creates and execute a select query from QueryBuilder and set it's field list of selection
// to COUNT(id).
func (q *QueryBuilder[OUTPUT]) Count() *QueryBuilder[int] {
	q.selected = &selected{Columns: []string{"COUNT(id)"}}
	q.SetSelect()
	qCount := NewQueryBuilder[int](q.schema)

	copyQueryBuilder(q, qCount)

	return qCount
}

// First returns first record of database using OrderBy primary key
// ascending order.
func (q *QueryBuilder[OUTPUT]) First() *QueryBuilder[OUTPUT] {
	q.OrderBy(q.schema.pkName(), ASC).Limit(1)
	return q
}

// Latest is like Get but it also do a OrderBy(primary key, DESC)
func (q *QueryBuilder[OUTPUT]) Latest() *QueryBuilder[OUTPUT] {
	q.OrderBy(q.schema.pkName(), DESC).Limit(1)
	return q
}

// WherePK adds a where clause to QueryBuilder and also gets primary key name
// from type parameter schema.
func (q *QueryBuilder[OUTPUT]) WherePK(value interface{}) *QueryBuilder[OUTPUT] {
	return q.Where(q.schema.pkName(), value)
}

func (d *QueryBuilder[OUTPUT]) toSqlDelete() (string, []interface{}, error) {
	base := fmt.Sprintf("DELETE FROM %s", d.table)
	var args []interface{}
	if d.where != nil {
		d.where.PlaceHolderGenerator = d.placeholderGenerator
		where, whereArgs, err := d.where.ToSql()
		if err != nil {
			return "", nil, err
		}
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

func (u *QueryBuilder[OUTPUT]) kvString() string {
	phs := u.placeholderGenerator(len(u.sets))
	var sets []string
	for _, pair := range u.sets {
		sets = append(sets, fmt.Sprintf("%s=%s", pair[0], pop(&phs)))
	}
	return strings.Join(sets, ",")
}

func (u *QueryBuilder[OUTPUT]) args() []interface{} {
	var values []interface{}
	for _, pair := range u.sets {
		values = append(values, pair[1])
	}
	return values
}

func (u *QueryBuilder[OUTPUT]) toSqlUpdate() (string, []interface{}, error) {
	if u.table == "" {
		return "", nil, fmt.Errorf("table cannot be empty")
	}
	base := fmt.Sprintf("UPDATE %s SET %s", u.table, u.kvString())
	args := u.args()
	if u.where != nil {
		u.where.PlaceHolderGenerator = u.placeholderGenerator
		where, whereArgs, err := u.where.ToSql()
		if err != nil {
			return "", nil, err
		}
		args = append(args, whereArgs...)
		base += " WHERE " + where
	}
	return base, args, nil
}
func (s *QueryBuilder[OUTPUT]) toSqlSelect() (string, []interface{}, error) {
	if s.err != nil {
		return "", nil, s.err
	}
	base := "SELECT"
	var args []interface{}
	// select
	if s.selected == nil {
		s.selected = &selected{
			Columns: []string{"*"},
		}
	}
	base += " " + s.selected.String()
	// from
	if s.table == "" && s.subQuery == nil {
		return "", nil, fmt.Errorf("Table name cannot be empty")
	} else if s.table != "" && s.subQuery != nil {
		return "", nil, fmt.Errorf("cannot have both Table and subquery")
	}
	if s.table != "" {
		base += " " + "FROM " + s.table
	}
	if s.subQuery != nil {
		s.subQuery.placeholderGenerator = s.placeholderGenerator
		base += " " + "FROM (" + s.subQuery.q + " )"
		args = append(args, s.subQuery.args...)
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
		where, whereArgs, err := s.where.ToSql()
		if err != nil {
			return "", nil, err
		}
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

// ToSql creates sql query from QueryBuilder based on internal fields it would decide what kind
// of query to build.
func (q *QueryBuilder[OUTPUT]) ToSql() (string, []interface{}, error) {
	if q.err != nil {
		return "", nil, q.err
	}
	if q.typ == queryTypeSELECT {
		return q.toSqlSelect()
	} else if q.typ == queryTypeDelete {
		return q.toSqlDelete()
	} else if q.typ == queryTypeUPDATE {
		return q.toSqlUpdate()
	} else {
		return "", nil, fmt.Errorf("no sql type matched")
	}
}

type orderByOrder string

const (
	ASC  = "ASC"
	DESC = "DESC"
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

// OrderBy adds an OrderBy section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) OrderBy(column string, how string) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	if q.orderBy == nil {
		q.orderBy = &orderByClause{}
	}
	q.orderBy.Columns = append(q.orderBy.Columns, [2]string{column, how})
	return q
}

// LeftJoin adds a left join section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) LeftJoin(table string, onLhs string, onRhs string) *QueryBuilder[OUTPUT] {
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

// RightJoin adds a right join section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) RightJoin(table string, onLhs string, onRhs string) *QueryBuilder[OUTPUT] {
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

// InnerJoin adds a inner join section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) InnerJoin(table string, onLhs string, onRhs string) *QueryBuilder[OUTPUT] {
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

// Join adds a inner join section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) Join(table string, onLhs string, onRhs string) *QueryBuilder[OUTPUT] {
	return q.InnerJoin(table, onLhs, onRhs)
}

// FullOuterJoin adds a full outer join section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) FullOuterJoin(table string, onLhs string, onRhs string) *QueryBuilder[OUTPUT] {
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

// Where Adds a where clause to query, if already have where clause append to it
// as AndWhere.
func (q *QueryBuilder[OUTPUT]) Where(parts ...interface{}) *QueryBuilder[OUTPUT] {
	if q.where != nil {
		return q.addWhere("AND", parts...)
	}
	if len(parts) == 1 {
		if r, isRaw := parts[0].(*raw); isRaw {
			q.where = &whereClause{raw: r.sql, args: r.args, PlaceHolderGenerator: q.placeholderGenerator}
			return q
		} else {
			q.err = fmt.Errorf("when you have one argument passed to where, it should be *raw")
			return q
		}

	} else if len(parts) == 2 {
		if strings.Index(parts[0].(string), " ") == -1 {
			// Equal mode
			q.where = &whereClause{cond: cond{Lhs: parts[0].(string), Op: Eq, Rhs: parts[1]}, PlaceHolderGenerator: q.placeholderGenerator}
		}
		return q
	} else if len(parts) == 3 {
		// operator mode
		q.where = &whereClause{cond: cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2]}, PlaceHolderGenerator: q.placeholderGenerator}
		return q
	} else if len(parts) > 3 && parts[1].(string) == "IN" {
		q.where = &whereClause{cond: cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2:]}, PlaceHolderGenerator: q.placeholderGenerator}
		return q
	} else {
		q.err = fmt.Errorf("wrong number of arguments passed to Where")
		return q
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

type cond struct {
	PlaceHolderGenerator func(n int) []string

	Lhs string
	Op  binaryOp
	Rhs interface{}
}

func (b cond) ToSql() (string, []interface{}, error) {
	var phs []string
	if b.Op == In {
		rhs, isInterfaceSlice := b.Rhs.([]interface{})
		if isInterfaceSlice {
			phs = b.PlaceHolderGenerator(len(rhs))
			return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), rhs, nil
		} else if rawThing, isRaw := b.Rhs.(*raw); isRaw {
			return fmt.Sprintf("%s IN (%s)", b.Lhs, rawThing.sql), rawThing.args, nil
		} else {
			return "", nil, fmt.Errorf("Right hand side of Cond when operator is IN should be either a interface{} slice or *raw")
		}

	} else {
		phs = b.PlaceHolderGenerator(1)
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, pop(&phs)), []interface{}{b.Rhs}, nil
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
	cond
	raw  string
	args []interface{}
}

func (w whereClause) ToSql() (string, []interface{}, error) {
	var base string
	var args []interface{}
	var err error
	if w.raw != "" {
		base = w.raw
		args = w.args
	} else {
		w.cond.PlaceHolderGenerator = w.PlaceHolderGenerator
		base, args, err = w.cond.ToSql()
		if err != nil {
			return "", nil, err
		}
	}
	if w.next == nil {
		return base, args, nil
	}
	if w.next != nil {
		next, nextArgs, err := w.next.ToSql()
		if err != nil {
			return "", nil, err
		}
		base += " " + w.nextTyp + " " + next
		args = append(args, nextArgs...)
		return base, args, nil
	}

	return base, args, nil
}

//func (q *QueryBuilder[OUTPUT]) WhereKeyValue(m map) {}

// WhereIn adds a where clause to QueryBuilder using In operator.
func (q *QueryBuilder[OUTPUT]) WhereIn(column string, values ...interface{}) *QueryBuilder[OUTPUT] {
	return q.Where(append([]interface{}{column, In}, values...)...)
}

// AndWhere appends a where clause to query builder as And where clause.
func (q *QueryBuilder[OUTPUT]) AndWhere(parts ...interface{}) *QueryBuilder[OUTPUT] {
	return q.addWhere(nextType_AND, parts...)
}

// OrWhere appends a where clause to query builder as Or where clause.
func (q *QueryBuilder[OUTPUT]) OrWhere(parts ...interface{}) *QueryBuilder[OUTPUT] {
	return q.addWhere(nextType_OR, parts...)
}

func (q *QueryBuilder[OUTPUT]) addWhere(typ string, parts ...interface{}) *QueryBuilder[OUTPUT] {
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
		w.cond = cond{Lhs: parts[0].(string), Op: Eq, Rhs: parts[1]}
		return q
	} else if len(parts) == 3 {
		// operator mode
		w.cond = cond{Lhs: parts[0].(string), Op: binaryOp(parts[1].(string)), Rhs: parts[2]}
		return q
	} else {
		panic("wrong number of arguments passed to Where")
	}
}

// Offset adds offset section to query builder.
func (q *QueryBuilder[OUTPUT]) Offset(n int) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	q.offset = &Offset{N: n}
	return q
}

// Limit adds limit section to query builder.
func (q *QueryBuilder[OUTPUT]) Limit(n int) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	q.limit = &Limit{N: n}
	return q
}

// Table sets table of QueryBuilder.
func (q *QueryBuilder[OUTPUT]) Table(t string) *QueryBuilder[OUTPUT] {
	q.table = t
	return q
}

// SetSelect sets query type of QueryBuilder to Select.
func (q *QueryBuilder[OUTPUT]) SetSelect() *QueryBuilder[OUTPUT] {
	q.typ = queryTypeSELECT
	return q
}

// GroupBy adds a group by section to QueryBuilder.
func (q *QueryBuilder[OUTPUT]) GroupBy(columns ...string) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	if q.groupBy == nil {
		q.groupBy = &GroupBy{}
	}
	q.groupBy.Columns = append(q.groupBy.Columns, columns...)
	return q
}

// Select adds columns to QueryBuilder select field list.
func (q *QueryBuilder[OUTPUT]) Select(columns ...string) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	if q.selected == nil {
		q.selected = &selected{}
	}
	q.selected.Columns = append(q.selected.Columns, columns...)
	return q
}

// FromQuery sets subquery of QueryBuilder to be given subquery so
// when doing select instead of from table we do from(subquery).
func (q *QueryBuilder[OUTPUT]) FromQuery(subQuery *QueryBuilder[OUTPUT]) *QueryBuilder[OUTPUT] {
	q.SetSelect()
	subQuery.SetSelect()
	subQuery.placeholderGenerator = q.placeholderGenerator
	subQueryString, args, err := subQuery.ToSql()
	q.err = err
	q.subQuery = &struct {
		q                    string
		args                 []interface{}
		placeholderGenerator func(n int) []string
	}{
		subQueryString, args, q.placeholderGenerator,
	}
	return q
}

func (q *QueryBuilder[OUTPUT]) SetUpdate() *QueryBuilder[OUTPUT] {
	q.typ = queryTypeUPDATE
	return q
}

func (q *QueryBuilder[OUTPUT]) Set(keyValues ...any) *QueryBuilder[OUTPUT] {
	if len(keyValues)%2 != 0 {
		q.err = fmt.Errorf("when using Set, passed argument count should be even: %w", q.err)
		return q
	}
	q.SetUpdate()
	for i := 0; i < len(keyValues); i++ {
		if i != 0 && i%2 == 1 {
			q.sets = append(q.sets, [2]any{keyValues[i-1], keyValues[i]})
		}
	}
	return q
}

func (q *QueryBuilder[OUTPUT]) SetDialect(dialect *Dialect) *QueryBuilder[OUTPUT] {
	q.placeholderGenerator = dialect.PlaceHolderGenerator
	return q
}
func (q *QueryBuilder[OUTPUT]) SetDelete() *QueryBuilder[OUTPUT] {
	q.typ = queryTypeDelete
	return q
}

type raw struct {
	sql  string
	args []interface{}
}

// Raw creates a Raw sql query chunk that you can add to several components of QueryBuilder like
// Wheres.
func Raw(sql string, args ...interface{}) *raw {
	return &raw{sql: sql, args: args}
}

func NewQueryBuilder[OUTPUT any](s *schema) *QueryBuilder[OUTPUT] {
	return &QueryBuilder[OUTPUT]{schema: s}
}

type insertStmt struct {
	PlaceHolderGenerator func(n int) []string
	Table                string
	Columns              []string
	Values               [][]interface{}
	Returning            string
}

func (i insertStmt) flatValues() []interface{} {
	var values []interface{}
	for _, row := range i.Values {
		values = append(values, row...)
	}
	return values
}

func (i insertStmt) getValuesStr() string {
	phs := i.PlaceHolderGenerator(len(i.Values) * len(i.Values[0]))

	var output []string
	for _, valueRow := range i.Values {
		output = append(output, fmt.Sprintf("(%s)", strings.Join(phs[:len(valueRow)], ",")))
		phs = phs[len(valueRow):]
	}
	return strings.Join(output, ",")
}

func (i insertStmt) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		i.Table,
		strings.Join(i.Columns, ","),
		i.getValuesStr(),
	)
	if i.Returning != "" {
		base += "RETURNING " + i.Returning
	}
	return base, i.flatValues()
}

func postgresPlaceholder(n int) []string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return output
}

func questionMarks(n int) []string {
	output := []string{}
	for i := 0; i < n; i++ {
		output = append(output, "?")
	}

	return output
}
