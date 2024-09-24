package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// Site represents a website with its URL and RSS feed link
type Site struct {
	URL string `json:"url"`
	RSS string `json:"rss,omitempty"`
}

func main() {
	// Ensure correct usage: program expects one argument (input JSON file)
	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s <input_json_file>\n", os.Args[0])
		os.Exit(1)
	}

	jsonFile := os.Args[1]

	// Read and parse the input JSON file
	data, err := ioutil.ReadFile(jsonFile)
	if err != nil {
		fmt.Printf("Error reading JSON file: %v\n", err)
		os.Exit(1)
	}

	var sites []Site
	err = json.Unmarshal(data, &sites)
	if err != nil {
		fmt.Printf("Error parsing JSON file: %v\n", err)
		os.Exit(1)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	// Process each site concurrently
	for i := range sites {
		// Skip sites that already have an RSS feed or are marked as having no feed
		if sites[i].RSS != "" && sites[i].RSS != "NO_RSS_FEED" {
			continue
		}

		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(site *Site) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			fmt.Printf("Processing URL: %s\n", site.URL)
			rssLink := findRSSFeed(site.URL)
			if rssLink != "" {
				fmt.Printf("RSS feed found for %s: %s\n", site.URL, rssLink)
				site.RSS = rssLink
			} else {
				fmt.Printf("No RSS feed found for %s\n", site.URL)
				site.RSS = "NO_RSS_FEED"
			}
		}(&sites[i])
	}

	wg.Wait() // Wait for all goroutines to finish

	// Write updated sites data back to the JSON file
	outputData, err := json.MarshalIndent(sites, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	err = ioutil.WriteFile(jsonFile, outputData, 0644)
	if err != nil {
		fmt.Printf("Error writing JSON file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Processing complete. Updated JSON file: %s\n", jsonFile)
}

// findRSSFeed attempts to discover the RSS feed URL for a given website
func findRSSFeed(url string) string {
	// Define common patterns for RSS feed URLs
	prefixes := []string{"", "feed/", "feeds/", "rss/", "blog/"}
	middles := []string{"", "all", "atom", "feed", "index", "posts", "posts/default", "rss", "en", "default", "rssfeed", "blog"}
	suffixes := []string{"", ".rss", ".atom", ".rss2"}
	extensions := []string{"", ".xml", "?feed=rss2", "?format=atom"}

	// Generate all possible combinations of RSS feed paths
	paths := make([]string, 0)
	for _, prefix := range prefixes {
		for _, middle := range middles {
			for _, suffix := range suffixes {
				for _, ext := range extensions {
					path := fmt.Sprintf("%s%s%s%s", prefix, middle, suffix, ext)
					paths = append(paths, path)
				}
			}
		}
	}

	rssChan := make(chan string, 1)
	var wg sync.WaitGroup
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	semaphore := make(chan struct{}, 10) // Limit concurrency to 10 goroutines

	// Try each possible RSS feed path concurrently
	for _, path := range paths {
		wg.Add(1)
		semaphore <- struct{}{} // Acquire semaphore
		go func(path string) {
			defer wg.Done()
			defer func() { <-semaphore }() // Release semaphore
			fullURL := strings.TrimRight(url, "/") + "/" + path
			fmt.Printf("Trying path: %s\n", fullURL)
			resp, err := client.Get(fullURL)
			if err != nil {
				return
			}
			defer resp.Body.Close()

			// Read the first 512 bytes to check for RSS feed indicators
			buf := make([]byte, 512)
			n, err := resp.Body.Read(buf)
			if err != nil && err != io.EOF {
				return
			}
			content := string(buf[:n])

			// Check if the content looks like an RSS feed
			if !strings.Contains(content, "xhtml") && (strings.Contains(content, "feed") || strings.Contains(content, "xml")) {
				finalURL := resp.Request.URL.String()
				select {
				case rssChan <- finalURL:
				default:
				}
			}
		}(path)
	}

	// Close the channel when all goroutines are done
	go func() {
		wg.Wait()
		close(rssChan)
	}()

	// Return the first discovered RSS feed URL, or an empty string if none found
	if rssLink, ok := <-rssChan; ok {
		return rssLink
	}
	return ""
}
