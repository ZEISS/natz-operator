package config

import (
	"bytes"
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strconv"
	"strings"

	"github.com/zeiss/pkg/reflectx"
)

const (
	DefaulTabSize  = 4
	DefaultLineLen = 70
)

// A Number represents a JSON number literal.
type Number string

var numberType = reflect.TypeFor[Number]()

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

// nolint:unparam
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
	return newTypeEncoder(t)
}

type isZeroer interface {
	IsZero() bool
}

var isZeroerType = reflect.TypeFor[isZeroer]()

func newTypeEncoder(t reflect.Type) encoderFunc {
	// nolint:exhaustive
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intEncoder
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return uintEncoder
	// case reflect.Float32, reflect.Float64:
	// 	return floatEncoder
	case reflect.String:
		return stringEncoder
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

func dominantField(fields []field) (field, bool) {
	// The fields are sorted in increasing index-length order, then by presence of tag.
	// That means that the first field is the dominant one. We need only check
	// for error cases: two fields at top level, either both tagged or neither tagged.
	if len(fields) > 1 && len(fields[0].index) == len(fields[1].index) && fields[0].tag == fields[1].tag {
		return field{}, false
	}
	return fields[0], true
}

func typeByIndex(t reflect.Type, index []int) reflect.Type {
	for _, i := range index {
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}
		t = t.Field(i).Type
	}
	return t
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
	isZero    func(reflect.Value) bool
	omitEmpty bool
	omitZero  bool
	quoted    bool

	encoder encoderFunc
}

func intEncoder(e *Encoder, v reflect.Value) {
	b := e.AvailableBuffer()
	b = strconv.AppendInt(b, v.Int(), 10)
	e.Write(b)
}

func uintEncoder(e *Encoder, v reflect.Value) {
	b := e.AvailableBuffer()
	b = strconv.AppendUint(b, v.Uint(), 10)
	e.Write(b)
}

type structFields struct {
	list         []field
	byExactName  map[string]*field
	byFoldedName map[string]*field
}

