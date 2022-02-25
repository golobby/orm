package orm

//func TestSetValueAtOffset(t *testing.T) {
//	a := &A{}
//	var b Entity
//	b = a
//	offset := unsafe.Offsetof(a.ID)
//	setValueAtOffset(unsafe.Pointer(&b), offset, 2)
//
//	assert.Equal(t, 2, a.ID)
//}
//
//func TestGetValueAtOffset(t *testing.T) {
//	a := &A{ID: 1}
//	var b Entity
//	b = a
//	offset := unsafe.Offsetof(a.ID)
//	var id int
//	getValueAtOffset(&id, unsafe.Pointer(&b), offset)
//
//	assert.Equal(t, 1, id)
//}
