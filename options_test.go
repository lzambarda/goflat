package goflat_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/lzambarda/goflat"
)

func TestOptions(t *testing.T) {
	expectedStrict := goflat.Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
		UnmarshalIgnoreEmpty:    false,
	}

	if diff := cmp.Diff(expectedStrict, goflat.StrictOptions(), cmpopts.EquateComparable(goflat.Options{})); diff != "" {
		t.Errorf("(-expected,+got):\n%s", diff)
	}
}