// nolint:gocyclo
func typeField(t reflect.Type) structFields {
	// Anonymous fields to explore at the current level and the next.
	current := []field{}
	next := []field{{typ: t}}

	// Count of queued names for current level and the next.
	var count, nextCount map[reflect.Type]int

	// Types already visited at an earlier level.
	visited := map[reflect.Type]bool{}

	// Fields found.
	var fields []field

	// Buffer to run appendHTMLEscape on field names.
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
					// nolint:exhaustive
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
						omitZero:  opts.Contains("omitzero"),
						quoted:    quoted,
					}
					field.nameBytes = []byte(field.name)

					if field.omitZero {
						t := sf.Type
						// Provide a function that uses a type's IsZero method.
						switch {
						case t.Kind() == reflect.Interface && t.Implements(isZeroerType):
							field.isZero = func(v reflect.Value) bool {
								// Avoid panics calling IsZero on a nil interface or
								// non-nil interface with nil pointer.
								return v.IsNil() ||
									(v.Elem().Kind() == reflect.Pointer && v.Elem().IsNil()) ||
									v.Interface().(isZeroer).IsZero()
							}
						case t.Kind() == reflect.Pointer && t.Implements(isZeroerType):
							field.isZero = func(v reflect.Value) bool {
								// Avoid panics calling IsZero on nil pointer.
								return v.IsNil() || v.Interface().(isZeroer).IsZero()
							}
						case t.Implements(isZeroerType):
							field.isZero = func(v reflect.Value) bool {
								return v.Interface().(isZeroer).IsZero()
							}
						case reflect.PointerTo(t).Implements(isZeroerType):
							field.isZero = func(v reflect.Value) bool {
								if !v.CanAddr() {
									// Temporarily box v so we can take the address.
									v2 := reflect.New(v.Type()).Elem()
									v2.Set(v)
									v = v2
								}
								return v.Addr().Interface().(isZeroer).IsZero()
							}
						}
					}

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

	slices.SortFunc(fields, func(a, b field) int {
		// sort field by name, breaking ties with depth, then
		// breaking ties with "name came from json tag", then
		// breaking ties with index sequence.
		if c := strings.Compare(a.name, b.name); c != 0 {
			return c
		}
		if c := cmp.Compare(len(a.index), len(b.index)); c != 0 {
			return c
		}
		if a.tag != b.tag {
			if a.tag {
				return -1
			}
			return +1
		}
		return slices.Compare(a.index, b.index)
	})

	// Delete all fields that are hidden by the Go rules for embedded fields,
	// except that fields with JSON tags are promoted.

	// The fields are sorted in primary order of name, secondary order
	// of field index length. Loop over names; for each name, delete
	// hidden fields by choosing the one dominant field that survives.
	out := fields[:0]
	for advance, i := 0, 0; i < len(fields); i += advance {
		// One iteration per name.
		// Find the sequence of fields with the name of this first field.
		fi := fields[i]
		name := fi.name
		for advance = 1; i+advance < len(fields); advance++ {
			fj := fields[i+advance]
			if fj.name != name {
				break
			}
		}
		if advance == 1 { // Only one field with this name
			out = append(out, fi)
			continue
		}
		dominant, ok := dominantField(fields[i : i+advance])
		if ok {
			out = append(out, dominant)
		}
	}

	fields = out
	slices.SortFunc(fields, func(i, j field) int {
		return slices.Compare(i.index, j.index)
	})

	for i := range fields {
		f := &fields[i]
		f.encoder = typeEncoder(typeByIndex(t, f.index))
	}
	exactNameIndex := make(map[string]*field, len(fields))
	foldedNameIndex := make(map[string]*field, len(fields))
	for i, field := range fields {
		exactNameIndex[field.name] = &fields[i]
		if _, ok := foldedNameIndex[string(reflectx.FoldName(field.nameBytes))]; !ok {
			foldedNameIndex[string(reflectx.FoldName(field.nameBytes))] = &fields[i]
		}
	}
	return structFields{fields, exactNameIndex, foldedNameIndex}
}

func stringEncoder(e *Encoder, v reflect.Value) {
	if v.Type() == numberType {
		numStr := v.String()
		// In Go1.5 the empty string encodes to "0", while this is not a valid number literal
		// we keep compatibility so check validity after this.
		if numStr == "" {
			numStr = "0" // Number's zero-val
		}
		if !isValidNumber(numStr) {
			e.error(fmt.Errorf("json: invalid number literal %q", numStr))
		}
		b := e.AvailableBuffer()
		b = append(b, numStr...)

		e.Write(b)

		return
	}

	e.Write(appendString(e.AvailableBuffer(), v.String()))
}

