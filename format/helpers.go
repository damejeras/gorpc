package format

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

func LoadTemplateFile(path string) (*template.Template, error) {
	return template.New(filepath.Base(path)).Funcs(map[string]interface{}{
		"camelize_down":       CamelizeDown,
		"camelize_up":         CamelizeUp,
		"json":                toJSONHelper,
		"format_comment_line": commentLine,
		"format_comment_text": commentText,
		"format_comment_html": commentHTML,
		"format_tags":         tags,
	}).ParseFiles(path)
}

func toJSONHelper(v interface{}) (string, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func commentLine(s string) string {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "", "", 2000)

	return strings.TrimSpace(buf.String())
}

func commentText(s string) string {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "// ", "", 80)

	return buf.String()
}

func commentHTML(s string) string {
	var buf bytes.Buffer
	doc.ToHTML(&buf, s, nil)

	return buf.String()
}

// tags formats a list of struct tag strings into one.
// Will return an error if any of the tag strings are invalid.
func tags(tags ...string) (string, error) {
	alltags := &structtag.Tags{}

	for _, tag := range tags {
		theseTags, err := structtag.Parse(tag)
		if err != nil {
			return "", errors.Wrapf(err, "parse tags: `%s`", tag)
		}

		for _, t := range theseTags.Tags() {
			alltags.Set(t)
		}
	}

	tagsStr := alltags.String()
	if tagsStr == "" {
		return "", nil
	}

	return "`" + tagsStr + "`", nil
}
