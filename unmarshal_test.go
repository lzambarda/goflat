package goflat_test

import (
	"context"
	"embed"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/lzambarda/goflat"
)

func TestUnmarshal(t *testing.T) {
	t.Run("error empty", testUnmarshalErrorEmpty)
	t.Run("success", testUnmarshalSuccess)
	t.Run("success ignore empty", testUnmarshalSuccessIgnoreEmpty)
	t.Run("success pointer", testUnmarshalSuccessPointer)
}

//go:embed testdata
var testdata embed.FS

func testUnmarshalErrorEmpty(t *testing.T) {
	file, err := testdata.Open("testdata/unmarshal/success empty.csv")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	channel := make(chan record)
	assertChannel(t, channel, nil, cmp.AllowUnexported(record{}))

	ctx := context.Background()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, options, channel)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func testUnmarshalSuccess(t *testing.T) {
	file, err := testdata.Open("testdata/unmarshal/success.csv")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	expected := []record{
		{
			FirstName: "Guybrush",
			LastName:  "Threepwood",
			Age:       28,
			Height:    1.78,
		},
		{
			FirstName: "Elaine",
			LastName:  "Marley",
			Age:       20,
			Height:    1.6,
		},
		{
			FirstName: "LeChuck",
			LastName:  "",
			Age:       100,
			Height:    2.01,
		},
	}

	channel := make(chan record)
	assertChannel(t, channel, expected, cmp.AllowUnexported(record{}))

	ctx := context.Background()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, options, channel)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func testUnmarshalSuccessIgnoreEmpty(t *testing.T) {
	file, err := testdata.Open("testdata/unmarshal/success empty.csv")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	expected := []record{
		{
			FirstName: "Guybrush",
			LastName:  "Threepwood",
			Age:       28,
			Height:    0,
		},
		{
			FirstName: "Elaine",
			LastName:  "Marley",
			Age:       0,
			Height:    1.6,
		},
		{
			FirstName: "LeChuck",
			LastName:  "",
			Age:       0,
			Height:    0,
		},
	}

	channel := make(chan record)
	assertChannel(t, channel, expected, cmp.AllowUnexported(record{}))

	ctx := context.Background()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
		UnmarshalIgnoreEmpty:    true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, options, channel)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func testUnmarshalSuccessPointer(t *testing.T) {
	file, err := testdata.Open("testdata/unmarshal/success.csv")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  string  `flat:"last_name"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	expected := []*record{
		{
			FirstName: "Guybrush",
			LastName:  "Threepwood",
			Age:       28,
			Height:    1.78,
		},
		{
			FirstName: "Elaine",
			LastName:  "Marley",
			Age:       20,
			Height:    1.6,
		},
		{
			FirstName: "LeChuck",
			LastName:  "",
			Age:       100,
			Height:    2.01,
		},
	}

	channel := make(chan *record)
	assertChannel(t, channel, expected, cmp.AllowUnexported(record{}))

	ctx := context.Background()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, options, channel)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func assertChannel[T any](t *testing.T, ch <-chan T, expected []T, cmpOpts ...cmp.Option) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	var got []T

	go func() {
		defer cancel()

		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-ch:
				if !ok {
					return
				}

				got = append(got, v)
			}
		}
	}()

	t.Cleanup(func() {
		<-ctx.Done()

		if diff := cmp.Diff(expected, got, cmpOpts...); diff != "" {
			t.Errorf("(-expected,+got):\n%s", diff)
		}
	})
}
