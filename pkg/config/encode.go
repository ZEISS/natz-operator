package config

import (
	"bytes"
	"reflect"
	"strconv"

	"github.com/zeiss/pkg/reflectx"
)

const (
	DefaulTabSize  = 4
	DefaultLineLen = 70
)

// Encoder ...
type Encoder struct {
	len int
	tab int
	bytes.Buffer
}

// Marshaler ...
type Marshaler interface {
	Marshal() ([]byte, error)
}

// Marshal ...
func Marshal(v interface{}) ([]byte, error) {
	e := NewEncoder()

	err := e.marshal(v)
	if err != nil {
		return nil, err
	}

	buf := append([]byte(nil), e.Bytes()...)

	return buf, nil
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder() *Encoder {
	return &Encoder{
		tab: DefaulTabSize,
		len: DefaultLineLen,
	}
}

func (e *Encoder) marshal(v interface{}) error {
	e.reflectValue(reflect.ValueOf(v))

	return nil
}

func (e *Encoder) reflectValue(v reflect.Value) {
	valueEncoder(v)(e, v)
}

func (e *Encoder) error(err error) {
	panic(err)
}

type encoderFunc func(e *Encoder, v reflect.Value)

func valueEncoder(v reflect.Value) encoderFunc {
	if !v.IsValid() {
		return invalidValueEncoder
	}

	return typeEncoder(v.Type())
}

func typeEncoder(t reflect.Type) encoderFunc {
	return newTypeEncoder(t, true)
}

func newTypeEncoder(t reflect.Type, allowAddr bool) encoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	// case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 	return intEncoder
	// case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	// 	return uintEncoder
	// case reflect.Float32, reflect.Float64:
	// 	return floatEncoder
	// case reflect.String:
	// 	return stringEncoder
	// case reflect.Slice:
	// 	return sliceEncoder
	// case reflect.Map:
	// 	return mapEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	// case reflect.Ptr:
	// 	if allowAddr {
	// 		return ptrEncoder
	// 	}
	default:
		return unsupportedTypeEncoder
	}
}

func boolEncoder(e *Encoder, v reflect.Value) {
	b := e.AvailableBuffer()
	b = strconv.AppendBool(b, v.Bool())
	e.Write(b)
}

func unsupportedTypeEncoder(e *Encoder, v reflect.Value) {
	e.error(&UnsupportedTypeError{v.Type()})
}

func invalidValueEncoder(e *Encoder, v reflect.Value) {
	e.WriteString("null")
}

type structEncoder struct {
	fields structFields
}

type field struct {
	name      string
	nameBytes []byte // []byte(name)

	tag       bool
	index     []int
	typ       reflect.Type
	omitEmpty bool
	quoted    bool

	encoder encoderFunc
}

type structFields struct {
	list         []field
	byExactName  map[string]*field
	byFoldedName map[string]*field
}

func typeField(t reflect.Type, index []int) structFields {
	current := []field{}
	next := []field{{typ: t}}

	var count, nextCount map[reflect.Type]int

	visited := map[reflect.Type]bool{}

	var fields []field
	// var nameEscBuf []byte

	for len(next) > 0 {
		current, next = next, current[:0]
		count, nextCount = nextCount, map[reflect.Type]int{}

		for _, f := range current {
			if visited[f.typ] {
				continue
			}
			visited[f.typ] = true

			// Scan f.typ for fields to include.
			for i := 0; i < f.typ.NumField(); i++ {
				sf := f.typ.Field(i)
				if sf.Anonymous {
					t := sf.Type
					if t.Kind() == reflect.Pointer {
						t = t.Elem()
					}
					if !sf.IsExported() && t.Kind() != reflect.Struct {
						// Ignore embedded fields of unexported non-struct types.
						continue
					}
					// Do not ignore embedded fields of unexported struct types
					// since they may have exported fields.
				} else if !sf.IsExported() {
					// Ignore unexported non-embedded fields.
					continue
				}
				tag := sf.Tag.Get("json")
				if tag == "-" {
					continue
				}

				name, opts := reflectx.ParseTag(tag)
				if !reflectx.IsValidTag(name) {
					name = ""
				}

				index := make([]int, len(f.index)+1)
				copy(index, f.index)
				index[len(f.index)] = i

				ft := sf.Type
				if ft.Name() == "" && ft.Kind() == reflect.Pointer {
					// Follow pointer.
					ft = ft.Elem()
				}

				// Only strings, floats, integers, and booleans can be quoted.
				quoted := false

				if opts.Contains("string") {
					switch ft.Kind() {
					case reflect.Bool,
						reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
						reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
						reflect.Float32, reflect.Float64,
						reflect.String:
						quoted = true
					}
				}

				// Record found field and index sequence.
				if name != "" || !sf.Anonymous || ft.Kind() != reflect.Struct {
					tagged := name != ""
					if name == "" {
						name = sf.Name
					}
					field := field{
						name:      name,
						tag:       tagged,
						index:     index,
						typ:       ft,
						omitEmpty: opts.Contains("omitempty"),
						quoted:    quoted,
					}
					field.nameBytes = []byte(field.name)

					fields = append(fields, field)
					if count[f.typ] > 1 {
						// If there were multiple instances, add a second,
						// so that the annihilation code will see a duplicate.
						// It only cares about the distinction between 1 and 2,
						// so don't bother generating any more copies.
						fields = append(fields, fields[len(fields)-1])
					}
					continue
				}

				// Record new anonymous struct to explore in next round.
				nextCount[ft]++
				if nextCount[ft] == 1 {
					next = append(next, field{name: ft.Name(), index: index, typ: ft})
				}
			}
		}
	}

	return structFields{list: fields}
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: typeField(t, nil)}
	return se.encode
}

func (se structEncoder) encode(e *Encoder, v reflect.Value) {}

// UnsupportedTypeError ...
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "unsupported type: " + e.Type.String()
}
