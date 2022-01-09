package render

import (
	"bytes"
	"encoding/json"
	"go/doc"
	"html/template"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	"github.com/pkg/errors"
)

func LoadTemplateFile(path string) (*template.Template, error) {
	return template.New(filepath.Base(path)).Funcs(map[string]interface{}{
		"camelize_down":       camelizeDown,
		"camelize_up":         camelizeUp,
		"json":                toJSONHelper,
		"format_comment_line": formatCommentLine,
		"format_comment_text": formatCommentText,
		"format_comment_html": formatCommentHTML,
		"format_tags":         formatTags,
	}).ParseFiles(path)
}

func toJSONHelper(v interface{}) (template.HTML, error) {
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return "", err
	}
	return template.HTML(b), nil
}

func formatCommentLine(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "", "", 2000)
	s = strings.TrimSpace(buf.String())
	return template.HTML(s)
}

func formatCommentText(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToText(&buf, s, "// ", "", 80)
	return template.HTML(buf.String())
}

func formatCommentHTML(s string) template.HTML {
	var buf bytes.Buffer
	doc.ToHTML(&buf, s, nil)
	return template.HTML(buf.String())
}

// formatTags formats a list of struct tag strings into one.
// Will return an error if any of the tag strings are invalid.
func formatTags(tags ...string) (template.HTML, error) {
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
	tagsStr = "`" + tagsStr + "`"
	return template.HTML(tagsStr), nil
}
