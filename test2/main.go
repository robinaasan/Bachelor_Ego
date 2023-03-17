package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type response struct {
	url    string
	status string
	err    error
}

func main() {
	start := time.Now()
	var urls = []string{
		//"https://googleamsdmnmnmnamnsnd.com",
		"https://bing.com",
		"https://github.com/astaxie/beego",
	}

	client := &http.Client{}

	checkURLs(urls, client)
	fmt.Printf("%.2fs elapsed", time.Since(start).Seconds())

}

func checkURLs(urls []string, ht *http.Client) {
	var wg sync.WaitGroup
	c := make(chan response)

	for _, url := range urls {
		wg.Add(1)
		go checkURL(url, c, &wg, ht)
	}

	go func() {
		wg.Wait()
		close(c)
	}()

	for resp := range c {
		if resp.err != nil {
			fmt.Printf("Error requesting %s: %v\n", resp.url, resp.err)
			continue
		}

		fmt.Printf("%s: %s\n", resp.url, resp.status)
	}
}

func checkURL(url string, c chan response, wg *sync.WaitGroup, ht *http.Client) {
	defer (*wg).Done()
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		c <- response{url, "", err}
		return
	}

	res, err := ht.Do(req)

	//resString = resString[0:1000]
	if err != nil {
		c <- response{url, "", err}
		return
		//responseData, _ := io.ReadAll(res.Body)
		//resString := string(responseData[0:100])

	}
	defer res.Body.Close()
	c <- response{url, res.Status, nil}

}
