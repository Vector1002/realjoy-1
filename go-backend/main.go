package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type ScrapeRequest struct {
	ArrivalDate   string   `json:"arrivalDate"`
	DepartureDate string   `json:"departureDate"`
	URLs          []string `json:"urls"`
}

type ScrapeResult struct {
	URL   string `json:"url"`
	Price string `json:"price"`
}

var cachedURLs []string

// Read the list of URLs from a file
func readURLs() ([]string, error) {
	file, err := os.Open("list.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var urls []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			urls = append(urls, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

// CORS middleware to allow requests from your frontend
func withCORS(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "https://realjoy-1.vercel.app")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		handler(w, r)
	}
}

// Scrape handler to process scraping requests
func scrapeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ScrapeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var urls []string
	if len(req.URLs) > 0 {
		urls = req.URLs
	} else {
		urls = cachedURLs
	}

	// Launch the Chromium browser using Rod
	l := launcher.New().Bin("/usr/bin/chromium").Headless(true).NoSandbox(true)
	url := l.MustLaunch()
	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	var wg sync.WaitGroup
	resultChan := make(chan ScrapeResult, len(urls))
	semaphore := make(chan struct{}, 6) // Limit concurrency to 6 requests

	// Set up the request interceptor to block unnecessary resources
	browser.MustSetRequestInterceptor(func(req *rod.Request) {
		// Block unnecessary resource types such as Image, Font, Stylesheet, Script
		if req.ResourceType == "Image" || req.ResourceType == "Font" || req.ResourceType == "Stylesheet" || req.ResourceType == "Script" {
			req.Abort() // Abort the request if it's an image, font, stylesheet, or script
		} else {
			req.Continue() // Continue the request for other types
		}
	})

	// Iterate over the URLs and scrape them concurrently
	for _, baseURL := range urls {
		baseURL := baseURL
		wg.Add(1)
		semaphore <- struct{}{} // Reserve a spot in the semaphore

		go func() {
			defer wg.Done()
			defer func() { <-semaphore }() // Release the spot in the semaphore

			// Create the full URL with arrival and departure dates
			fullURL := fmt.Sprintf("%s?checkin=%s&checkout=%s", baseURL, req.ArrivalDate, req.DepartureDate)
			page := browser.MustPage(fullURL)
			defer page.MustClose()

			// Wait until the page is fully loaded
			page.MustWaitLoad()

			// Find the price on the page
			var bestPrice string
			hels, _ := page.Elements(".pdp-quote-total span")
			for _, el := range hels {
				txt := strings.TrimSpace(el.MustText())
				if strings.HasPrefix(txt, "$") {
					bestPrice = txt
				}
			}

			// If no price is found, set it to "N/A"
			if bestPrice == "" {
				bestPrice = "N/A"
			}

			// Send the result back to the main goroutine
			resultChan <- ScrapeResult{URL: fullURL, Price: bestPrice}
		}()
	}

	// Wait for all goroutines to finish
	wg.Wait()
	close(resultChan)

	// Collect the results and send them as JSON
	var results []ScrapeResult
	for r := range resultChan {
		results = append(results, r)
	}

	// Set response content type to JSON and send the results
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func main() {
	// Read URLs from list.txt
	var err error
	cachedURLs, err = readURLs()
	if err != nil {
		fmt.Println("❌ Failed to load list.txt:", err)
		os.Exit(1)
	}

	// Set up HTTP server and routes
	http.HandleFunc("/scrape", withCORS(scrapeHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port if not set in environment
	}

	// Start the server
	fmt.Printf("🚀 Server started at http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}
