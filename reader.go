package goflat

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
)

//nolint:gochecknoglobals // We are fine for now.
var commonSeparators = []string{",", ";", "\t", "|"}

func readFirstLine(reader io.Reader) (string, error) {
	b := make([]byte, 1) //nolint:varnamelen // Fine here.

	var line string

	for {
		_, err := reader.Read(b)
		if err != nil {
			if err == io.EOF {
				return line, nil
			}

			return "", fmt.Errorf("read row: %w", err)
		}

		line += string(b)

		if b[0] == '\n' {
			return line, nil
		}
	}
}

// DetectReader returns a CSV reader with a separator based on a best guess
// about the first line.
func DetectReader(reader io.Reader) (*csv.Reader, error) {
	headers, err := readFirstLine(reader)
	if err != nil {
		return nil, fmt.Errorf("read first line: %w", err)
	}

	var (
		bestSeparator string
		bestCount     int
	)

	for _, sep := range commonSeparators {
		count := strings.Count(headers, sep)
		if count > bestCount {
			bestCount = count
			bestSeparator = sep
		}
	}

	// Read headers again
	rr := io.MultiReader(bytes.NewBufferString(headers), reader)

	csvReader := csv.NewReader(rr)
	if bestSeparator != "," {
		csvReader.Comma = rune(bestSeparator[0])
	}

	return csvReader, nil
}
