package main

import (
	"path/filepath"
	"testing"
)

func TestScan(t *testing.T) {
	params := &Params{
		Root: "test",
	}

	res, err := walkFiles(params)
	if len(err) > 0 {
		t.Fatal(err)
	}

	files := map[string]FileInfo{
		"cppFile.cpp":   FileInfo{Lines: 3, Size: 6},
		"javaFile.java": FileInfo{Lines: 5, Size: 10},
		"goFile.go":     FileInfo{Lines: 7, Size: 14},
		"jsFile.js":     FileInfo{Lines: 9, Size: 18},
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
}
