package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/Mikhalevich/argparser"
	"github.com/Mikhalevich/jober"
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
	PrintBySize     bool     `json:"print_by_size,omitempty"`
	PrintValue      int      `json:"print_value,omitempty"`
}

func NewParams() *Params {
	return &Params{
		Root:            "",
		SkipDirectories: make([]string, 0),
		Extentions:      make([]string, 0),
		PrintBySize:     false,
		PrintValue:      -1,
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

type GroupFileInfo struct {
	Count int
	Size  ByteSize
	Lines int
}

func (self GroupFileInfo) String() string {
	return fmt.Sprintf("Count = %d; Size %s; LineCount = %d", self.Count, self.Size, self.Lines)
}

func parseArguments() (*Params, error) {
	rootDir := argparser.String("root", "", "root directory to scan")

	params := NewParams()
	p, err, _ := argparser.Parse(params)

	if err != nil {
		return nil, err
	}
	params = p.(*Params)

	for _, dirName := range params.SkipDirectories {
		skipDirectories[dirName] = true
	}

	for _, ext := range params.Extentions {
		extensionToProcess[ext] = true
	}

	if *rootDir != "" {
		params.Root = *rootDir
	}

	if params.Root == "" {
		return nil, errors.New("Please specify root directory")
	}

	return params, nil
}

func walkFiles(params *Params) ([]FileInfo, []error) {
	fileJob := jober.NewWorkerPool(jober.NewAll(), 1000)

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

	res, errs := fileJob.Get()
	results := make([]FileInfo, len(res))
	for index, value := range res {
		results[index] = value.(FileInfo)
	}

	return results, errs
}

func computeResults(pc PrintChecker, results []FileInfo) *PrintFileInfo {
	pfi := NewPrintFileInfo()

	sort.Slice(results, func(i, j int) bool { return results[i].Lines < results[j].Lines })

	for _, info := range results {
		if pc.Match(info) {
			pfi.Print = append(pfi.Print, info)
		}
		extInfo := pfi.Extentions[info.Extention]
		extInfo.Count++
		extInfo.Size += info.Size
		extInfo.Lines += info.Lines
		pfi.Extentions[info.Extention] = extInfo

		pfi.Total.Count += 1
		pfi.Total.Lines += info.Lines
		pfi.Total.Size += info.Size
	}

	return pfi
}

func main() {
	startTime := time.Now()

	params, err := parseArguments()
	if err != nil {
		fmt.Println(err)
		return
	}

	results, errors := walkFiles(params)
	cfr := computeResults(checker(params), results)
	printResults(cfr, errors)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
