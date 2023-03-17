package main

import (
	"fmt"
	"net/http"
	"sync"
)

type response struct {
	url    string
	status string
	err    error
}

func main() {
	urls := []string{
		"https://www.google.com",
		"https://www.facebooklkasdklansd.com",
		"https://www.twitter.com",
	}

	client := &http.Client{}

	responses := make(chan response)
	var wg sync.WaitGroup
	wg.Add(len(urls))

	for _, url := range urls {
		go func(url string) {
			defer wg.Done()

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				responses <- response{url, "", err}
				return
			}

			resp, err := client.Do(req)
			if err != nil {
				responses <- response{url, "", err}
				return
			}

			defer resp.Body.Close()

			responses <- response{url, resp.Status, nil}
		}(url)
	}

	go func() {
		wg.Wait()
		close(responses)
	}()

	for resp := range responses {
		if resp.err != nil {
			fmt.Printf("Error requesting %s: %v\n", resp.url, resp.err)
			continue
		}

		fmt.Printf("%s: %s\n", resp.url, resp.status)
	}
}
