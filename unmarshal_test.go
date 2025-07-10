package goflat_test

import (
	"embed"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/lzambarda/goflat"
)

func TestUnmarshal(t *testing.T) {
	t.Run("error", testUnmarshalError)
	t.Run("success", testUnmarshalSuccess)
}

//go:embed testdata
var testdata embed.FS

func testUnmarshalError(t *testing.T) {
	t.Run("empty", testUnmarshalErrorEmpty)
}

func testUnmarshalErrorEmpty(t *testing.T) {
	file, err := testdata.Open("testdata/unmarshal/success empty.csv")
	if err != nil {
		t.Fatalf("open test file: %v", err)
	}

	type record struct {
		FirstName string  `flat:"first_name"`
		LastName  *string `flat:"last_name"`
		Age       int     `flat:"age"`
		Height    float32 `flat:"height"`
	}

	channel := make(chan record)
	assertChannel(t, channel, nil, cmp.AllowUnexported(record{}))

	ctx := t.Context()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, channel, options)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func testUnmarshalSuccess(t *testing.T) {
	t.Run("full", testUnmarshalSuccessFull)
	t.Run("ignore empty", testUnmarshalSuccessIgnoreEmpty)
	t.Run("pointer", testUnmarshalSuccessPointer)
	t.Run("slice", testUnmarshalSuccessSlice)
	t.Run("callback", testUnmarshalSuccessCallback)
}

func testUnmarshalSuccessFull(t *testing.T) {
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

	ctx := t.Context()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, channel, options)
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

	err = goflat.UnmarshalToChannel(t.Context(), csvReader, channel, options)
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

	ctx := t.Context()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, channel, options)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func testUnmarshalSuccessSlice(t *testing.T) {
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

	ctx := t.Context()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	got, err := goflat.UnmarshalToSlice[record](ctx, csvReader, options)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if diff := cmp.Diff(expected, got, cmp.AllowUnexported(record{})); diff != "" {
		t.Errorf("(-expected,+got):\n%s", diff)
	}
}

func testUnmarshalSuccessCallback(t *testing.T) {
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

	ctx := t.Context()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	var got []record

	err = goflat.UnmarshalToCallback(ctx, csvReader, options, func(r record) error {
		got = append(got, r)

		return nil
	})
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if diff := cmp.Diff(expected, got, cmp.AllowUnexported(record{})); diff != "" {
		t.Errorf("(-expected,+got):\n%s", diff)
	}
}
