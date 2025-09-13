package goflat

import (
	"context"
	"encoding/csv"
	"fmt"
	"iter"
)

// Marshaller can be used to tell goflat to use custom logic to convert a field
// into a string.
type Marshaller interface {
	Marshal() (string, error)
}

// MarshalSliceToWriter marshals a slice of structs to a CSV file.
//
// NOTE: this is a direct wrapper of [MarshalIteratorToWriter].
func MarshalSliceToWriter[T any](ctx context.Context, values []T, writer *csv.Writer, opts Options) error {
	return MarshalIteratorToWriter(ctx, sliceToIterator(values), writer, opts)
}

// MarshalIteratorToWriter marshals an iterator of structs to a CSV file.
func MarshalIteratorToWriter[T any](ctx context.Context, seq iter.Seq[T], writer *csv.Writer, opts Options) error {
	ch := make(chan T) //nolint:varnamelen // Fine here.

	go func() {
		defer close(ch)

		for value := range seq {
			select {
			case <-ctx.Done():
				return
			case ch <- value:
			}
		}
	}()

	return MarshalChannelToWriter(ctx, ch, writer, opts)
}

// MarshalChannelToWriter marshals a channel of structs to a CSV file.
func MarshalChannelToWriter[T any](ctx context.Context, inputCh <-chan T, writer *csv.Writer, opts Options) error {
	opts.headersFromStruct = true

	factory, err := newFactory[T](nil, opts)
	if err != nil {
		return fmt.Errorf("new factory: %w", err)
	}

	err = writer.Write(factory.marshalHeaders())
	if err != nil {
		return fmt.Errorf("write headers: %w", err)
	}

	var (
		currentLine int
		value       T
	)

	for {
		var channelHasValue bool

		select {
		case <-ctx.Done():
			return context.Cause(ctx) //nolint:wrapcheck // Fine here.
		case value, channelHasValue = <-inputCh:
		}

		if !channelHasValue {
			break
		}

		record, err := factory.marshal(value)
		if err != nil {
			return fmt.Errorf("marshal %d: %w", currentLine, err)
		}

		err = writer.Write(record)
		if err != nil {
			return fmt.Errorf("write line %d: %w", currentLine, err)
		}

		currentLine++
	}

	writer.Flush()

	err = writer.Error()
	if err != nil {
		return fmt.Errorf("flush: %w", err)
	}

	return nil
}

func sliceToIterator[T any](slice []T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, value := range slice {
			if !yield(value) {
				return
			}
		}
	}
}
