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
	t.Run("success", testUnmarshalSuccess)
}

//go:embed testdata
var testdata embed.FS

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

	expected := []record{}

	channel := make(chan record)
	go assertChannel(t, channel, expected)

	ctx := context.Background()

	csvReader, err := goflat.DetectReader(file)
	if err != nil {
		t.Fatalf("detect reader: %v", err)
	}

	options := goflat.Options{
		Strict:                  true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
	}

	err = goflat.UnmarshalToChannel(ctx, csvReader, options, channel)
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

func assertChannel[T any](t *testing.T, ch <-chan T, expected []T) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	var got []T

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case v, ok := <-ch:
			if !ok {
				break loop
			}

			got = append(got, v)
		}
	}

	var zero T
	if diff := cmp.Diff(expected, got, cmp.AllowUnexported(zero)); diff != "" {
		t.Errorf("(-expected,+got):\n%s", diff)
	}
}
