package goflat_test

import (
	"bytes"
	"context"
	"encoding/csv"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/lzambarda/goflat"
)

func TestMarshal(t *testing.T) {
	t.Run("success", testMarshalSuccess)
	t.Run("success pointer", testMarshalSuccessPointer)
}

func testMarshalSuccess(t *testing.T) {
	expected, err := testdata.ReadFile("testdata/marshal/success.csv")
	if err != nil {
		t.Fatalf("read test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Ignore    uint8   `flat:"-"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	input := []record{
		{
			FirstName: "John",
			LastName:  "Doe",
			Ignore:    123,
			Age:       30,
			Height:    1.75,
		},
		{
			FirstName: "Jane",
			LastName:  "Doe",
			Ignore:    123,
			Age:       25,
			Height:    1.65,
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

			err = goflat.MarshalSliceToWriter(context.Background(), input, writer, options)
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
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Ignore    uint8   `flat:"-"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	input := []*record{
		{
			FirstName: "John",
			LastName:  "Doe",
			Ignore:    123,
			Age:       30,
			Height:    1.75,
		},
		{
			FirstName: "Jane",
			LastName:  "Doe",
			Ignore:    123,
			Age:       25,
			Height:    1.65,
		},
	}
	var got bytes.Buffer

	writer := csv.NewWriter(&got)

	err = goflat.MarshalSliceToWriter(context.Background(), input, writer, goflat.Options{})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	if diff := cmp.Diff(string(expected), got.String()); diff != "" {
		t.Errorf("(-expected, +got):\n%s", diff)
	}
}
