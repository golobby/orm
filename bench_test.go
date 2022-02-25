package orm

import (
	"testing"
	"unsafe"
)

type A struct {
	ID int
}

func (a A) Schema() *Schema {
	return &Schema{Table: "as"}
}

func setup() {
	Initialize(ConnectionConfig{
		Name:             "default",
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
		Entities:         []Entity{A{}},
	})

	GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text)`)
	GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, body text)`)
	GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
	GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)
}

func BenchmarkGenericGet10(b *testing.B) {
	setup()
	var a A
	for i := 0; i < b.N; i++ {
		genericGetPKValue(a)
	}
}

func BenchmarkUnsafeGetValueAtOffset(b *testing.B) {
	setup()
	a := &A{ID: 1}
	offset := unsafe.Offsetof(a.ID)
	var id int
	for i := 0; i < b.N; i++ {
		getValueAtOffset(&id, unsafe.Pointer(a), offset)
	}
}
