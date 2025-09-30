package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type ServiceEndpoint struct {
	File        string
	Function    string
	ID          string
	Summary     string
	Description string
	MarkdownRef string
	Tags        string
	Router      string
}

type DuplicateReport struct {
	ExactSummaryMatches    map[string][]ServiceEndpoint
	ExactDescriptionMatches map[string][]ServiceEndpoint
	MarkdownFileUsage      map[string][]ServiceEndpoint
	SimilarDescriptions    []SimilarityMatch
}

type SimilarityMatch struct {
	Service1    ServiceEndpoint
	Service2    ServiceEndpoint
	Distance    int
	Similarity  float64
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run analyze_service_descriptions.go <path-to-ctl-api>")
		os.Exit(1)
	}

	ctlAPIPath := os.Args[1]

	// Parse all service endpoints
	endpoints, err := parseServiceEndpoints(ctlAPIPath)
	if err != nil {
		fmt.Printf("Error parsing endpoints: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d service endpoints with swagger annotations\n\n", len(endpoints))

	// Generate duplicate report
	report := generateDuplicateReport(endpoints)

	// Print report
	printDuplicateReport(report)
}

func parseServiceEndpoints(rootPath string) ([]ServiceEndpoint, error) {
	var endpoints []ServiceEndpoint

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !strings.HasSuffix(path, ".go") || strings.Contains(path, "vendor/") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		endpointsInFile := parseSwaggerAnnotations(path, file)
		endpoints = append(endpoints, endpointsInFile...)

		return nil
	})

	return endpoints, err
}

func parseSwaggerAnnotations(filePath string, file io.Reader) []ServiceEndpoint {
	var endpoints []ServiceEndpoint
	var currentEndpoint *ServiceEndpoint

	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Regex patterns for swagger annotations
	idRegex := regexp.MustCompile(`@ID\s+(\S+)`)
	summaryRegex := regexp.MustCompile(`@Summary\s+(.+)`)
	descriptionRegex := regexp.MustCompile(`@Description\s+(.+)`)
	descriptionMarkdownRegex := regexp.MustCompile(`@Description\.markdown\s+(\S+)`)
	tagsRegex := regexp.MustCompile(`@Tags\s+(.+)`)
	routerRegex := regexp.MustCompile(`@Router\s+(.+)`)
	funcRegex := regexp.MustCompile(`^func\s+\([^)]*\)\s*(\w+)\s*\(`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lineNum++

		// Check for function definition to end current endpoint
		if matches := funcRegex.FindStringSubmatch(line); matches != nil {
			if currentEndpoint != nil {
				currentEndpoint.Function = matches[1]
				endpoints = append(endpoints, *currentEndpoint)
				currentEndpoint = nil
			}
		}

		// Skip non-comment lines
		if !strings.HasPrefix(line, "//") {
			continue
		}

		// Remove comment prefix
		line = strings.TrimPrefix(line, "//")
		line = strings.TrimSpace(line)

		// Parse swagger annotations
		if matches := idRegex.FindStringSubmatch(line); matches != nil {
			// Start new endpoint
			if currentEndpoint != nil {
				endpoints = append(endpoints, *currentEndpoint)
			}
			currentEndpoint = &ServiceEndpoint{
				File: filePath,
				ID:   matches[1],
			}
		} else if currentEndpoint != nil {
			if matches := summaryRegex.FindStringSubmatch(line); matches != nil {
				currentEndpoint.Summary = strings.TrimSpace(matches[1])
			} else if matches := descriptionMarkdownRegex.FindStringSubmatch(line); matches != nil {
				currentEndpoint.MarkdownRef = matches[1]
				currentEndpoint.Description = fmt.Sprintf("markdown:%s", matches[1])
			} else if matches := descriptionRegex.FindStringSubmatch(line); matches != nil {
				currentEndpoint.Description = strings.TrimSpace(matches[1])
			} else if matches := tagsRegex.FindStringSubmatch(line); matches != nil {
				currentEndpoint.Tags = strings.TrimSpace(matches[1])
			} else if matches := routerRegex.FindStringSubmatch(line); matches != nil {
				currentEndpoint.Router = strings.TrimSpace(matches[1])
			}
		}
	}

	// Don't forget the last endpoint if file doesn't end with function
	if currentEndpoint != nil {
		endpoints = append(endpoints, *currentEndpoint)
	}

	return endpoints
}

func generateDuplicateReport(endpoints []ServiceEndpoint) DuplicateReport {
	report := DuplicateReport{
		ExactSummaryMatches:     make(map[string][]ServiceEndpoint),
		ExactDescriptionMatches: make(map[string][]ServiceEndpoint),
		MarkdownFileUsage:       make(map[string][]ServiceEndpoint),
		SimilarDescriptions:     []SimilarityMatch{},
	}

	// Group by exact summary matches
	for _, endpoint := range endpoints {
		if endpoint.Summary != "" {
			report.ExactSummaryMatches[endpoint.Summary] = append(
				report.ExactSummaryMatches[endpoint.Summary], endpoint)
		}
	}

	// Group by exact description matches
	for _, endpoint := range endpoints {
		if endpoint.Description != "" {
			report.ExactDescriptionMatches[endpoint.Description] = append(
				report.ExactDescriptionMatches[endpoint.Description], endpoint)
		}
	}

	// Group by markdown file usage
	for _, endpoint := range endpoints {
		if endpoint.MarkdownRef != "" {
			report.MarkdownFileUsage[endpoint.MarkdownRef] = append(
				report.MarkdownFileUsage[endpoint.MarkdownRef], endpoint)
		}
	}

	// Find similar descriptions using Levenshtein distance
	for i, ep1 := range endpoints {
		for j, ep2 := range endpoints {
			if i >= j || ep1.Summary == "" || ep2.Summary == "" {
				continue
			}

			distance := levenshteinDistance(ep1.Summary, ep2.Summary)
			maxLen := max(len(ep1.Summary), len(ep2.Summary))

			// Calculate similarity percentage
			similarity := float64(maxLen-distance) / float64(maxLen) * 100

			// Report if similarity is high but not exact match (85% threshold)
			if similarity >= 85.0 && similarity < 100.0 {
				report.SimilarDescriptions = append(report.SimilarDescriptions, SimilarityMatch{
					Service1:   ep1,
					Service2:   ep2,
					Distance:   distance,
					Similarity: similarity,
				})
			}
		}
	}

	// Sort similar descriptions by similarity (descending)
	sort.Slice(report.SimilarDescriptions, func(i, j int) bool {
		return report.SimilarDescriptions[i].Similarity > report.SimilarDescriptions[j].Similarity
	})

	return report
}

