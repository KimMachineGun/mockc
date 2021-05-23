package basic

import (
	"unsafe"
)

type anon struct {
	C int
	D int8
}

type TypeCode interface {
	Bool(...bool) bool
	Int(...int) int
	Int8(...int8) int8
	Int16(...int16) int16
	Int32(...int32) int32
	Int64(...int64) int64
	Uint(...uint) uint
	Uint8(...uint8) uint8
	Uint16(...uint16) uint16
	Uint32(...uint32) uint32
	Uint64(...uint64) uint64
	Uintptr(...uintptr) uintptr
	Float32(...float32) float32
	Float64(...float64) float64
	Complex64(...complex64) complex64
	Complex128(...complex128) complex128
	String(...string) string
	Pointer(...unsafe.Pointer) unsafe.Pointer
	Byte(...byte) byte
	Rune(...rune) rune
	Array(...[0]bool) [0]bool
	Slice(...[]bool) []bool
	Struct(...struct{
		A bool
		anon
	}) struct {
		B int
	}
	BoolP(...*bool)	*bool
	Tuple()	(bool, int, int8)
	Func(func(bool, int, ...int8) (int32, int64)) func() error
	Interface(...interface{}) interface{
		Hello() string
		World() string
	}
	Map(...map[bool]int) map[bool]int
	Chan(...chan bool) (chan<- int, <-chan int8)
}
