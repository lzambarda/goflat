package goflat

import "errors"

var (
	// ErrNotAStruct is returned when the value to be worked with is not a struct.
	ErrNotAStruct = errors.New("not a struct")
	// ErrTaglessField is returned when goflat works in strict mode and a field
	// of the input struct has no "flat" tag.
	ErrTaglessField = errors.New("tagless field")
	// ErrDuplicatedHeader is returned when there is more than one header with
	// the same value. Only returned if [Option.ErrorIfDuplicateHeaders] is set
	// to true.
	ErrDuplicatedHeader = errors.New("duplicated header")
	// ErrMissingHeader is returned when a header referenced in a "flat" tag
	// does not appear in the input file. Only returned if
	// [Option.ErrorIfMissingHeaders] is set to true.
	ErrMissingHeader = errors.New("missing header")
	// ErrUnsupportedType is returned when the unmarshaller encounters an
	// unsupported type.
	ErrUnsupportedType = errors.New("unsupported type")
)
