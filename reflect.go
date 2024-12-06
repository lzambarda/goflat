package goflat

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type structFactory[T any] struct {
	structType   reflect.Type
	pointer      bool
	columnMap    map[int]int
	columnValues []any
	columnNames  []string
	options      Options
}

// FieldTag is the tag that must be used in the struct fields so that goflat can
// work with them.
const FieldTag = "flat"

//nolint:varnamelen,cyclop // Fine-ish here.
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
		structType:   t,
		pointer:      pointer,
		columnMap:    make(map[int]int, len(headers)),
		columnValues: make([]any, t.NumField()),
		columnNames:  make([]string, t.NumField()),
		options:      options,
	}

	covered := make([]bool, len(headers))

	for i := range t.NumField() {
		fieldT := t.Field(i)
		fieldV := rv.Field(i)

		factory.columnValues[i] = fieldV.Interface()

		v, ok := fieldT.Tag.Lookup(FieldTag)
		if !ok && options.ErrorIfTaglessField {
			return nil, fmt.Errorf("field %q breaks strict mode: %w", fieldT.Name, ErrTaglessField)
		}

		if v == "" || v == "-" {
			continue
		}

		factory.columnNames[i] = v

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

//nolint:forcetypeassert,gocyclo,cyclop,ireturn // Fine for now.
func (s *structFactory[T]) unmarshal(record []string) (T, error) {
	var zero T

	newStruct := reflect.New(s.structType).Elem()

	var value any
	var err error

	//nolint:varnamelen // Fine here.
	for i, column := range record {
		mappedIndex, found := s.columnMap[i]
		if !found {
			continue
		}

		if column == "" && s.options.UnmarshalIgnoreEmpty {
			continue
		}

		columnBaseValue := s.columnValues[mappedIndex]

		// special case
		if u, ok := columnBaseValue.(Unmarshaller); ok {
			value, err = u.Unmarshal(column)
		} else {
			switch columnBaseValue.(type) {
			case bool:
				value, err = strconv.ParseBool(column)
			case int:
				value, err = strconv.Atoi(column)
			case int8:
				value, err = strconv.ParseInt(column, 10, 8)
				value = int8(value.(int64)) //nolint:gosec // Safe.
			case int16:
				value, err = strconv.ParseInt(column, 10, 16)
				value = uint16(value.(int64)) //nolint:gosec // Safe.
			case int32:
				value, err = strconv.ParseInt(column, 10, 32)
				value = int32(value.(int64)) //nolint:gosec // Safe.
			case int64:
				value, err = strconv.ParseInt(column, 10, 64)
			case uint:
				value, err = strconv.Atoi(column)
				value = uint(value.(int)) //nolint:gosec // Safe.
			case uint8: // aka byte
				value, err = strconv.ParseUint(column, 10, 8)
				value = uint8(value.(uint64)) //nolint:gosec // Safe.
			case uint16:
				value, err = strconv.ParseUint(column, 10, 16)
				value = uint16(value.(uint64)) //nolint:gosec // Safe.
			case uint32:
				value, err = strconv.ParseUint(column, 10, 32)
				value = uint32(value.(uint64)) //nolint:gosec // Safe.
			case uint64:
				value, err = strconv.ParseUint(column, 10, 64)
			case float32:
				value, err = strconv.ParseFloat(column, 32)
				value = float32(value.(float64))
			case float64:
				value, err = strconv.ParseFloat(column, 64)
			case string:
				value = column
			default:
				return zero, fmt.Errorf("type %T: %w", columnBaseValue, ErrUnsupportedType)
			}
		}

		if err != nil {
			return zero, fmt.Errorf("parse column %d: %w", i, err)
		}

		newStruct.Field(mappedIndex).Set(reflect.ValueOf(value))
	}

	if s.pointer {
		newStruct = newStruct.Addr()
	}

	return newStruct.Interface().(T), nil
}

func (s *structFactory[T]) marshalHeaders() []string {
	headers := []string{}

	for _, name := range s.columnNames {
		if name == "" {
			continue
		}

		headers = append(headers, name)
	}

	return headers
}

func (s *structFactory[T]) marshal(t T, separator string) ([]string, error) {
	reflectValue := reflect.ValueOf(t)

	if s.pointer {
		reflectValue = reflectValue.Elem()
	}

	record := make([]string, 0, len(s.columnNames))

	var strValue string
	var err error

	//nolint:varnamelen // Fine here.
	for i, name := range s.columnNames {
		if name == "" {
			continue
		}

		fieldV := reflectValue.Field(i)

		// special case
		if m, ok := fieldV.Interface().(Marshaller); ok {
			strValue, err = m.Marshal()
			if err != nil {
				return nil, fmt.Errorf("marshal column %d: %w", i, err)
			}
		} else {
			strValue = fmt.Sprintf("%v", fieldV.Interface())
		}

		strValue = strings.ReplaceAll(strValue, separator, "\\"+separator)

		record = append(record, strValue)
	}

	record = record[0:len(record):len(record)]

	return record, nil
}
