package main

import (
	"path/filepath"
	"testing"
)

var (
	files = map[string]FileInfo{
		"cppFile.cpp":       FileInfo{Lines: 3, Size: 6},
		"javaFile.java":     FileInfo{Lines: 5, Size: 10},
		"goFile.go":         FileInfo{Lines: 7, Size: 14},
		"jsFile.js":         FileInfo{Lines: 9, Size: 18},
		"secondCppFile.cpp": FileInfo{Lines: 2, Size: 4},
	}
)

func TestScan(t *testing.T) {
	params := &Params{
		Root: "test",
	}

	res, err := walkFiles(params)
	if len(err) > 0 {
		t.Fatal(err)
	}

	if len(res) != len(files) {
		t.Fatal("res count is not valid")
	}

	for _, r := range res {
		fileName := filepath.Base(r.Path)
		sv, ok := files[fileName]

		if !ok {
			t.Fatalf("%s no such stored filename\n", fileName)
		}

		if sv.Lines != r.Lines {
			t.Fatalf("%s lines not mutch\n", fileName)
		}

		if sv.Size != r.Size {
			t.Fatalf("%s size not mutch\n", fileName)
		}
	}

	pfi := computeResults(params, res)
	if len(pfi.Extentions) != 4 {
		t.Fatal("Invalid compute extention count")
	}

	if pfi.Total.Count != 5 {
		t.Fatal("Invalid compute total count")
	}

	if pfi.Total.Lines != 26 {
		t.Fatal("Invalid compute total lines")
	}

	if pfi.Total.Size != 52 {
		t.Fatal("Invalid compute total size")
	}
}
