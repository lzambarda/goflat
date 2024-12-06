package goflat

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"golang.org/x/sync/errgroup"
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

// Unmarshaller can be used to tell goflat to use custom logic to convert the
// input string into the type itself.
type Unmarshaller interface {
	Unmarshal(value string) (Unmarshaller, error)
}

// UnmarshalToChannel unmarshals a CSV file to a channel of structs. It
// automatically closes the channel at the end.
func UnmarshalToChannel[T any](ctx context.Context, reader *csv.Reader, opts Options, outputCh chan<- T) error {
	defer close(outputCh)

	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("read headers: %w", err)
	}

	factory, err := newFactory[T](headers, opts)
	if err != nil {
		return fmt.Errorf("new factory: %w", err)
	}

	var currentLine int

	for {
		record, err := reader.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}

			return fmt.Errorf("read row: %w", err)
		}

		value, err := factory.unmarshal(record)
		if err != nil {
			return fmt.Errorf("get struct at line %d: %w", currentLine, err)
		}

		currentLine++

		select {
		case <-ctx.Done():
			return ctx.Err() //nolint:wrapcheck // No need here.
		case outputCh <- value:
		}
	}
}

// UnmarshalToSlice unmarshals a CSV file to a slice of structs.
func UnmarshalToSlice[T any](ctx context.Context, reader *csv.Reader, opts Options) ([]T, error) {
	g, ctx := errgroup.WithContext(ctx) //nolint:varnamelen // Fine here.

	ch := make(chan T) //nolint:varnamelen // Fine here.

	var slice []T

	g.Go(func() error {
		for v := range ch {
			slice = append(slice, v)
		}

		return nil
	})

	g.Go(func() error {
		defer close(ch)

		return UnmarshalToChannel(ctx, reader, opts, ch)
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	return slice, nil
}
