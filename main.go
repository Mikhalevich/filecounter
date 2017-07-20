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
	"sort"
	"time"

	"github.com/Mikhalevich/gojob"
)

const (
	cConfigFile = "config.json"
)

type ByteSize float64

const (
	_           = iota // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
	ZB
	YB
)

func (b ByteSize) String() string {
	switch {
	case b >= YB:
		return fmt.Sprintf("%.2fYB", b/YB)
	case b >= ZB:
		return fmt.Sprintf("%.2fZB", b/ZB)
	case b >= EB:
		return fmt.Sprintf("%.2fEB", b/EB)
	case b >= PB:
		return fmt.Sprintf("%.2fPB", b/PB)
	case b >= TB:
		return fmt.Sprintf("%.2fTB", b/TB)
	case b >= GB:
		return fmt.Sprintf("%.2fGB", b/GB)
	case b >= MB:
		return fmt.Sprintf("%.2fMB", b/MB)
	case b >= KB:
		return fmt.Sprintf("%.2fKB", b/KB)
	}
	return fmt.Sprintf("%.2fB", b)
}

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
	Size      ByteSize
	Lines     int
	Extention string
}

func (self FileInfo) String() string {
	return fmt.Sprintf("Path = %s; Size = %s; LineCount = %d", self.Path, self.Size, self.Lines)
}

type TotalFileInfo struct {
	Count int
	Size  ByteSize
	Lines int
}

func parseArguments() (*Params, error) {
	rootDir := flag.String("root", "", "root directory to scan")
	configFile := flag.String("config", cConfigFile, "json configuration file")

	flag.Parse()

	var params *Params
	var err error
	if *configFile != "" {
		params, err = parseConfig(*configFile)
		if err != nil {
			return nil, err
		}
	}
	if params == nil {
		params = NewParams()
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
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	bytes, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	params := NewParams()
	err = json.Unmarshal(bytes, params)
	if err != nil {
		return nil, err
	}

	for _, dirName := range params.SkipDirectories {
		skipDirectories[dirName] = true
	}

	for _, ext := range params.Extentions {
		extensionToProcess[ext] = true
	}

	return params, nil
}

func walkFiles(params *Params) ([]FileInfo, []error) {
	fileJob := gojob.NewJob()

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

		workerFunc := func() (interface{}, error) {
			file, err := os.Open(path)
			if err != nil {
				return nil, err
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
					return nil, err
				}
				lineCount += 1
			}

			var bs ByteSize = ByteSize(info.Size())
			return FileInfo{Path: path, Size: bs, Lines: lineCount, Extention: extention}, nil
		}

		fileJob.Add(workerFunc)

		return nil
	})

	fileJob.Wait()

	results := make([]FileInfo, len(fileJob.Results))
	for index, value := range fileJob.Results {
		results[index] = value.(FileInfo)
	}

	return results, fileJob.Errors
}

func printResults(params *Params, results []FileInfo, errors []error) {
	totalFiles := make(map[string]TotalFileInfo)
	totalFilesCount := 0
	totalLinesCount := 0

	sort.Slice(results, func(i, j int) bool { return results[i].Lines < results[j].Lines })

	for _, info := range results {
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
		fmt.Printf("%s => count = %d; size = %s; lines = %d\n", ext, info.Count, info.Size, info.Lines)
	}
	fmt.Printf("Total files = %d\n", totalFilesCount)
	fmt.Printf("Total lines = %d\n", totalLinesCount)

	for _, err := range errors {
		fmt.Printf("Error: %v\n", err)
	}
}

func main() {
	startTime := time.Now()

	params, err := parseArguments()
	if err != nil {
		fmt.Println(err)
		return
	}

	results, errors := walkFiles(params)
	printResults(params, results, errors)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
