package main

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	os.Args = []string{
		"gorpc",
		"--template", "./testdata/test.tmpl",
		"--package", "stuff",
		"--output", "output.test",
		"./testdata/services/pleasantries",
	}

	main()

	outputBytes, err := ioutil.ReadFile("output.test")
	if err != nil {
		t.Fatal(err)
	}

	defer func() { _ = os.Remove("output.test") }()

	output := string(outputBytes)

	for _, should := range []string{
		"GreeterService.GetGreetings",
		"GreeterService.Greet",
		"Welcomer.Welcome",
	} {
		if !strings.Contains(output, should) {
			t.Fatalf("missing: %s", should)
		}
	}
}
