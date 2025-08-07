package goflat

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Options is used to configure the marshalling and unmarshalling processes.
type Options struct {
	headersFromStruct bool
	// ErrorIfTaglessField causes goflat to error out if any struct field is
	// missing the `flat` tag.
	ErrorIfTaglessField bool
	// ErrorIfDuplicateHeaders causes goflat to error out if two struct fields
	// share the same `flat` tag value.
	ErrorIfDuplicateHeaders bool
	// ErrorIfMissingHeaders causes goflat to error out at unmarshalling time if
	// a header has no struct field with a corresponding `flat` tag.
	ErrorIfMissingHeaders bool
	// UnmarshalIgnoreEmpty causes the unmarshaller to skip any column which is
	// an empty string. This is useful for instance if you have integer values
	// and you are okay with empty string mapping to the zero value (0). For the
	// same reason this will cause booleans to be false if the column is empty.
	UnmarshalIgnoreEmpty bool
}

type structFactory[T any] struct {
	structType reflect.Type
	pointer    bool
	columnMap  map[int]int
	columns    []*columnDescriptor
	options    Options
}

type columnDescriptor struct {
	name        string
	value       any
	reflectType reflect.Type
}

// FieldTag is the tag that must be used in the struct fields so that goflat can
// work with them.
const FieldTag = "flat"

//nolint:varnamelen,cyclop,gocyclo // Fine-ish here.
func newFactory[T any](headers []string, options Options) (*structFactory[T], error) {
	var v T

	t := reflect.TypeOf(v)
	rv := reflect.ValueOf(v)

	pointer := false

	//nolint:exhaustive // Fine here, there's a default.
	switch t.Kind() {
	case reflect.Struct:
	case reflect.Pointer:
		pointer = true
		t = t.Elem()
		rv = reflect.New(t).Elem()
	default:
		return nil, fmt.Errorf("type %T: %w", v, ErrNotAStruct)
	}

	factory := &structFactory[T]{
		structType: t,
		pointer:    pointer,
		columnMap:  make(map[int]int, len(headers)),
		columns:    make([]*columnDescriptor, t.NumField()),
		options:    options,
	}

	covered := make([]bool, len(headers))

	for i := range t.NumField() {
		fieldT := t.Field(i)
		fieldV := rv.Field(i)

		v, ok := fieldT.Tag.Lookup(FieldTag)
		if !ok && options.ErrorIfTaglessField {
			return nil, fmt.Errorf("field %q breaks strict mode: %w", fieldT.Name, ErrTaglessField)
		}

		factory.columns[i] = &columnDescriptor{
			name:        v,
			value:       fieldV.Interface(),
			reflectType: fieldT.Type,
		}

		//nolint:exhaustive // Fine here.
		switch fieldT.Type.Kind() {
		case reflect.Slice:
			factory.columns[i].value = reflect.New(fieldV.Type().Elem()).Elem().Interface()
		case reflect.Pointer:
			factory.columns[i].value = reflect.Zero(fieldV.Type().Elem()).Interface()
		}

		if v == "" || v == "-" {
			factory.columns[i].name = ""

			continue
		}

		if options.headersFromStruct {
			continue
		}

		handledAt := -1

		for j, header := range headers {
			if covered[j] {
				continue
			}

			if header != v {
				continue
			}

			if handledAt >= 0 {
				if options.ErrorIfDuplicateHeaders {
					return nil, fmt.Errorf("header %q, index %d and %d: %w", header, j, handledAt, ErrDuplicatedHeader)
				}

				continue
			}

			handledAt = j
			covered[j] = true
			factory.columnMap[j] = i
		}

		if handledAt == -1 && options.ErrorIfMissingHeaders {
			return nil, fmt.Errorf("header %q: %w", v, ErrMissingHeader)
		}
	}

	return factory, nil
}

//nolint:varnamelen,ireturn // Fine for now.
func (s *structFactory[T]) unmarshal(record []string) (T, error) {
	var zero T

	newStruct := reflect.New(s.structType).Elem()

	for i, column := range record {
		mappedIndex, found := s.columnMap[i]
		if !found {
			continue
		}

		if column == "" && s.options.UnmarshalIgnoreEmpty {
			continue
		}

		columnDescriptor := s.columns[mappedIndex]

		value, err := columnDescriptor.parseColumn(column)
		if err != nil {
			return zero, fmt.Errorf("parse column %d: %w", i, err)
		}

		if columnDescriptor.reflectType.Kind() == reflect.Pointer {
			value = ptr(value)
		}

		newStruct.Field(mappedIndex).Set(reflect.ValueOf(value))
	}

	if s.pointer {
		newStruct = newStruct.Addr()
	}

	return newStruct.Interface().(T), nil //nolint:forcetypeassert // Safe here.
}

