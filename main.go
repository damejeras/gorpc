package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/damejeras/gorpc/parser"
	"github.com/damejeras/gorpc/render"
	"github.com/pkg/errors"
)

// Version is set during build.
var Version = "dev"

func main() {
	if err := run(os.Stdout, os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(stdout io.Writer, args []string) error {
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Println(args[0] + " " + Version + ` usage:
	gorpc [flags] paths [[path2] [path3]...]`)
		fmt.Println(`
flags:`)
		flags.PrintDefaults()
	}
	var (
		template   = flags.String("template", "", "go template to render")
		outfile    = flags.String("out", "", "output file (default: stdout)")
		pkg        = flags.String("pkg", "", "explicit package name (default: inferred)")
		v          = flags.Bool("v", false, "verbose output")
		paramsStr  = flags.String("params", "", "list of parameters in the format: \"key:value,key:value\"")
		ignoreList = flags.String("ignore", "", "comma separated list of interfaces to ignore")
	)
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	if *template == "" {
		flags.PrintDefaults()
		return errors.New("missing template")
	}
	params, err := parseParams(*paramsStr)
	if err != nil {
		flags.PrintDefaults()
		return errors.Wrap(err, "params")
	}
	p := parser.New(flags.Args()...)
	ignoreItems := strings.Split(*ignoreList, ",")
	if ignoreItems[0] != "" {
		p.ExcludeInterfaces = ignoreItems
	}
	p.Verbose = *v
	if p.Verbose {
		fmt.Println("gorpc - github.com/damejeras/gorpc", Version)
	}
	def, err := p.Parse()
	if err != nil {
		return err
	}
	if *pkg != "" {
		def.PackageName = *pkg
	}
	b, err := ioutil.ReadFile(*template)
	if err != nil {
		return err
	}

	var w io.Writer = stdout
	if *outfile != "" {
		f, err := os.Create(*outfile)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}

	err = render.Render(string(b), w, def, params)
	if err != nil {
		return err
	}

	return nil
}

// parseParams returns a map of data parsed from the params string.
func parseParams(s string) (map[string]interface{}, error) {
	params := make(map[string]interface{})
	if s == "" {
		// empty map for an empty string
		return params, nil
	}
	pairs := strings.Split(s, ",")
	for i := range pairs {
		pair := strings.TrimSpace(pairs[i])
		segs := strings.Split(pair, ":")
		if len(segs) != 2 {
			return nil, errors.New("malformed params")
		}
		params[strings.TrimSpace(segs[0])] = strings.TrimSpace(segs[1])
	}
	return params, nil
}
