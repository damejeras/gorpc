package render

import (
	"strings"
	"testing"
)

func TestCamelizeDown(t *testing.T) {
	for in, expected := range map[string]string{
		"CamelsAreGreat": "camelsAreGreat",
		"ID":             "id",
		"HTML":           "html",
		"PreviewHTML":    "previewHTML",
	} {
		actual := camelizeDown(in)
		if actual != expected {
			t.Errorf("%s expected: %q but got %q", in, expected, actual)
		}
	}
}

func TestFormatTags(t *testing.T) {
	trimBackticks := func(s string) string {
		if !strings.HasPrefix(s, "`") {
			t.Errorf("%q doesnt have prefix %q", s, "`")
		}
		if !strings.HasSuffix(s, "`") {
			t.Errorf("%q doesnt have suffix %q", s, "`")
		}

		return strings.Trim(s, "`")
	}

	tagStr, err := formatTags(`json:"field,omitempty"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty"`)
	}

	tagStr, err = formatTags(`json:"field,omitempty" monkey:"true"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty" monkey:"true"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty" monkey:"true"`)
	}

	tagStr, err = formatTags(`json:"field,omitempty"`, `monkey:"true"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty" monkey:"true"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty" monkey:"true"`)
	}
}

func TestFormatCommentText(t *testing.T) {
	actual := strings.TrimSpace(string(formatCommentText("card's")))
	if actual != "// card's" {
		t.Errorf("%q not equal to %q", actual, "// card's")
	}

	actual = strings.TrimSpace(string(formatCommentText(`What happens if I use "quotes"?`)))
	if actual != `// What happens if I use "quotes"?` {
		t.Errorf("%q not equal to %q", actual, `// What happens if I use "quotes"?`)
	}

	actual = strings.TrimSpace(string(formatCommentText("What about\nnew lines?")))
	if actual != `// What about new lines?` {
		t.Errorf("%q not equal to %q", actual, `// What about new lines?`)
	}
}