// we need to do this because otherwise we get strange behaviour with interface
// pointers.
func ptr(v any) any {
	rv := reflect.ValueOf(v)
	pt := reflect.PointerTo(rv.Type())
	pv := reflect.New(pt.Elem())
	pv.Elem().Set(rv)

	return pv.Interface()
}

//nolint:varnamelen // Fine for now.
func (c *columnDescriptor) parseColumn(column string) (any, error) {
	if c.reflectType.Kind() == reflect.Slice {
		column = strings.Trim(column, "[]{}")
		values := strings.Split(column, ",")
		slice := reflect.MakeSlice(c.reflectType, len(values), len(values))

		// NOTE: text slices with commands inside are currently not supported.
		for i, item := range values {
			v, err := c.parseString(item)
			if err != nil {
				return nil, fmt.Errorf("parse slice index %d, string %q: %w", i, item, err)
			}

			slice.Index(i).Set(reflect.ValueOf(v))
		}

		return slice.Interface(), nil
	}

	v, err := c.parseString(column)
	if err != nil {
		return nil, fmt.Errorf("parse string %q: %w", column, err)
	}

	return v, nil
}

//nolint:gocyclo,cyclop // Fine for now.
func (c *columnDescriptor) parseString(str string) (any, error) {
	// special case
	//nolint:wrapcheck // Fine for now.
	if u, ok := c.value.(Unmarshaller); ok {
		return u.Unmarshal(str)
	}

	var (
		value any
		err   error
	)

	//nolint:forcetypeassert,gosec // Safe context, we know what we're doing.
	switch c.value.(type) {
	case bool:
		value, err = strconv.ParseBool(str)
	case int:
		value, err = strconv.Atoi(str)
	case int8:
		value, err = strconv.ParseInt(str, 10, 8)
		value = int8(value.(int64))
	case int16:
		value, err = strconv.ParseInt(str, 10, 16)
		value = uint16(value.(int64))
	case int32:
		value, err = strconv.ParseInt(str, 10, 32)
		value = int32(value.(int64))
	case int64:
		value, err = strconv.ParseInt(str, 10, 64)
	case uint:
		value, err = strconv.Atoi(str)
		value = uint(value.(int))
	case uint8: // aka byte
		value, err = strconv.ParseUint(str, 10, 8)
		value = uint8(value.(uint64))
	case uint16:
		value, err = strconv.ParseUint(str, 10, 16)
		value = uint16(value.(uint64))
	case uint32:
		value, err = strconv.ParseUint(str, 10, 32)
		value = uint32(value.(uint64))
	case uint64:
		value, err = strconv.ParseUint(str, 10, 64)
	case float32:
		value, err = strconv.ParseFloat(str, 32)
		value = float32(value.(float64))
	case float64:
		value, err = strconv.ParseFloat(str, 64)
	case string:
		value = str
	default:
		return nil, fmt.Errorf("type %T: %w", c.value, ErrUnsupportedType)
	}

	if err != nil {
		return nil, fmt.Errorf("parse string %q: %w", str, err)
	}

	return value, nil
}

func (s *structFactory[T]) marshalHeaders() []string {
	headers := make([]string, 0, len(s.columns))

	for _, column := range s.columns {
		if column.name == "" {
			continue
		}

		headers = append(headers, column.name)
	}

	return headers[0:len(headers):len(headers)]
}

func (s *structFactory[T]) marshal(t T, separator string) ([]string, error) {
	reflectValue := reflect.ValueOf(t)

	if s.pointer {
		reflectValue = reflectValue.Elem()
	}

	record := make([]string, 0, len(s.columns))

	var (
		strValue string
		err      error
	)

	//nolint:varnamelen // Fine for now.
	for i, column := range s.columns {
		if column.name == "" {
			continue
		}

		strValue, err = reflectValueToStr(reflectValue.Field(i))
		if err != nil {
			return nil, fmt.Errorf("column %d: %w", i, err)
		}

		strValue = strings.ReplaceAll(strValue, separator, "\\"+separator)

		record = append(record, strValue)
	}

	record = record[0:len(record):len(record)]

	return record, nil
}

func reflectValueToStr(value reflect.Value) (string, error) {
	const nilStrValue = "nil"

	// Handle pointer values
	if value.Kind() == reflect.Pointer {
		if value.IsNil() {
			return nilStrValue, nil
		}

		value = value.Elem()
	}

	if m, ok := value.Interface().(Marshaller); ok {
		// Custom marshaller special case
		strValue, err := m.Marshal()
		if err != nil {
			return "", fmt.Errorf("marshal: %w", err)
		}

		return strValue, nil
	}

	return fmt.Sprintf("%v", value.Interface()), nil
}
