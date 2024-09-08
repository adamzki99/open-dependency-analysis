package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const (
	pythonRegex       = `\b\w+\s*\.\s*\w+\s*\(`
	pythonFileExt     = ".py"
	plsqlRegex        = `\b([a-zA-Z_][a-zA-Z0-9_]*)\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`
	plsqlFileExt      = ".pkb"
)

// File represents the structure of a file with relevant metadata.
type File struct {
	PackageName   string         `json:"PackageName"`
	NumberOfLines int            `json:"NumberOfLines"`
	DirectoryName string         `json:"DirectoryName"`
	References    map[string]int `json:"References"`
}

// getFilename extracts the filename from a given path.
func getFilename(path string) string {
	return filepath.Base(path)
}

// getDirectory extracts the directory name from a given file path.
func getDirectory(filePath string) string {
	directories := strings.Split(filepath.Dir(filePath), "/")
	return directories[len(directories)-1]
}

// countLines reads and counts the number of lines in a file.
func countLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	return lineCount, nil
}

// summarizeOccurrences summarizes the occurrences of strings and returns a map.
func summarizeOccurrences(references *[]string) map[string]int {
	occurrences := make(map[string]int)
	for _, ref := range *references {
		occurrences[ref]++
	}
	return occurrences
}

// processFile reads, extracts references, counts lines, and compiles metadata for each file.
func processFile(path string, regex *regexp.Regexp) (*File, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading file %s: %w", path, err)
	}

	fileName := getFilename(path)
	packageName := strings.Split(fileName, ".")[0]

	var references []string
	for _, match := range regex.FindAllString(string(content), -1) {
		references = append(references, strings.Split(match, ".")[0])
	}

	numLines, err := countLines(path)
	if err != nil {
		return nil, fmt.Errorf("error counting lines in file %s: %w", path, err)
	}

	return &File{
		PackageName:   packageName,
		DirectoryName: getDirectory(path),
		NumberOfLines: numLines,
		References:    summarizeOccurrences(&references),
	}, nil
}

// fileReader processes a batch of files in a worker goroutine.
func fileReader(id int, regex *regexp.Regexp, paths []string, results chan<- *File, wg *sync.WaitGroup) {
	defer wg.Done()
	for _, path := range paths {
		fileData, err := processFile(path, regex)
		if err != nil {
			fmt.Printf("Worker %d encountered error: %s\n", id, err)
			continue
		}
		results <- fileData
	}
}

// collectPaths collects file paths from a directory that match the specified extension.
func collectPaths(directory, extension string) ([]string, error) {
	var paths []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == extension {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error walking through directory: %w", err)
	}
	return paths, nil
}

// writeJSONFile serializes data to JSON and writes it to a file.
func writeJSONFile(filename string, data []File) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing data to JSON: %w", err)
	}

	return os.WriteFile(filename, jsonData, 0644)
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <directory> <python | plsql>")
	}

	directory := os.Args[1]
	language := os.Args[2]

	var pattern, fileExtension string
	switch language {
	case "python":
		pattern = pythonRegex
		fileExtension = pythonFileExt
	case "plsql":
		pattern = plsqlRegex
		fileExtension = plsqlFileExt
	default:
		log.Fatal("Unsupported language selected")
	}

	// Compile the regular expression.
	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(fmt.Errorf("error compiling regex: %w", err))
	}

	// Collect file paths based on the selected language.
	paths, err := collectPaths(directory, fileExtension)
	if err != nil {
		log.Fatal(err)
	}

	runtime.GOMAXPROCS(runtime.NumCPU())
	numWorkers := runtime.NumCPU()

	// Set up goroutines to process files concurrently.
	var wg sync.WaitGroup
	results := make(chan *File, len(paths))
	batchSize := len(paths) / numWorkers

	// Start worker goroutines.
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		start := (i - 1) * batchSize
		end := start + batchSize
		if i == numWorkers {
			end = len(paths)
		}
		go fileReader(i, regex, paths[start:end], results, &wg)
	}

	// Collect results from workers.
	go func() {
		wg.Wait()
		close(results)
	}()

	var files []File
	for fileData := range results {
		files = append(files, *fileData)
	}

	// Write data to JSON file.
	if err := writeJSONFile("data.json", files); err != nil {
		log.Fatal(err)
	}
}
