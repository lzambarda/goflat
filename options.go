package goflat

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

// StrictOptions returns an [Options] struct with all options set to the strict
// value. This is a good general-purpose struct if you have no special needs.
func StrictOptions() Options {
	return Options{
		ErrorIfTaglessField:     true,
		ErrorIfDuplicateHeaders: true,
		ErrorIfMissingHeaders:   true,
		UnmarshalIgnoreEmpty:    false,
	}
}
