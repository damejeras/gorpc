package format

import (
	"path/filepath"
	"text/template"
)

type loader struct {
	functions map[string]interface{}
}

func LoadTemplateFile(path string, options ...Option) (*template.Template, error) {
	l := loader{functions: map[string]interface{}{
		"camelize_down":       CamelizeDown,
		"camelize_up":         CamelizeUp,
		"json":                toJSONHelper,
		"format_comment_line": commentLine,
		"format_comment_text": commentText,
		"format_comment_html": commentHTML,
		"format_tags":         tags,
	}}

	for i := range options {
		options[i](&l)
	}

	return template.New(filepath.Base(path)).Funcs(l.functions).ParseFiles(path)
}
