package format

import (
	"strings"
	"testing"
)

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

	tagStr, err := tags(`json:"field,omitempty"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty"`)
	}

	tagStr, err = tags(`json:"field,omitempty" monkey:"true"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty" monkey:"true"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty" monkey:"true"`)
	}

	tagStr, err = tags(`json:"field,omitempty"`, `monkey:"true"`)
	if err != nil {
		t.Error(err)
	}

	if trimBackticks(string(tagStr)) != `json:"field,omitempty" monkey:"true"` {
		t.Errorf("%q not equal to %q", trimBackticks(string(tagStr)), `json:"field,omitempty" monkey:"true"`)
	}
}

func TestFormatCommentText(t *testing.T) {
	actual := strings.TrimSpace(string(commentText("card's")))
	if actual != "// card's" {
		t.Errorf("%q not equal to %q", actual, "// card's")
	}

	actual = strings.TrimSpace(string(commentText(`What happens if I use "quotes"?`)))
	if actual != `// What happens if I use "quotes"?` {
		t.Errorf("%q not equal to %q", actual, `// What happens if I use "quotes"?`)
	}

	actual = strings.TrimSpace(string(commentText("What about\nnew lines?")))
	if actual != `// What about new lines?` {
		t.Errorf("%q not equal to %q", actual, `// What about new lines?`)
	}
}
