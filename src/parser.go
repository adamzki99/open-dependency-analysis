package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	pythonRegex   = `\b\w+\s*\.\s*\w+\s*\(`
	pythonFileExt = ".py"
	plsqlRegex    = `\b([a-zA-Z_][a-zA-Z0-9_]*)\.([a-zA-Z_][a-zA-Z0-9_]*)\s*\(`
	plsqlFileExt  = ".pkb"
)

// Edge represents a connection between two nodes with a weight
type Edge struct {
	To     string  // The node that this node depends on
	Weight float64 // The weight of the dependency
}

// Graph structure representing a weighted dependency graph
type Graph struct {
	Nodes map[string][]Edge // Adjacency list: maps a node to its dependencies with weights
}

// AddNode adds a node to the graph
func (g *Graph) addNode(node string) {
	if _, exists := g.Nodes[node]; !exists {
		g.Nodes[node] = []Edge{}
	}
}

// AddEdge adds or increments a weighted dependency (edge) between two nodes (from -> to)
func (g *Graph) AddEdge(from, to string, weight float64) {
	g.addNode(from)
	g.addNode(to)

	// Check if the edge already exists
	for i, edge := range g.Nodes[from] {
		if edge.To == to {
			// Increment the weight of the existing edge
			g.Nodes[from][i].Weight += weight
			return
		} else {
			g.Nodes[from][i].Weight = 0
		}
	}

	// If the edge doesn't exist, add a new edge
	g.Nodes[from] = append(g.Nodes[from], Edge{To: to, Weight: weight})
}

// PrintGraph prints the graph structure with weights
func (g *Graph) PrintGraph() {
	for node, edges := range g.Nodes {
		for _, edge := range edges {
			fmt.Printf("%s gets called by %s %.2f times\n", node, edge.To, edge.Weight)
		}
	}
}

// GetFirstSubstringBeforeSeparator returns the first substring before the given separator
func getFirstSubstringBeforeSeparator(input string, separator string) string {
	// Find the index of the separator
	index := strings.Index(input, separator)
	if index == -1 {
		// If the separator is not found, return the original string
		return input
	}
	// Return the substring from the start to the separator's index
	return input[:index]
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

func populateGraph(path string, regex *regexp.Regexp, dependencyGraph *Graph) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", path, err)
	}

	packageName := getFirstSubstringBeforeSeparator(getFilename(path), ".")

	for _, match := range regex.FindAllString(string(content), -1) {
		dependencyGraph.AddEdge(getFirstSubstringBeforeSeparator(match, "."), packageName, 1.0)
	}

	return nil
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
func writeJSONFile(filename string, data []Package) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error serializing data to JSON: %w", err)
	}

	return os.WriteFile(filename, jsonData, 0644)
}

type Package struct {
	Package       string             `json:"Package"`
	NumberOfLines float64            `json:"NumberOfLines"`
	DirectoryName string             `json:"DirectoryName"`
	CalledBy      map[string]float64 `json:"CalledBy"`
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

	// Create new graph
	dependencyGraph := Graph{
		Nodes: make(map[string][]Edge),
	}

	for _, path := range paths {

		err = populateGraph(path, regex, &dependencyGraph)
		if err != nil {
			log.Fatal(err)
		}

	}

	dependencyGraph.PrintGraph()

	var packages []Package

	for node, edges := range dependencyGraph.Nodes {
		var localPackage Package

		localPackage.Package = node

		localPackage.CalledBy = make(map[string]float64)

		for _, edge := range edges {

			localPackage.CalledBy[edge.To] = edge.Weight

		}
		packages = append(packages, localPackage)
	}

	//Write data to JSON file.
	if err := writeJSONFile("data.json", packages); err != nil {
		log.Fatal(err)
	}
}
