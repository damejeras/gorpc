package format

import (
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/pkg/errors"
)

type File struct {
	Reader   io.Reader
	Filename string
}

func beginFile(name ...string) string {
	return ">>>BEGIN " + strings.Join(name, "") + "\n"
}

func endFile(name ...string) string {
	return "<<<END " + strings.Join(name, "") + "\n"
}

func SliceToFiles(content []byte) ([]File, error) {
	regex := regexp.MustCompile(`(?s)>>>BEGIN ([a-zA-Z0-9.]*?)\n(.*?)<<<END ([a-zA-Z0-9.]*?)\n`)
	matches := regex.FindAllSubmatch(content, -1)

	result := make([]File, 0)

	for i := range matches {
		if len(matches[i]) != 4 {
			return nil, errors.New("unexpected length of regex finds")
		}

		if bytes.Equal(matches[i][1], matches[i][3]) {
			result = append(result, File{
				Reader:   bytes.NewReader(matches[i][2]),
				Filename: string(matches[i][1]),
			})
		}
	}

	return result, nil
}
