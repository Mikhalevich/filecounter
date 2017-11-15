package main

import (
	"fmt"
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

	for _, v := range res {
		fmt.Println(v)
	}

	if len(res) != 2 {
		t.Fatal("res count is not valid")
	}

	if res[0].Lines+res[1].Lines != (3 + 5) {
		t.Fatal("Invalid file size")
	}
}
