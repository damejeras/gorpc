package main

import (
	"fmt"
	"github.com/damejeras/gorpc/definition"
	"github.com/damejeras/gorpc/format"
	"github.com/jessevdk/go-flags"
	"os"
	"strings"
)

var options struct {
	Template   string `short:"t" long:"template" description:"path of the template" required:"true"`
	Output     string `short:"o" long:"output" description:"output file (default: stdout)"`
	Package    string `short:"p" long:"package" description:"explicit package name (default: inferred)"`
	Ignore     string `short:"i" long:"ignore"  description:"comma separated list of interfaces to ignore"`
	Parameters string `long:"parameters" description:"list of parameters in the format \"key:value,key:value\""`

	Arguments struct {
		Input []string `positional-arg-name:"service definition" required:"1"`
	} `positional-args:"true"`
}

func main() {
	if _, err := flags.Parse(&options); err != nil {
		return
	}

	serviceDefinitionParser := definition.New(options.Arguments.Input...)
	exclusions := strings.Split(options.Ignore, ",")
	if exclusions[0] != "" {
		serviceDefinitionParser.ExcludeInterfaces = exclusions
	}

	parameters, err := definition.ParseParams(options.Parameters)
	if err != nil {
		logError(err)
		os.Exit(1)
	}

	definition, err := serviceDefinitionParser.ParseWithParams(parameters)
	if err != nil {
		logError(err)
		os.Exit(1)
	}

	if options.Package != "" {
		definition.PackageName = options.Package
	}

	template, err := format.LoadTemplateFile(options.Template)
	if err != nil {
		logError(err)
		os.Exit(1)
	}

	output := os.Stdout
	if options.Output != "" {
		outputFile, err := os.Create(options.Output)
		if err != nil {
			logError(err)
			os.Exit(1)
		}

		defer func() { _ = outputFile.Close() }()
		output = outputFile
	}

	if err := template.Execute(output, definition); err != nil {
		logError(err)
	}
}

func logError(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
}
