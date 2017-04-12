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
	"sync"
	"time"
)

const (
	cConfigFile = "config.json"
)

var (
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

type TotalFileInfo struct {
	Count int
	Size  int64
	Lines int
}

func parseArguments() (*Params, error) {
	rootDir := flag.String("root", "", "root directory to scan")
	configFile := flag.String("config", cConfigFile, "json configuration file")

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
	defer file.Close()

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

func walkFiles(params *Params) *[]FileInfo {
	var waitGroup sync.WaitGroup
	results := make([]FileInfo, 0)
	fileResult := make(chan FileInfo)
	resultFinish := make(chan bool)

	go func() {
		for info := range fileResult {
			results = append(results, info)
		}

		resultFinish <- true
	}()

	filepath.Walk(params.Root, func(path string, info os.FileInfo, err error) error {
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

		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			file, err := os.Open(path)
			if err != nil {
				return
			}
			defer file.Close()

			var lineCount int = 0
			lineReader := bufio.NewReader(file)

			for {
				_, err := lineReader.ReadString('\n')
				if err != nil {
					if err == io.EOF {
						break
					}
					return
				}
				lineCount += 1
			}

			fileResult <- FileInfo{Path: path, Size: info.Size(), Lines: lineCount, Extention: extention}
		}()

		return nil
	})

	waitGroup.Wait()
	close(fileResult)
	<-resultFinish

	return &results
}

func printResults(params *Params, results *[]FileInfo) {
	totalFiles := make(map[string]TotalFileInfo)
	totalFilesCount := 0
	totalLinesCount := 0

	for _, info := range *results {
		if params.PrintLines >= 0 && params.PrintLines < info.Lines {
			fmt.Println(info)
		}
		totalInfo := totalFiles[info.Extention]
		totalInfo.Count++
		totalInfo.Size += info.Size
		totalInfo.Lines += info.Lines
		totalFiles[info.Extention] = totalInfo

		totalFilesCount += 1
		totalLinesCount += info.Lines
	}

	fmt.Println("File count by suffix:")
	for ext, info := range totalFiles {
		fmt.Printf("%s => count = %d; size = %d; lines = %d\n", ext, info.Count, info.Size, info.Lines)
	}
	fmt.Printf("Total files = %d\n", totalFilesCount)
	fmt.Printf("Total lines = %d\n", totalLinesCount)
}

func main() {
	startTime := time.Now()

	params, err := parseArguments()
	if err != nil {
		fmt.Println(err)
		return
	}

	results := walkFiles(params)
	printResults(params, results)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
