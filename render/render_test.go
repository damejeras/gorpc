package render

import (
	"bytes"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/damejeras/gorpc/parser"
)

func TestRender(t *testing.T) {
	template := `// {{ index .Params "Description" }}
package {{ .PackageName }}`

	buffer := bytes.NewBuffer([]byte{})
	err := Render(template, buffer, parser.Definition{
		PackageName: "services",
	}, map[string]interface{}{
		"Description": "Package services contains services.",
	})
	if err != nil {
		t.Fatal("failed to render template")
	}

	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal("failed to read buffer")
	}

	for _, should := range []string{
		"// Package services contains services.",
		"package services",
	} {
		if !strings.Contains(string(output), should) {
			t.Errorf("missing: %s", should)
		}
	}
}

func TestRenderCommentsWithQuotes(t *testing.T) {
	template := `{{ range $service := .Services }}
{{ format_comment_text $service.Comment }}
type {{ $service.Name }} struct
{{ end }}`

	buffer := bytes.NewBuffer([]byte{})
	err := Render(template, buffer, parser.Definition{
		PackageName: "services",
		Services: []parser.Service{
			{
				Comment: `This comment contains "quotes"`,
				Name:    "MyService",
			},
		},
	}, nil)
	if err != nil {
		t.Fatalf("failed to render template: %v", err)
	}

	output, err := ioutil.ReadAll(buffer)
	if err != nil {
		t.Fatal("failed to read buffer")
	}

	for _, should := range []string{
		`// This comment contains "quotes"`,
	} {
		if !strings.Contains(string(output), should) {
			t.Errorf("missing: %s", should)
		}
	}
}
