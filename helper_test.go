package goflat_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

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