func printDuplicateReport(report DuplicateReport) {
	fmt.Println("=== DUPLICATE SERVICE DESCRIPTIONS REPORT ===\n")

	// Print exact summary matches
	fmt.Println("📋 EXACT SUMMARY MATCHES:")
	fmt.Println("These services have identical @Summary annotations")
	printDuplicateMatches(report.ExactSummaryMatches)

	// Print exact description matches
	fmt.Println("\n📄 EXACT DESCRIPTION MATCHES:")
	fmt.Println("These services have identical @Description content")
	printDuplicateMatches(report.ExactDescriptionMatches)

	// Print markdown file usage
	fmt.Println("\n📁 SHARED MARKDOWN FILES:")
	fmt.Println("These services reference the same @Description.markdown file")
	printDuplicateMatches(report.MarkdownFileUsage)

	// Print similar descriptions
	fmt.Println("\n🔍 SIMILAR DESCRIPTIONS (Levenshtein Distance Analysis):")
	fmt.Printf("Services with summaries that are 85%% or more similar:\n\n")

	if len(report.SimilarDescriptions) == 0 {
		fmt.Println("No similar descriptions found.")
	} else {
		for _, match := range report.SimilarDescriptions {
			fmt.Printf("Similarity: %.1f%% (distance: %d)\n", match.Similarity, match.Distance)
			fmt.Printf("  1. %s (%s)\n", match.Service1.Summary, getRelativePath(match.Service1.File))
			fmt.Printf("  2. %s (%s)\n", match.Service2.Summary, getRelativePath(match.Service2.File))
			fmt.Printf("     Function: %s vs %s\n", match.Service1.Function, match.Service2.Function)
			fmt.Printf("     Routes: %s vs %s\n\n", match.Service1.Router, match.Service2.Router)
		}
	}

	// Print summary statistics
	fmt.Println("\n📊 SUMMARY STATISTICS:")

	exactSummaryDupes := 0
	for _, endpoints := range report.ExactSummaryMatches {
		if len(endpoints) > 1 {
			exactSummaryDupes += len(endpoints)
		}
	}

	exactDescDupes := 0
	for _, endpoints := range report.ExactDescriptionMatches {
		if len(endpoints) > 1 {
			exactDescDupes += len(endpoints)
		}
	}

	sharedMarkdownFiles := 0
	for _, endpoints := range report.MarkdownFileUsage {
		if len(endpoints) > 1 {
			sharedMarkdownFiles += len(endpoints)
		}
	}

	fmt.Printf("- Services with duplicate summaries: %d\n", exactSummaryDupes)
	fmt.Printf("- Services with duplicate descriptions: %d\n", exactDescDupes)
	fmt.Printf("- Services sharing markdown files: %d\n", sharedMarkdownFiles)
	fmt.Printf("- Similar description pairs found: %d\n", len(report.SimilarDescriptions))
}

func printDuplicateMatches(matches map[string][]ServiceEndpoint) {
	duplicateCount := 0

	// Sort keys for consistent output
	var keys []string
	for key := range matches {
		if len(matches[key]) > 1 {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)

	if len(keys) == 0 {
		fmt.Println("No duplicates found.")
		return
	}

	for _, key := range keys {
		endpoints := matches[key]
		duplicateCount += len(endpoints)

		fmt.Printf("\n🔄 \"%s\" (%d occurrences):\n", key, len(endpoints))
		for i, ep := range endpoints {
			fmt.Printf("   %d. %s:%s", i+1, getRelativePath(ep.File), ep.Function)
			if ep.Router != "" {
				fmt.Printf(" [%s]", ep.Router)
			}
			if ep.Tags != "" {
				fmt.Printf(" {%s}", ep.Tags)
			}
			fmt.Println()
		}
	}
}

func getRelativePath(fullPath string) string {
	// Extract relative path from ctl-api directory
	if idx := strings.Index(fullPath, "services/ctl-api/"); idx != -1 {
		return fullPath[idx:]
	}
	return filepath.Base(fullPath)
}

func levenshteinDistance(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	matrix := make([][]int, len(s1)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(s2)+1)
	}

	for i := 0; i <= len(s1); i++ {
		matrix[i][0] = i
	}
	for j := 0; j <= len(s2); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(s1); i++ {
		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			matrix[i][j] = min(
				min(matrix[i-1][j]+1, matrix[i][j-1]+1), // deletion, insertion
				matrix[i-1][j-1]+cost,                    // substitution
			)
		}
	}

	return matrix[len(s1)][len(s2)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}