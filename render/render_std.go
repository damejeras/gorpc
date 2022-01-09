package render

import (
	"github.com/damejeras/gorpc/parser"
	"io"
	"text/template"
)

type definition struct {
	parser.Definition
	Params map[string]interface{}
}

func RenderStd(t string, out io.Writer, def parser.Definition, params map[string]interface{}) error {
	tpl, err := template.New("test").Funcs(map[string]interface{}{
		"camelize_down":       camelizeDown,
		"camelize_up":         camelizeUp,
		"json":                toJSONHelper,
		"format_comment_line": formatCommentLine,
		"format_comment_text": formatCommentText,
		"format_comment_html": formatCommentHTML,
		"format_tags":         formatTags,
	}).Parse(t)
	if err != nil {
		return err
	}

	return tpl.Execute(out, definition{
		Definition: def,
		Params:     params,
	})
}
