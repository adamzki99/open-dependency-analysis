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

const pythonRegex = `\b\w+\s*\.\s*\w+\s*\(`
const pythonFileExtension = ".py"

type file struct {
	PackageName   string         `json:"PackageName"`
	NumberOfLines int            `json:"NumberOfLines"`
	DirectoryName string         `json:"DirectoryName"`
	References    map[string]int `json:"References"`
}

// getFilename extracts the filename from a given path.
func getFilename(path string) string {
	return filepath.Base(path)
}

// getDirectory takes a file path and returns the directory the file is in.
func getDirectory(filePath string) string {

	directories := strings.Split(filepath.Dir(filePath), "/")

	return directories[len(directories)-1]
}

// countLines takes a path to a file and returns the number of lines in the file.
func countLines(path string) (int, error) {
	// Open the file
	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(file)

	// Count the lines
	lineCount := 0
	for scanner.Scan() {
		lineCount++
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("error reading file: %w", err)
	}

	return lineCount, nil
}

func fileReader(id int, regex *regexp.Regexp, paths *[]string, files *[]file, mu *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()

	var localFiles []file

	for _, path := range *paths {

		var localFile file

		var references []string

		fileName := getFilename(path)
		packageName := strings.Split(fileName, ".")[0]

		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("Worker %d into error %s\n", id, err)
			return
		}

		for _, foundMatch := range regex.FindAllString(string(content), -1) {
			references = append(references, strings.Split(foundMatch, ".")[0])
		}

		localFile.References = summarizeOccurrences(&references)
		localFile.PackageName = packageName
		localFile.DirectoryName = getDirectory(path)

		localFile.NumberOfLines, err = countLines(path)
		if err != nil {
			fmt.Printf("Worker %d into error %s\n", id, err)
			return
		}

		localFiles = append(localFiles, localFile)
	}

	mu.Lock()
	(*files) = append((*files), localFiles...)
	mu.Unlock()
}

// summarizeOccurrences takes a slice of strings and returns a map where
// the keys are the unique strings from the slice and the values are the
// number of occurrences of each string.
func summarizeOccurrences(strings *[]string) map[string]int {
	// Create a map to store the occurrences
	occurrences := make(map[string]int)

	// Iterate over the slice and count occurrences
	for _, str := range *strings {
		occurrences[str]++
	}

	return occurrences
}

// writeJSONFile takes a filename and an array of the file struct,
// serializes it to JSON, and writes the JSON data to the specified file.
func writeJSONFile(filename string, data *[]file) error {
	// Serialize the array of structs to JSON with indentation
	jsonData, err := json.MarshalIndent(&data, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing data to JSON: %w", err)
	}

	// Create or open the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}
	defer file.Close()

	// Write JSON data to the file
	_, err = file.Write(jsonData)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: go run main.go <directory> <python>")
	}

	var paths []string

	var pattern string
	var fileExtension string

	directory := os.Args[1]

	if os.Args[2] == "python" {
		pattern = pythonRegex
		fileExtension = pythonFileExtension
	} else {
		log.Fatal("Unsupported programming language selected")
	}

	var mu sync.Mutex
	var wg sync.WaitGroup
	var files []file

	runtime.GOMAXPROCS(runtime.NumCPU())
	numWorkers := runtime.NumCPU()

	// Compile the regular expression.
	regex, err := regexp.Compile(pattern)
	if err != nil {
		log.Fatal(err)
	}

	// Walk through the directory and its subdirectories.
	err = filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// If it's a file, process it.
		if !info.IsDir() && filepath.Ext(path) == fileExtension {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	// Start worker goroutines
	for i := 1; i <= numWorkers; i++ {
		wg.Add(1)
		go fileReader(i, regex, &paths, &files, &mu, &wg)
	}

	// Wait for all workers to finish
	wg.Wait()

	err = writeJSONFile("data.json", &files)
	if err != nil {
		log.Fatal(err)
	}

}
