package main

import (
	"fmt"
)

type PrintFileInfo struct {
	Extentions map[string]GroupFileInfo
	Total      GroupFileInfo
	Print      []FileInfo
}

func NewPrintFileInfo() *PrintFileInfo {
	return &PrintFileInfo{
		Extentions: make(map[string]GroupFileInfo),
		Print:      make([]FileInfo, 0),
	}
}

type PrintChecker interface {
	Match(info FileInfo) bool
}

type LineChecker struct {
	Value int
}

func (lc LineChecker) Match(info FileInfo) bool {
	return lc.Value >= 0 && lc.Value <= info.Lines
}

type SizeChecker struct {
	Value ByteSize
}

func (sc SizeChecker) Match(info FileInfo) bool {
	return sc.Value >= 0 && sc.Value <= info.Size
}

func checker(p *Params) PrintChecker {
	if p.PrintBySize {
		return SizeChecker{Value: ByteSize(p.PrintValue)}
	} else {
		return LineChecker{Value: p.PrintValue}
	}
}

func printResults(pfi *PrintFileInfo, errors []error) {
	fmt.Println("Files:")
	for _, info := range pfi.Print {
		fmt.Println(info)
	}

	fmt.Println("Errors:")
	for _, err := range errors {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("File count by suffix:")
	for ext, info := range pfi.Extentions {
		fmt.Printf("%s => count = %d; size = %s; lines = %d\n", ext, info.Count, info.Size, info.Lines)
	}

	fmt.Println("Total file info:")
	fmt.Println(pfi.Total)
}
