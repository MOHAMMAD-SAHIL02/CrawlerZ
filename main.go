package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/gorilla/mux"
)

type Result struct {
	URL     string   `json:"url"`
	Content string   `json:"content"`
	Links   []string `json:"links"`
}

// crawlDomain fetches the given URL, parses the content, and extracts all text and links using colly.
func crawlDomain(url string) (*Result, error) {
	c := colly.NewCollector()

	result := &Result{URL: url}

	c.OnHTML("body", func(e *colly.HTMLElement) {
		result.Content = e.Text
	})

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		if link != "" {
			result.Links = append(result.Links, link)
		}
	})

	err := c.Visit(url)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// handleCrawl handles the incoming HTTP requests, calls crawlDomain, and returns the result.
func handleCrawl(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if !strings.HasPrefix(url, "http") {
		url = "http://" + url
	}

	result, err := crawlDomain(url)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	responseFormat := r.URL.Query().Get("format")
	if responseFormat == "markdown" {
		fmt.Fprintf(w, "# URL: %s\n\n## Content:\n%s\n\n## Links:\n", result.URL, result.Content)
		for _, link := range result.Links {
			fmt.Fprintf(w, "- %s\n", link)
		}
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
	}
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/crawl", handleCrawl).Methods("GET")

	// Serve static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./static/")))

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
