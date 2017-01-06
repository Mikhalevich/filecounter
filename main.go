package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

type FileInfo struct {
	Path  string
	Size  int64
	Lines int
}

func (self FileInfo) String() string {
	return fmt.Sprintf("Path = %s; Size = %d; LineCount = %d", self.Path, self.Size, self.Lines)
}

var (
	results         = make([]FileInfo, 0)
	skipDirectories = map[string]bool{"skip_dir": true}
)

func processFile(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		if _, ok := skipDirectories[info.Name()]; ok {
			return filepath.SkipDir
		}
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

	filepath.Walk("test", processFile)

	totalCount := 0
	for _, info := range results {
		fmt.Println(info)
		totalCount += info.Lines
	}
	fmt.Printf("Total lines = %d\n", totalCount)

	fmt.Printf("Execution time = %v\n", time.Now().Sub(startTime))
}
