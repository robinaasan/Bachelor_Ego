package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

const usage string = "Usage: client <cmd> <key> <value>"

func main() {
	err := uploadFile("simple.wasm", "http://localhost:8080")
	if err != nil {
		fmt.Println(err)
	}
	// resp, err := http.Get("http://localhost:8080/monkeys")
	// if err != nil {
	// 	panic("wups")
	// }
	// bs := make([]byte, 99999)
	// resp.Body.Read(bs)
	// fmt.Println(string(bs))
}

func uploadFile(filepath string, url_s string) error {

	//serverURL := flag.String("url", "localhost:8080", "Server's url")
	flag.Parse()
	args := flag.Args()
	//req := url.URL{Scheme: "http", Host: *serverURL, Path: "/secret"}

	q := url.Values{}

	if len(args) < 2 {
		panic(usage)
	} else { // bare for å sjekke
		fmt.Printf(args[0])
		fmt.Printf(args[1])
		fmt.Printf(args[2])
	}

	switch args[0] {
	case "add":
		if len(args) < 3 {
			panic(usage)
		}
		q.Add("cmd", "add")
		q.Add("val1", args[1])
		q.Add("val2", args[2])
	default: // optimalt panic(usage)
		q.Add("cmd", "noe")
		q.Add("val1", args[1])
		q.Add("val2", args[2])
	}

	fmt.Println(q)

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}

	defer file.Close()

	b := &bytes.Buffer{}

	writer := multipart.NewWriter(b)

	part, err := writer.CreateFormFile("file", filepath)

	if err != nil {
		return err
	}

	_, err = io.Copy(part, file)

	if err != nil {
		return err
	}

	err = writer.Close()

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url_s, b)
	//req.RawQuery = q.Encode()
	req.URL.RawQuery = q.Encode() // Encode and assign back to the original query.
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Byte for reading the response
	bs := make([]byte, 99999)
	resp.Body.Read(bs)
	fmt.Println(string(bs))

	defer resp.Body.Close()
	//fmt.Println(string(b.Bytes()))
	return nil
}
