package goflat

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	"golang.org/x/sync/errgroup"
)

// Unmarshaller can be used to tell goflat to use custom logic to convert the
// input string into the type itself.
type Unmarshaller interface {
	Unmarshal(value string) (Unmarshaller, error)
}

// UnmarshalToChannel unmarshals a CSV file to a channel of structs. It
// automatically closes the channel at the end.
func UnmarshalToChannel[T any](ctx context.Context, reader *csv.Reader, outputCh chan<- T, opts Options) error {
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
		return UnmarshalToChannel(ctx, reader, ch, opts)
	})

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("wait: %w", err)
	}

	return slice, nil
}

// UnmarshalToCallback unamrshals a CSV file invoking a callback function on
// each row.
func UnmarshalToCallback[T any](ctx context.Context, reader *csv.Reader, opts Options, callback func(T) error) error {
	g, ctx := errgroup.WithContext(ctx) //nolint:varnamelen // Fine here.

	ch := make(chan T) //nolint:varnamelen // Fine here.

	g.Go(func() error {
		for v := range ch {
			if err := callback(v); err != nil {
				return fmt.Errorf("callback: %w", err)
			}
		}

		return nil
	})

	g.Go(func() error {
		return UnmarshalToChannel(ctx, reader, ch, opts)
	})

	if err := g.Wait(); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	return nil
}
