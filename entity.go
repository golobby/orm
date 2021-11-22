package orm

type entity struct {
	repo *Repository
	obj  interface{}
}

func (r *Repository) Entity(obj interface{}) *entity {
	return &entity{r, obj}
}
func (e *entity) Save() error {
	return e.repo.Save(e.obj)
}
func (e *entity) Update() error {
	return e.repo.Update(e.obj)
}
func (e *entity) Delete() error {
	return e.repo.Delete(e.obj)
}
