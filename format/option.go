package format

type Option func(l *loader)

func WithTemplateFunc(name string, function interface{}) Option {
	return func(l *loader) {
		l.functions[name] = function
	}
}
