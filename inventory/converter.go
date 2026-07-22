// Package inventory converts Mend inventory JSON reports to CSV.
package inventory

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// URLSource selects which Mend reference URL is written to the CSV report.
type URLSource int

const (
	ReferenceURL URLSource = iota
	POMURL
)

type report struct {
	Libraries []library `json:"libraries"`
}

type library struct {
	Name       string     `json:"name"`
	ArtifactID string     `json:"artifactId"`
	References references `json:"references"`
	Licenses   []license  `json:"licenses"`
}

type references struct {
	URL    *string `json:"url"`
	POMURL *string `json:"pomUrl"`
}

type license struct {
	Name string `json:"name"`
}

// ConvertFile reads a Mend alert JSON file and writes its package inventory as CSV.
func ConvertFile(source, destination string, urlSource URLSource) error {
	input, err := os.Open(source)
	if err != nil {
		return fmt.Errorf("open inventory source %s: %w", source, err)
	}
	defer input.Close()

	output, err := os.Create(destination)
	if err != nil {
		return fmt.Errorf("create inventory CSV %s: %w", destination, err)
	}
	conversionErr := Convert(input, output, urlSource)
	closeErr := output.Close()
	if conversionErr != nil {
		return conversionErr
	}
	if closeErr != nil {
		return fmt.Errorf("close inventory CSV %s: %w", destination, closeErr)
	}
	return nil
}

// Convert converts Mend inventory JSON from input to the legacy inventory CSV format.
func Convert(input io.Reader, output io.Writer, urlSource URLSource) error {
	var inventory report
	if err := json.NewDecoder(input).Decode(&inventory); err != nil {
		return fmt.Errorf("decode inventory JSON: %w", err)
	}

	writer := bufio.NewWriter(output)
	if err := writeRow(writer, []csvValue{
		{text: "name", present: true},
		{text: "artifactId", present: true},
		{text: "url", present: true},
		{text: "licenses_name", present: true},
	}); err != nil {
		return err
	}

	for _, item := range inventory.Libraries {
		url := item.References.URL
		if urlSource == POMURL {
			url = item.References.POMURL
			if url != nil {
				converted := strings.ReplaceAll(*url, ".pom", ".jar")
				url = &converted
			}
		}
		licenseNames := make([]string, 0, len(item.Licenses))
		for _, itemLicense := range item.Licenses {
			licenseNames = append(licenseNames, itemLicense.Name)
		}
		row := []csvValue{
			{text: item.Name, present: true},
			{text: item.ArtifactID, present: true},
			valueFromPointer(url),
			{text: strings.Join(licenseNames, ","), present: true},
		}
		if err := writeRow(writer, row); err != nil {
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("write inventory CSV: %w", err)
	}
	return nil
}

type csvValue struct {
	text    string
	present bool
}

func valueFromPointer(value *string) csvValue {
	if value == nil {
		return csvValue{}
	}
	return csvValue{text: *value, present: true}
}

// writeRow mirrors jq's @csv behavior: JSON strings are always quoted while
// missing URL values are emitted as empty, unquoted CSV fields.
func writeRow(writer *bufio.Writer, row []csvValue) error {
	for index, value := range row {
		if index > 0 {
			if err := writer.WriteByte(','); err != nil {
				return fmt.Errorf("write inventory CSV: %w", err)
			}
		}
		if value.present {
			if err := writer.WriteByte('"'); err != nil {
				return fmt.Errorf("write inventory CSV: %w", err)
			}
			if _, err := writer.WriteString(strings.ReplaceAll(value.text, `"`, `""`)); err != nil {
				return fmt.Errorf("write inventory CSV: %w", err)
			}
			if err := writer.WriteByte('"'); err != nil {
				return fmt.Errorf("write inventory CSV: %w", err)
			}
		}
	}
	if err := writer.WriteByte('\n'); err != nil {
		return fmt.Errorf("write inventory CSV: %w", err)
	}
	return nil
}
