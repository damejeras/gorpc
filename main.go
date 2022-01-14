package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/damejeras/gorpc/definition"
	"github.com/damejeras/gorpc/format"
	"github.com/jessevdk/go-flags"
)

var options struct {
	Template   string `short:"t" long:"template" description:"path of the template" required:"true"`
	Output     string `short:"o" long:"output" description:"output file or directory in case of multiple files (default: stdout)"`
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

	definitionParser := definition.NewParser(options.Arguments.Input...)
	exclusions := strings.Split(options.Ignore, ",")
	if exclusions[0] != "" {
		definitionParser.Exclusions = exclusions
	}

	parameters, err := definition.ParseParams(options.Parameters)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	rootDefinition, err := definitionParser.ParseWithParams(parameters)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	if options.Package != "" {
		rootDefinition.PackageName = options.Package
	}

	template, err := format.LoadTemplateFile(
		options.Template,
		format.WithTemplateFunc("is_input", rootDefinition.ObjectIsInput),
		format.WithTemplateFunc("is_output", rootDefinition.ObjectIsOutput),
	)
	if err != nil {
		printErr(err)
		os.Exit(1)
	}

	output := bytes.NewBuffer([]byte{})

	if err := template.Execute(output, rootDefinition); err != nil {
		printErr(err)
		os.Exit(1)
	}

	if err := printOutput(output); err != nil {
		printErr(err)
	}
}

func printOutput(output *bytes.Buffer) error {
	if options.Output != "" {
		stat, err := os.Stat(options.Output)
		if err != nil && !os.IsNotExist(err) {
			return err
		} else if os.IsNotExist(err) || !stat.IsDir() {
			outputFile, fileErr := os.Create(options.Output)
			if fileErr != nil {
				return fileErr
			}

			defer outputFile.Close()

			if _, copyErr := io.Copy(outputFile, output); copyErr != nil {
				return copyErr
			}

			return nil
		}

		content, readErr := ioutil.ReadAll(output)
		if readErr != nil {
			return readErr
		}

		files, sliceErr := format.SliceToFiles(content)
		if sliceErr != nil {
			return sliceErr
		}

		for i := range files {
			outputFile, fileErr := os.Create(filepath.Join(options.Output, files[i].Filename))
			if fileErr != nil {
				return fileErr
			}

			if _, copyErr := io.Copy(outputFile, files[i].Reader); copyErr != nil {
				outputFile.Close()

				return copyErr
			}

			if closeErr := outputFile.Close(); closeErr != nil {
				return closeErr
			}
		}
	}

	if _, err := io.Copy(os.Stdout, output); err != nil {
		return err
	}

	return nil
}

func printErr(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
}
