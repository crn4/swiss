package hash

import "unsafe"

type HFunc func(unsafe.Pointer, uintptr) uintptr

// general map struct from go/src/runtime/runtime2.go
type eface struct {
	_type *mapType
	data  unsafe.Pointer
}

// runtime map type from go/src/internal/abi/type.go
type mapType struct {
	rtype
	Key    *rtype
	Elem   *rtype
	Bucket *rtype // internal type representing a hash bucket
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	KeySize    uint8  // size of key slot
	ValueSize  uint8  // size of elem slot
	BucketSize uint16 // size of bucket
	Flags      uint32
}

type rtype struct {
	Size_       uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       tFlag   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       uint8   // enumeration for C
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// If the KindGCProg bit is set in kind, GCData is a GC program.
	// Otherwise it is a ptrmask bitmap. See mbitmap.go for details.
	GCData    *byte
	Str       nameOff // string form
	PtrToThis typeOff // type for pointer to this type, may be zero
}

type tFlag uint8
type nameOff int32
type typeOff int32

func GetHashFuncRnt[K comparable]() HFunc {
	m := any((map[K]struct{})(nil))
	return (*eface)(unsafe.Pointer(&m))._type.Hasher
}

//go:linkname runtime_memhash runtime.memhash
func runtime_memhash(p unsafe.Pointer, seed, s uintptr) uintptr

func GetHashFuncMemhash[K comparable]() HFunc {
	var key K
	sz := unsafe.Sizeof(key)
	return func(p unsafe.Pointer, u uintptr) uintptr {
		return runtime_memhash(p, u, sz)
	}
}

func GetHashFunc[K comparable]() HFunc {
	var k K
	switch any(k).(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16,
		uint32, uint64, uintptr, float32, float64:
		return GetHashFuncRnt[K]()
	default:
		return GetHashFuncMemhash[K]()
	}
}
