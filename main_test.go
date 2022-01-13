package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplatesAgainstGoldenFiles(t *testing.T) {
	files, err := ioutil.ReadDir("./templates")
	if err != nil {
		t.Fatal(err)
	}

	for i := range files {
		filename := files[i].Name()

		goldenFileContent, err := ioutil.ReadFile(
			filepath.Join("./testdata/golden-files", strings.Replace(filename, ".tmpl", ".golden", 1)),
		)
		if err != nil {
			t.Fatal(err)
		}

		outputFile, err := ioutil.TempFile("", filename)
		if err != nil {
			t.Fatal(err)
		}

		os.Args = []string{
			"gorpc",
			"--template", filepath.Join("templates", filename),
			"--package", "main",
			"--output", outputFile.Name(),
			"./testdata/services/pleasantries",
		}

		main()

		outputContent, err := ioutil.ReadFile(outputFile.Name())
		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(outputContent, goldenFileContent) {
			t.Errorf("ouput for %q template doesn't match golden file", filename)
		}

		outputFile.Close()
	}
}
