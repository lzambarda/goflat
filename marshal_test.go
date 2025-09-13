package goflat_test

import (
	"bytes"
	"encoding/csv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/lzambarda/goflat"
)

func TestMarshal(t *testing.T) {
	t.Run("escaping", testMarshalEscaping)
	t.Run("success", testMarshalSuccess)
	t.Run("success pointer", testMarshalSuccessPointer)
}

func testMarshalEscaping(t *testing.T) {
	type foo struct {
		ID   int    `flat:"id"`
		Name string `flat:"name"`
	}

	values := []foo{
		{ID: 1, Name: "LENNON, JOHN"},
		{ID: 2, Name: `RICHARD "RINGO" STARR`},
		{ID: 3, Name: `PAUL`},
	}

	t.Run("comma", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		writer := csv.NewWriter(buffer)

		err := goflat.MarshalSliceToWriter(t.Context(), values, writer, goflat.Options{})
		if err != nil {
			t.Fatal(err)
		}

		expected := `id,name
1,"LENNON, JOHN"
2,"RICHARD ""RINGO"" STARR"
3,PAUL
`
		if diff := cmp.Diff(expected, buffer.String()); diff != "" {
			t.Errorf("(-expected, +got):\n%s", diff)
		}
	})

	t.Run("tab", func(t *testing.T) {
		buffer := &bytes.Buffer{}
		writer := csv.NewWriter(buffer)
		writer.Comma = '\t'

		err := goflat.MarshalSliceToWriter(t.Context(), values, writer, goflat.Options{})
		if err != nil {
			t.Fatal(err)
		}

		expected := `id	name
1	LENNON, JOHN
2	"RICHARD ""RINGO"" STARR"
3	PAUL
`
		if diff := cmp.Diff(expected, buffer.String()); diff != "" {
			t.Errorf("(-expected, +got):\n%s", diff)
		}
	})
}

func testMarshalSuccess(t *testing.T) {
	expected, err := testdata.ReadFile("testdata/marshal/success.csv")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}

	type record struct {
		FirstName    string  `flat:"first_name"`
		LastName     string  `flat:"last_name"`
		Ignore       uint8   `flat:"-"`
		Age          int     `flat:"age"`
		Height       float32 `flat:"height"`
		OptionalName *string `flat:"optional_name"`
	}

	input := []record{
		{
			FirstName:    "John, Sir",
			LastName:     "Doe",
			Ignore:       123,
			Age:          30,
			Height:       1.75,
			OptionalName: nil,
		},
		{
			FirstName:    "Jane",
			LastName:     "Doe",
			Ignore:       123,
			Age:          25,
			Height:       1.65,
			OptionalName: ptrTo(""),
		},
		{
			FirstName:    "John",
			LastName:     "Smith",
			Ignore:       123,
			Age:          40,
			Height:       2.00,
			OptionalName: ptrTo("Secret"),
		},
	}

	tcs := map[string]goflat.Options{
		"simple": {},
		"strict": {
			ErrorIfTaglessField:     true,
			ErrorIfDuplicateHeaders: true,
			ErrorIfMissingHeaders:   true,
			UnmarshalIgnoreEmpty:    true,
		},
	}

	for name, options := range tcs {
		t.Run(name, func(t *testing.T) {
			var got bytes.Buffer

			writer := csv.NewWriter(&got)

			err = goflat.MarshalSliceToWriter(t.Context(), input, writer, options)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}

			if diff := cmp.Diff(string(expected), got.String()); diff != "" {
				t.Errorf("(-expected, +got):\n%s", diff)
			}
		})
	}
}

func testMarshalSuccessPointer(t *testing.T) {
	expected, err := testdata.ReadFile("testdata/marshal/success.csv")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}

	type record struct {
		FirstName    string  `flat:"first_name"`
		LastName     string  `flat:"last_name"`
		Ignore       uint8   `flat:"-"`
		Age          int     `flat:"age"`
		Height       float32 `flat:"height"`
		OptionalName *string `flat:"optional_name"`
	}

	input := []*record{
		{
			FirstName:    "John, Sir",
			LastName:     "Doe",
			Ignore:       123,
			Age:          30,
			Height:       1.75,
			OptionalName: nil,
		},
		{
			FirstName:    "Jane",
			LastName:     "Doe",
			Ignore:       123,
			Age:          25,
			Height:       1.65,
			OptionalName: ptrTo(""),
		},
		{
			FirstName:    "John",
			LastName:     "Smith",
			Ignore:       123,
			Age:          40,
			Height:       2.00,
			OptionalName: ptrTo("Secret"),
		},
	}

	var got bytes.Buffer

	writer := csv.NewWriter(&got)

	err = goflat.MarshalSliceToWriter(t.Context(), input, writer, goflat.Options{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if diff := cmp.Diff(string(expected), got.String()); diff != "" {
		t.Errorf("(-expected, +got):\n%s", diff)
	}
}
