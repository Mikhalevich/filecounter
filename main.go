package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

const (
	configFile = "config.json"
)

var (
	results            = []FileInfo{}
	skipDirectories    = make(map[string]bool)
	extensionToProcess = make(map[string]bool)
)

type Params struct {
	Root            string   `json:"root"`
	SkipDirectories []string `json:"skip,omitempty"`
	Extentions      []string `json:"ext,omitempty"`
}

type FileInfo struct {
	Path  string
	Size  int64
	Lines int
}

func (self FileInfo) String() string {
	return fmt.Sprintf("Path = %s; Size = %d; LineCount = %d", self.Path, self.Size, self.Lines)
}

func parseConfig() (*Params, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var params Params
	json.Unmarshal(bytes, &params)

	if params.Root == "" {
		return nil, errors.New("Please specify root directory")
	}

	for _, dirName := range params.SkipDirectories {
		skipDirectories[dirName] = true
	}

	for _, ext := range params.Extentions {
		extensionToProcess[ext] = true
	}

	return &params, nil
}

func processFile(path string, info os.FileInfo, err error) error {
	if _, ok := skipDirectories[info.Name()]; ok {
		return filepath.SkipDir
	}

	if info.IsDir() {
		return nil
	}

	if _, ok := extensionToProcess[filepath.Ext(path)]; !ok {
		return nil
	}

	file, err := os.Open(path)
	if err != nil {
		return err
	}

	var lineCount int = 0
	lineReader := bufio.NewReader(file)

	for {
		_, err := lineReader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		lineCount += 1
	}

	results = append(results, FileInfo{Path: path, Size: info.Size(), Lines: lineCount})

	return nil
}

func main() {
	startTime := time.Now()

	params, err := parseConfig()
	if err != nil {
		fmt.Println(err)
		return
	}

	filepath.Walk(params.Root, processFile)

	totalCount := 0
	for _, info := range results {
		fmt.Println(info)
		totalCount += info.Lines
	}
	fmt.Printf("Total lines = %d\n", totalCount)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
