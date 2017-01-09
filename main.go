package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
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
	PrintLines      int      `json:"print_lines,omitempty"`
}

func NewParams() *Params {
	return &Params{
		Root:            "",
		SkipDirectories: make([]string, 0),
		Extentions:      make([]string, 0),
		PrintLines:      -1,
	}
}

type FileInfo struct {
	Path      string
	Size      int64
	Lines     int
	Extention string
}

func (self FileInfo) String() string {
	return fmt.Sprintf("Path = %s; Size = %d; LineCount = %d", self.Path, self.Size, self.Lines)
}

func parseArguments() (*Params, error) {
	rootDir := flag.String("root", "", "root directory to scan")
	configFile := flag.String("config", "", "json configuration file")

	flag.Parse()

	params := NewParams()
	var err error
	if *configFile != "" {
		params, err = parseConfig(*configFile)
		if err != nil {
			return nil, err
		}
	}

	if *rootDir != "" {
		params.Root = *rootDir
	}

	if params.Root == "" {
		return nil, errors.New("Please specify root directory")
	}

	return params, nil
}

func parseConfig(configFile string) (*Params, error) {
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

	extention := filepath.Ext(path)
	if len(extensionToProcess) > 0 {
		if _, ok := extensionToProcess[extention]; !ok {
			return nil
		}
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

	results = append(results, FileInfo{Path: path, Size: info.Size(), Lines: lineCount, Extention: extention})

	return nil
}

func main() {
	startTime := time.Now()

	params, err := parseArguments()
	if err != nil {
		fmt.Println(err)
		return
	}

	filepath.Walk(params.Root, processFile)

	files := make(map[string]int)
	totalFiles := 0
	totalLines := 0
	for _, info := range results {
		if params.PrintLines >= 0 && params.PrintLines < info.Lines {
			fmt.Println(info)
		}
		filecount := files[info.Extention]
		filecount++
		files[info.Extention] = filecount
		totalFiles += 1
		totalLines += info.Lines
	}

	for ext, count := range files {
		fmt.Printf("%s files = %d\n", ext, count)
	}
	fmt.Printf("Total files = %d\n", totalFiles)
	fmt.Printf("Total lines = %d\n", totalLines)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
