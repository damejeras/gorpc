package format

import (
	"bytes"
	"testing"
)

func TestSliceToFiles(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		output []File
	}{
		{
			name: "simple case",
			input: `
// generated
>>>BEGIN Test.file
<div>some html</div>
<<<END Test.file
`,
			output: []File{
				{Filename: "Test.file", Reader: bytes.NewReader([]byte(`<div>some html</div>`))},
			},
		},
		{
			name: "multiple files",
			input: `
// generated
>>>BEGIN Test.file
<div>some html</div>
<<<END Test.file
some unimportant text
>>>BEGIN Test2.file
<a href="#">some link</a>
<<<END Test2.file
`,
			output: []File{
				{Filename: "Test.file", Reader: bytes.NewReader([]byte(`<div>some html</div>`))},
				{Filename: "Test2.file", Reader: bytes.NewReader([]byte(`<a href="#">some link</a>`))},
			},
		},
		{
			name: "exclude mismatching names",
			input: `
// generated
>>>BEGIN Test.file
<div>some html</div>
<<<END Test.file
some unimportant text
>>>BEGIN Test2.file
<a href="#">some link</a>
<<<END Test2.file
some uninmportant text
>>>BEGIN Test3.file
<a href="#">some link</a>
<<<END Test4.file
`,
			output: []File{
				{Filename: "Test.file", Reader: bytes.NewReader([]byte(`<div>some html</div>`))},
				{Filename: "Test2.file", Reader: bytes.NewReader([]byte(`<a href="#">some link</a>`))},
			},
		},
		{
			name: "exclude too long names",
			input: `
// generated
>>>BEGIN Test.file some unimportant text
<div>some html</div>
<<<END Test.file some unimportant text
>>>BEGIN Test2.file
<a href="#">some link</a>
<<<END Test2.file
some uninmportant text
>>>BEGIN Test3.file
<a href="#">some link</a>
<<<END Test4.file
`,
			output: []File{
				{Filename: "Test2.file", Reader: bytes.NewReader([]byte(`<a href="#">some link</a>`))},
			},
		},
	}

	for i := range testCases {
		output, err := SliceToFiles([]byte(testCases[i].input))
		if err != nil {
			t.Error(err)
		}

		if len(output) != len(testCases[i].output) {
			t.Errorf("%s: output length expected to be %d, got %d", testCases[i].name, len(testCases[i].output), len(output))
		}
	}
}