func appendString[Bytes []byte | string](dst []byte, src Bytes) []byte {
	dst = append(dst, ' ')
	start := 0
	// for i := 0; i < len(src); {
	// 	if b := src[i]; b < utf8.RuneSelf {
	// 		dst = append(dst, src[start:i]...)
	// 		switch b {
	// 		case '\\', '"':
	// 			dst = append(dst, '\\', b)
	// 		case '\b':
	// 			dst = append(dst, '\\', 'b')
	// 		case '\f':
	// 			dst = append(dst, '\\', 'f')
	// 		case '\n':
	// 			dst = append(dst, '\\', 'n')
	// 		case '\r':
	// 			dst = append(dst, '\\', 'r')
	// 		case '\t':
	// 			dst = append(dst, '\\', 't')
	// 		default:
	// 			// This encodes bytes < 0x20 except for \b, \f, \n, \r and \t.
	// 			// If escapeHTML is set, it also escapes <, >, and &
	// 			// because they can lead to security holes when
	// 			// user-controlled strings are rendered into JSON
	// 			// and served to some browsers.
	// 			dst = append(dst, '\\', 'u', '0', '0', hex[b>>4], hex[b&0xF])
	// 		}
	// 		i++
	// 		start = i
	// 		continue
	// 	}

	// 	// TODO(https://go.dev/issue/56948): Use generic utf8 functionality.
	// 	// For now, cast only a small portion of byte slices to a string
	// 	// so that it can be stack allocated. This slows down []byte slightly
	// 	// due to the extra copy, but keeps string performance roughly the same.
	// 	n := len(src) - i
	// 	if n > utf8.UTFMax {
	// 		n = utf8.UTFMax
	// 	}

	// 	c, size := utf8.DecodeRuneInString(string(src[i : i+n]))
	// 	if c == utf8.RuneError && size == 1 {
	// 		dst = append(dst, src[start:i]...)
	// 		dst = append(dst, `\ufffd`...)
	// 		i += size
	// 		start = i
	// 		continue
	// 	}
	// 	// U+2028 is LINE SEPARATOR.
	// 	// U+2029 is PARAGRAPH SEPARATOR.
	// 	// They are both technically valid characters in JSON strings,
	// 	// but don't work in JSONP, which has to be evaluated as JavaScript,
	// 	// and can lead to security holes there. It is valid JSON to
	// 	// escape them, so we do so unconditionally.
	// 	// See https://en.wikipedia.org/wiki/JSON#Safety.
	// 	if c == '\u2028' || c == '\u2029' {
	// 		dst = append(dst, src[start:i]...)
	// 		dst = append(dst, '\\', 'u', '2', '0', '2', hex[c&0xF])
	// 		i += size
	// 		start = i
	// 		continue
	// 	}
	// 	i += size
	// }

	dst = append(dst, src[start:]...)

	return dst
}

// nolint:gocyclo
func isValidNumber(s string) bool {
	// This function implements the JSON numbers grammar.
	// See https://tools.ietf.org/html/rfc7159#section-6
	// and https://www.json.org/img/number.png

	if s == "" {
		return false
	}

	// Optional -
	if s[0] == '-' {
		s = s[1:]
		if s == "" {
			return false
		}
	}

	// Digits
	switch {
	default:
		return false

	case s[0] == '0':
		s = s[1:]

	case '1' <= s[0] && s[0] <= '9':
		s = s[1:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// . followed by 1 or more digits.
	if len(s) >= 2 && s[0] == '.' && '0' <= s[1] && s[1] <= '9' {
		s = s[2:]
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// e or E followed by an optional - or + and
	// 1 or more digits.
	if len(s) >= 2 && (s[0] == 'e' || s[0] == 'E') {
		s = s[1:]
		if s[0] == '+' || s[0] == '-' {
			s = s[1:]
			if s == "" {
				return false
			}
		}
		for len(s) > 0 && '0' <= s[0] && s[0] <= '9' {
			s = s[1:]
		}
	}

	// Make sure we are at the end.
	return s == ""
}

func newStructEncoder(t reflect.Type) encoderFunc {
	se := structEncoder{fields: typeField(t)}
	return se.encode
}

func (se structEncoder) encode(e *Encoder, v reflect.Value) {
	next := "{"
FieldLoop:
	for i := range se.fields.list {
		f := &se.fields.list[i]

		// Find the nested struct field by following f.index.
		fv := v
		for _, i := range f.index {
			if fv.Kind() == reflect.Pointer {
				if fv.IsNil() {
					continue FieldLoop
				}
				fv = fv.Elem()
			}
			fv = fv.Field(i)
		}

		if f.omitEmpty && reflectx.IsEmptyValue(fv) {
			continue
		}
		e.WriteString(next)
		next = "\n"

		e.WriteString(f.name + `:`)

		f.encoder(e, fv)
	}

	if next == "{" {
		e.WriteString("")
	} else {
		e.WriteByte('}')
	}
}

// UnsupportedTypeError ...
type UnsupportedTypeError struct {
	Type reflect.Type
}

func (e *UnsupportedTypeError) Error() string {
	return "unsupported type: " + e.Type.String()
}
