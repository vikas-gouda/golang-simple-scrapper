package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

// getHref function extracts the value of the "href" attribute from a given HTML token.
func getHref(t html.Token) (ok bool, href string) {
	for _, a := range t.Attr {
		if a.Key == "href" {
			href = a.Val
			ok = true
		}
	}
	return ok, href
}

// crawl function retrieves the URLs from the given webpage and sends the HTTPS URLs to the channel ch.
func crawl(url string, ch chan string, chFinished chan bool) {
	// Retrieve the webpage content
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error: failed to crawl", url)
		chFinished <- true
		return
	}
	defer resp.Body.Close()

	// Create a tokenizer to parse the HTML content
	z := html.NewTokenizer(resp.Body)

	// Iterate through the HTML tokens
	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document
			chFinished <- true
			return
		case tt == html.StartTagToken:
			// Check if the token is a start tag
			t := z.Token()

			// Check if the tag is an anchor tag
			isAnchor := t.Data == "a"
			if !isAnchor {
				continue
			}

			// Extract the URL from the anchor tag
			ok, url := getHref(t)
			if !ok {
				continue
			}

			// Check if the URL starts with "https"
			hasProto := strings.Index(url, "https") == 0
			if hasProto {
				// Send the HTTPS URL to the channel
				ch <- url
			}
		}
	}
}

func main() {
	// Create a map to store unique URLs
	foundUrls := make(map[string]bool)

	// Get seed URLs from command-line arguments
	seedUrls := os.Args[1:]

	// Create channels for communicating between goroutines
	chUrls := make(chan string)
	chFinished := make(chan bool)

	// Launch a goroutine for each seed URL
	for _, url := range seedUrls {
		go crawl(url, chUrls, chFinished)
	}

	// Wait for all goroutines to finish
	for c := 0; c < len(seedUrls); {
		select {
		case url := <-chUrls:
			// Receive URLs from the channel and store them in the map
			foundUrls[url] = true
		case <-chFinished:
			// Receive finished signals from goroutines
			c++
		}
	}

	// Print the unique URLs found
	fmt.Println("\nFound", len(foundUrls), "unique URLs:\n")
	for url := range foundUrls {
		fmt.Println("-", url)
	}

	// Close the URL channel
	close(chUrls)
}
