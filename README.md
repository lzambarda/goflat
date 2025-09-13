# goflat
![Coverage](https://img.shields.io/badge/Coverage-81.2%25-brightgreen)

[![Go Reference](https://pkg.go.dev/badge/github.com/lzambarda/goflat.svg)](https://pkg.go.dev/github.com/lzambarda/goflat)
[![Go Report Card](https://goreportcard.com/badge/github.com/lzambarda/goflat)](https://goreportcard.com/report/github.com/lzambarda/goflat)

Generic, context-aware flat file marshaller and unmarshaller using the `flat` field tag in structs.

## Overview

```go
type Record struct {
    FirstName string  `flat:"first_name"`
    LastName  *string `flat:"last_name"`
    Age       int     `flat:"age"`
    Height    float32 `flat:"-"` // ignored
}

...

goflat.MarshalIteratorToWriter[Record](ctx, seq, writer *csv.Writer, opts Options)

// or

goflat.MarshalSliceToWriter[Record](ctx, slice, csvWriter, options)

// or

goflat.MarshalChannelToWriter[Record](ctx, inputCh, csvWriter, options)

```

Will result in:

```
first_name,last_name,age
John,Doe,30
Jane,Doe,20
```

Can unmarshal too!

```go
goflat.UnmarshalToChan[Record](ctx, csvReader, options, outputCh)
```

## Options

Both marshal and unmarshal operations support `goflat.Options`, which allow to introduce automatic safety checks, such as duplicated headers, `flat` tag coverage and more.

## Custom marshal / unmarshal

Both operations can be customised for each field in a struct by having that value implementing `goflat.Marshal` and/or `goflat.Unmarshal`.

```go
type Record struct {
    Field MyType `flat:"field"`
}

type MyType struct {
    Value int
}

func (m *MyType) Marshal() (string,error) {
    if m.Value %2 == 0 {
        return "odd", nil
    }

    return "even", nil
}
```
