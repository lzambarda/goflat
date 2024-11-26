package goflat

import (
	"maps"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReflect(t *testing.T) {
	t.Run("error", testReflectError)
	t.Run("success", testReflectSuccess)
}

type s1 struct {
	Foo string
}

func testReflectError(t *testing.T) {
	t.Run("tagless", testReflectErrorTaglessStrict)
	t.Run("missing", testReflectErrorMissing)
	t.Run("duplicate", testReflectErrorDuplicate)
}

func testReflectErrorTaglessStrict(t *testing.T) {
	f, err := newFactory[s1]([]string{}, Options{Strict: true})
	if f != nil {
		t.Errorf("expected nil, got %v", f)
	}

	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func testReflectErrorMissing(t *testing.T) {
	type foo struct {
		Name   string `flat:"name"`
		Age    int    `flat:"age"`
		Skipme string `flat:"-"`
	}

	headers := []string{"name"}

	got, err := newFactory[foo](headers, Options{
		Strict:                true,
		ErrorIfMissingHeaders: true,
	})
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}

	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func testReflectErrorDuplicate(t *testing.T) {
	type foo struct {
		Name   string `flat:"name"`
		Age    int    `flat:"age"`
		Skipme string `flat:"-"`
	}

	headers := []string{"name", "age", "name"}

	got, err := newFactory[foo](headers, Options{
		Strict:                  true,
		ErrorIfDuplicateHeaders: true,
	})
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}

	if err == nil {
		t.Errorf("expected error, got nil")
	}
}

func testReflectSuccess(t *testing.T) {
	t.Run("duplicate", testReflectSuccessDuplicate)
	t.Run("simple", testReflectSuccessSimple)
	t.Run("subset struct", testReflectSuccessSubsetStruct)
}

func testReflectSuccessDuplicate(t *testing.T) {
	type foo struct {
		Name string `flat:"name"`
		Age  int    `flat:"age"`
	}

	headers := []string{"name", "age", "name"}

	got, err := newFactory[foo](headers, Options{
		Strict:                  true,
		ErrorIfDuplicateHeaders: false,
		ErrorIfMissingHeaders:   true,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	expected := &structFactory[foo]{
		structType:   reflect.TypeOf(foo{}),
		columnMap:    map[int]int{0: 0, 1: 1},
		columnValues: []any{"", int(0)},
		columnNames:  []string{"name", "age"},
	}
	comparers := []cmp.Option{
		cmp.AllowUnexported(structFactory[foo]{}),
		cmp.Comparer(func(a, b structFactory[foo]) bool {
			if a.structType.String() != b.structType.String() {
				return false
			}

			if !maps.Equal(a.columnMap, b.columnMap) {
				return false
			}

			return true
		}),
	}

	if diff := cmp.Diff(expected, got, comparers...); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}

func testReflectSuccessSimple(t *testing.T) {
	type foo struct {
		Name   string `flat:"name"`
		Age    int    `flat:"age"`
		Skipme string `flat:"-"`
	}

	headers := []string{"name", "age"}

	got, err := newFactory[foo](headers, Options{
		Strict:                  true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	expected := &structFactory[foo]{
		structType: reflect.TypeOf(foo{}),
		columnMap:  map[int]int{0: 0, 1: 1},
	}
	comparers := []cmp.Option{
		cmp.AllowUnexported(structFactory[foo]{}),
		cmp.Comparer(func(a, b structFactory[foo]) bool {
			if a.structType.String() != b.structType.String() {
				return false
			}

			if !maps.Equal(a.columnMap, b.columnMap) {
				return false
			}

			return true
		}),
	}

	if diff := cmp.Diff(expected, got, comparers...); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}

func testReflectSuccessSubsetStruct(t *testing.T) {
	type foo struct {
		Col2 float32 `flat:"col2"`
	}

	headers := []string{"col1", "col2", "col3"}

	got, err := newFactory[foo](headers, Options{
		Strict:                  false,
		ErrorIfDuplicateHeaders: false,
		ErrorIfMissingHeaders:   false,
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	expected := &structFactory[foo]{
		structType:   reflect.TypeOf(foo{}),
		columnMap:    map[int]int{1: 0},
		columnValues: []any{float32(0)},
	}
	comparers := []cmp.Option{
		cmp.AllowUnexported(structFactory[foo]{}),
		cmp.Comparer(func(a, b structFactory[foo]) bool {
			if a.structType.String() != b.structType.String() {
				return false
			}

			if !maps.Equal(a.columnMap, b.columnMap) {
				return false
			}

			return true
		}),
	}

	if diff := cmp.Diff(expected, got, comparers...); diff != "" {
		t.Errorf("(-want +got):\\n%s", diff)
	}
}
