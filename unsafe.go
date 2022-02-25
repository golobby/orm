package orm

import "unsafe"

func makeNewPtrFor(t string) interface{} {
	switch t {
	case "int":
		return new(int)
	case "int8":
		return new(int8)
	case "int16":
		return new(int16)
	case "int32":
		return new(int32)
	case "int64":
		return new(int64)
	case "uint8":
		return new(uint8)
	case "uint16":
		return new(uint16)
	case "uint32":
		return new(uint32)
	case "uint64":
		return new(uint64)
	case "float32":
		return new(float32)
	case "float64":
		return new(float64)
	case "string":
		return new(string)
	default:
		panic("no type matched")
	}
}

// &Entity -> *Entity

func getValueAtOffset(out interface{}, obj unsafe.Pointer, offset uintptr) {
	switch out.(type) {
	case *int:
		pb := (*int)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*int)) = *pb
	case *int8:
		pb := (*int8)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*int8)) = *pb
	case *int16:
		pb := (*int16)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*int16)) = *pb
	case *int32:
		pb := (*int32)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*int32)) = *pb
	case *int64:
		pb := (*int64)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*int64)) = *pb
	case *uint8:
		pb := (*uint8)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*uint8)) = *pb
	case *uint16:
		pb := (*uint16)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*uint16)) = *pb
	case *uint32:
		pb := (*uint32)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*uint32)) = *pb
	case *uint64:
		pb := (*uint64)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*uint64)) = *pb
	case *float32:
		pb := (*float32)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*float32)) = *pb
	case *float64:
		pb := (*float64)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*float64)) = *pb
	case *string:
		pb := (*string)(unsafe.Pointer(uintptr(obj) + offset))
		*(out.(*string)) = *pb
	default:
		panic("no type matched")
	}

}

func setValueAtOffset(obj unsafe.Pointer, offset uintptr, value interface{}) {
	switch value.(type) {
	case int:
		pb := (*int)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(int)
	case int8:
		pb := (*int8)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(int8)
	case int16:
		pb := (*int16)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(int16)
	case int32:
		pb := (*int32)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(int32)
	case int64:
		pb := (*int64)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(int64)
	case uint8:
		pb := (*uint8)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(uint8)
	case uint16:
		pb := (*uint16)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(uint16)
	case uint32:
		pb := (*uint32)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(uint32)
	case uint64:
		pb := (*uint64)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(uint64)
	case float32:
		pb := (*float32)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(float32)
	case float64:
		pb := (*float64)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(float64)
	case string:
		pb := (*string)(unsafe.Pointer(uintptr(obj) + offset))
		*pb = value.(string)
	default:
		panic("no type matched")
	}
}
