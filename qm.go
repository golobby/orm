package orm

//
//type SelectModifier interface {
//	SModify(s Select) Select
//}
//type WhereModifier interface {
//}
//
//type where struct {
//	cond Cond
//}
//
//func (w where) SModify(s Select) Select {
//	s.Where = &whereClause{Cond: w.cond}
//	return s
//}
//
//func WhereIn(column string, values ...interface{}) SelectModifier {
//	return where{
//		cond: Cond{
//			Lhs: column,
//			Op:  In,
//			Rhs: values,
//		},
//	}
//}
//func Where(parts ...interface{}) SelectModifier {
//
//}
//
//type order struct {
//	column  string
//	orderby orderByOrder
//}
//
//func (o order) SModify(s Select) Select {
//	if s.OrderBy == nil {
//		s.OrderBy = &orderByClause{
//			Columns: [][2]string{{o.column, string(o.orderby)}},
//		}
//	} else {
//		s.OrderBy.Columns = append(s.OrderBy.Columns, [2]string{o.column, string(o.orderby)})
//	}
//	return s
//}
//
//func OrderBy(column string, how orderByOrder) SelectModifier {
//	return order{column: column, orderby: how}
//}
