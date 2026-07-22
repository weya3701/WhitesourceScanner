package inventory

import (
	"strings"
	"testing"
)

const inventoryJSON = `{
  "libraries": [
    {
      "name": "example,lib.jar",
      "artifactId": "example\"artifact",
      "references": {
        "url": "https://example.test/source",
        "pomUrl": "https://repo.test/example.pom"
      },
      "licenses": [{"name": "Apache 2.0"}, {"name": "MIT"}]
    },
    {"name": "no-url", "artifactId": "empty", "references": {}, "licenses": []}
  ]
}`

func TestConvertReferenceURL(t *testing.T) {
	var output strings.Builder
	if err := Convert(strings.NewReader(inventoryJSON), &output, ReferenceURL); err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	want := "\"name\",\"artifactId\",\"url\",\"licenses_name\"\n" +
		"\"example,lib.jar\",\"example\"\"artifact\",\"https://example.test/source\",\"Apache 2.0,MIT\"\n" +
		"\"no-url\",\"empty\",,\"\"\n"
	if output.String() != want {
		t.Fatalf("Convert() = %q, want %q", output.String(), want)
	}
}

func TestConvertPOMURL(t *testing.T) {
	var output strings.Builder
	if err := Convert(strings.NewReader(inventoryJSON), &output, POMURL); err != nil {
		t.Fatalf("Convert() error = %v", err)
	}
	if !strings.Contains(output.String(), `"https://repo.test/example.jar"`) {
		t.Fatalf("Convert() did not convert POM URL: %s", output.String())
	}
}

func TestConvertRejectsInvalidJSON(t *testing.T) {
	if err := Convert(strings.NewReader("{"), &strings.Builder{}, ReferenceURL); err == nil {
		t.Fatal("Convert() accepted invalid JSON")
	}
}
